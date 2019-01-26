package dijct

import (
	"fmt"
	"reflect"
)

type (
	container struct {
		factoryInfos    map[reflect.Type]factoryInfo
		cache           map[reflect.Type]reflect.Value
		parentContainer *container
	}
	// Container は DIコンテナーです
	Container interface {
		Register(constructor Target, options ...RegisterOptions) error
		Invoke(invoker Invoker) error
		CreateChildContainer() Container
	}
)

// NewContainer はコンテナーを生成します
func NewContainer(options ...ContainerOptions) Container {
	return newContainer(nil)
}
func newContainer(parentContainer *container) *container {
	return &container{
		factoryInfos:    make(map[reflect.Type]factoryInfo),
		cache:           make(map[reflect.Type]reflect.Value),
		parentContainer: parentContainer,
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
	if options != nil && len(options) == 1 {
		option := options[0]
		lts = option.lifetimeScope
	}
	isFunc := ins != nil
	if !isFunc {
		lts = ContainerManaged
	}
	container.factoryInfos[out] = factoryInfo{target: reflect.ValueOf(target), lifetimeScope: lts, ins: ins, isFunc: isFunc}
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

	fn := reflect.ValueOf(t)
	outs := fn.Call(args)
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

	fn := reflect.ValueOf(t)
	outs := fn.Call(args)
	out := outs[0]
	c[t] = out
	return &out, nil
}
