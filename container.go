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
func (c *container) CreateChildContainer() Container {
	factoryInfos := make(map[reflect.Type]factoryInfo)
	for key, value := range c.factoryInfos {
		factoryInfos[key] = value
	}
	cache := make(map[reflect.Type]reflect.Value)
	for key, value := range c.cache {
		cache[key] = value
	}
	return newContainer(factoryInfos, cache)
}

// Register はコンストラクタまたは定数を登録します
func (c *container) Register(target Target, options ...RegisterOptions) error {
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
				c.factoryInfos[p] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
				_, ok := c.cache[p]
				if ok {
					delete(c.cache, p)
				}
				count++
			}
		}
	}
	if kind != reflect.Ptr {
		c.factoryInfos[out] = factoryInfo{target: value, lifetimeScope: lts, ins: ins, isFunc: isFunc}
		_, ok := c.cache[out]
		if ok {
			delete(c.cache, out)
		}
		count++
	} else if count == 0 {
		return ErrNeedInterfaceOnPointerRegistering
	}
	return nil
}

// Invoke はコンテナからインスタンスを解決して呼び出します
func (c *container) Invoke(invoker Invoker) error {
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
		v, err := c.resolve(in, &cache)
		if err != nil {
			return err
		}
		args[i] = *v
	}

	fn := reflect.ValueOf(invoker)
	outs := fn.Call(args)
	if err := c.getError(outs); err != nil {
		return err
	}
	return nil
}

func (c *container) getError(outs []reflect.Value) error {
	l := len(outs)
	if l > 0 {
		if err, ok := outs[l-1].Interface().(error); ok {
			return err
		}
	}
	return nil
}

func (c *container) resolve(t reflect.Type, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	if c.containerInterfaceType == t || c.ioCContainerInterfaceType == t || c.serviceLocatorInterfaceType == t {
		v := reflect.ValueOf(c)
		return &v, nil
	}
	factoryInfo, ok := c.factoryInfos[t]
	if !ok {
		return nil, newErrInvalidResolveComponent(t)
	}
	switch factoryInfo.lifetimeScope {
	case ContainerManaged:
		return c.resolveContainerManagedObject(t, factoryInfo, cache)
	}
	return c.resolveInvokeManagedObject(t, factoryInfo, cache)
}
func (c *container) resolveContainerManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	if v, ok := c.cache[t]; ok {
		return &v, nil
	}
	if !factoryInfo.isFunc {
		c.cache[t] = factoryInfo.target
		return &factoryInfo.target, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := c.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	for _, out := range outs {
		if err, ok := out.Interface().(error); ok {
			return nil, err
		}
	}
	if err := c.getError(outs); err != nil {
		return nil, err
	}
	out := outs[0]
	c.cache[t] = out
	return &out, nil
}
func (c *container) resolveInvokeManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	cch := *cache
	if v, ok := cch[t]; ok {
		return &v, nil
	}
	if !factoryInfo.isFunc {
		cch[t] = factoryInfo.target
		return &factoryInfo.target, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := c.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	if err := c.getError(outs); err != nil {
		return nil, err
	}
	out := outs[0]
	cch[t] = out
	return &out, nil
}

func (c *container) Verify() error {
	lenIns := len(c.factoryInfos)
	if lenIns == 0 {
		return ErrNotFoundComponent
	}
	args := make([]reflect.Value, lenIns)
	cache := make(map[reflect.Type]reflect.Value)
	i := 0
	for t := range c.factoryInfos {
		v, err := c.resolve(t, &cache)
		if err != nil {
			return err
		}
		args[i] = *v
		i++
	}
	return nil
}
