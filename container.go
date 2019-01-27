package dijct

import (
	"fmt"
	"reflect"
)

type (
	container struct {
		factoryInfos                map[reflect.Type]factoryInfo
		cache                       map[reflect.Type]reflect.Value
		parentContainer             *container
		containerInterfaceType      reflect.Type
		ioCContainerInterfaceType   reflect.Type
		serviceLocatorInterfaceType reflect.Type
	}
	// Container は DIコンテナーです
	Container interface {
		Register(constructor Target, options ...RegisterOptions) error
		IoCContainer
	}
	// IoCContainer です
	IoCContainer interface {
		ServiceLocator
		CreateChildContainer() Container
	}
	// ServiceLocator です
	ServiceLocator interface {
		Invoke(invoker Invoker) error
	}
)

// NewContainer はコンテナーを生成します
func NewContainer(options ...ContainerOptions) Container {
	return newContainer(nil)
}
func newContainer(parentContainer *container) *container {
	return &container{
		factoryInfos:                make(map[reflect.Type]factoryInfo),
		cache:                       make(map[reflect.Type]reflect.Value),
		parentContainer:             parentContainer,
		containerInterfaceType:      reflect.TypeOf((*Container)(nil)).Elem(),
		ioCContainerInterfaceType:   reflect.TypeOf((*IoCContainer)(nil)).Elem(),
		serviceLocatorInterfaceType: reflect.TypeOf((*ServiceLocator)(nil)).Elem(),
	}
}

// CreateChildContainer は子コンテナを生成します
func (container *container) CreateChildContainer() Container {
	return newContainer(container)
}

// Register はコンストラクタまたは定数を登録します
func (container *container) Register(target Target, options ...RegisterOptions) error {
	if options != nil && len(options) > 1 {
		return fmt.Errorf("オプションは単一である必要があります")
	}
	out, ins, err := getTargetReflectionInfos(target)
	if err != nil {
		return err
	}
	lts := InvokeManaged
	kind := out.Kind()
	isFunc := ins != nil
	if !isFunc {
		lts = ContainerManaged
	}
	count := 0
	value := reflect.ValueOf(target)
	if options != nil && len(options) == 1 {
		option := options[0]
		if isFunc {
			lts = option.LifetimeScope
		}
		if option.Interfaces != nil && len(option.Interfaces) > 0 {
			for _, p := range option.Interfaces {
				container.factoryInfos[p] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
				count++
			}
		}
	}
	if kind != reflect.Ptr {
		container.factoryInfos[out] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
		count++
	} else if count == 0 {
		return fmt.Errorf("ポインタを登録する場合は、インターフェイスを指定する必要があります")
	}
	return nil
}

// Invoke はコンテナからインスタンスを解決して呼び出します
func (container *container) Invoke(invoker Invoker) error {
	t := reflect.TypeOf(invoker)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("関数を指定してください")
	}
	ins := getIns(t)
	lenIns := len(ins)
	if lenIns == 0 {
		return fmt.Errorf("解決するオブジェクトが存在しません")
	}
	args := make([]reflect.Value, lenIns)
	cache := make(map[reflect.Type]reflect.Value)
	for i, in := range ins {
		v, err := container.resolve(in, &cache)
		if err != nil {
			return err
		}
		args[i] = *v
	}

	fn := reflect.ValueOf(invoker)
	outs := fn.Call(args)
	for _, out := range outs {
		if err, ok := out.Interface().(error); ok {
			return err
		}
	}
	return nil
}

func (container *container) resolve(t reflect.Type, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	if container.containerInterfaceType == t || container.ioCContainerInterfaceType == t || container.serviceLocatorInterfaceType == t {
		v := reflect.ValueOf(container)
		return &v, nil
	}
	factoryInfo, ok := container.factoryInfos[t]
	if !ok {
		if container.parentContainer != nil {
			return container.parentContainer.resolve(t, cache)
		}
		return nil, fmt.Errorf("指定されたタイプを解決できません。(%v)", t)
	}
	switch factoryInfo.lifetimeScope {
	case ContainerManaged:
		return container.resolveContainerManagedObject(t, factoryInfo, cache)
	}
	return container.resolveInvokeManagedObject(t, factoryInfo, cache)
}
func (container *container) resolveContainerManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	if v, ok := container.cache[t]; ok {
		return &v, nil
	}
	if !factoryInfo.isFunc {
		container.cache[t] = factoryInfo.target
		return &factoryInfo.target, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := container.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	out := outs[0]
	container.cache[t] = out
	return &out, nil
}
func (container *container) resolveInvokeManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	c := *cache
	if v, ok := c[t]; ok {
		return &v, nil
	}
	if !factoryInfo.isFunc {
		c[t] = factoryInfo.target
		return &factoryInfo.target, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := container.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	out := outs[0]
	c[t] = out
	return &out, nil
}
