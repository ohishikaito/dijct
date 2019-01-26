package dijct

import "reflect"

type (
	factoryInfo struct {
		target        reflect.Value
		ins           []reflect.Type
		isFunc        bool
		lifetimeScope LifetimeScope
	}
)
