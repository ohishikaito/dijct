package dijct

import (
	"reflect"
)

type (
	container struct {
		factoryInfos                map[reflect.Type]factoryInfo
		cache                       map[reflect.Type]reflect.Value
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
		Verify() error
	}
)

// NewContainer はコンテナーを生成します
func NewContainer(options ...ContainerOptions) Container {
	return newContainer(make(map[reflect.Type]factoryInfo), make(map[reflect.Type]reflect.Value))
}
func newContainer(factoryInfos map[reflect.Type]factoryInfo, cache map[reflect.Type]reflect.Value) *container {
	return &container{
		factoryInfos:                factoryInfos,
		cache:                       cache,
		containerInterfaceType:      reflect.TypeOf((*Container)(nil)).Elem(),
		ioCContainerInterfaceType:   reflect.TypeOf((*IoCContainer)(nil)).Elem(),
		serviceLocatorInterfaceType: reflect.TypeOf((*ServiceLocator)(nil)).Elem(),
	}
}

// CreateChildContainer は子コンテナを生成します
func (container *container) CreateChildContainer() Container {
	factoryInfos := make(map[reflect.Type]factoryInfo)
	for key, value := range container.factoryInfos {
		factoryInfos[key] = value
	}
	cache := make(map[reflect.Type]reflect.Value)
	for key, value := range container.cache {
		cache[key] = value
	}
	return newContainer(factoryInfos, cache)
}

// Register はコンストラクタまたは定数を登録します
func (container *container) Register(target Target, options ...RegisterOptions) error {
	if len(options) > 1 {
		return ErrNoMultipleOption
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
	if len(options) == 1 {
		option := options[0]
		if isFunc {
			lts = option.LifetimeScope
		}
		if option.Interfaces != nil && len(option.Interfaces) > 0 {
			for _, p := range option.Interfaces {
				container.factoryInfos[p] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
				_, ok := container.cache[p]
				if ok {
					delete(container.cache, p)
				}
				count++
			}
		}
	}
	if kind != reflect.Ptr {
		container.factoryInfos[out] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
		_, ok := container.cache[out]
		if ok {
			delete(container.cache, out)
		}
		count++
	} else if count == 0 {
		return ErrNeedInterfaceOnPointerRegistering
	}
	return nil
}

// Invoke はコンテナからインスタンスを解決して呼び出します
func (container *container) Invoke(invoker Invoker) error {
	t := reflect.TypeOf(invoker)
	if t.Kind() != reflect.Func {
		return ErrRequireFunction
	}
	ins := getIns(t)
	lenIns := len(ins)
	if lenIns == 0 {
		return ErrNotFoundComponent
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
		return nil, newErrInvalidResolveComponent(t)
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

func (container *container) Verify() error {
	lenIns := len(container.factoryInfos)
	if lenIns == 0 {
		return ErrNotFoundComponent
	}
	args := make([]reflect.Value, lenIns)
	cache := make(map[reflect.Type]reflect.Value)
	i := 0
	for t := range container.factoryInfos {
		v, err := container.resolve(t, &cache)
		if err != nil {
			return err
		}
		args[i] = *v
		i++
	}
	return nil
}
