package dijct

import (
	"fmt"
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
	len := t.NumOut()
	if len != 1 {
		return nil, fmt.Errorf("コンストラクタの戻り値は単一である必要があります。")
	}
	return t.Out(1), nil
}
func getTargetReflectionInfos(target Target) (out reflect.Type, in []reflect.Type, err error) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Func {
		out, err := getOut(t)
		if t != nil {
			return nil, nil, err
		}
		ins := getIns(t)
		return out, ins, nil
	}
	return t, nil, nil
}
