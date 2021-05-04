package dijct

import (
	"reflect"
)

func getIns(t reflect.Type) []reflect.Type {
	len := t.NumIn()
	in := make([]reflect.Type, len)
	for i := 0; i < len; i++ {
		in[i] = t.In(i)
	}
	return in
}
func getOut(t reflect.Type) (reflect.Type, error) {
	l := t.NumOut()
	if l < 1 {
		return nil, ErrRequireResponse
	}
	return t.Out(0), nil
}
func getTargetReflectionInfos(target Target) (out reflect.Type, in []reflect.Type, err error) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Func {
		out, err := getOut(t)
		if err != nil {
			return nil, nil, err
		}
		ins := getIns(t)
		return out, ins, nil
	}
	return t, nil, nil
}
