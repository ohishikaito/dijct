package dijct

import (
	"reflect"
)

type (
	// RegisterOptions は 登録時のオプションです
	RegisterOptions struct {
		LifetimeScope LifetimeScope
		Interfaces    []reflect.Type
	}
)
