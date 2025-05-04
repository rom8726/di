package di

import (
	"reflect"
)

type provider struct {
	name        string
	constructor reflect.Value
	returnType  reflect.Type
	paramTypes  []reflect.Type
	initFunc    func(args []any) (any, error)
}

func newProvider(constructor any) *provider {
	ctor := reflect.ValueOf(constructor)
	ctorType := ctor.Type()

	if ctorType.Kind() != reflect.Func || ctorType.NumOut() < 1 || ctorType.NumOut() > 2 {
		panic("constructor must be a function returning one value (instance) or two values (instance, error)")
	}

	retType := ctorType.Out(0)
	numIn := ctorType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = ctorType.In(i)
	}

	initFunc := func(args []interface{}) (interface{}, error) {
		var argv []reflect.Value
		for _, arg := range args {
			argv = append(argv, reflect.ValueOf(arg))
		}

		out := ctor.Call(argv)
		if len(out) == 1 {
			return out[0].Interface(), nil
		}

		result := out[0].Interface()
		err := out[1].Interface()
		if err != nil {
			return result, err.(error)
		}

		return result, nil
	}

	return &provider{
		name:        ctorType.String(),
		constructor: ctor,
		returnType:  retType,
		paramTypes:  paramTypes,
		initFunc:    initFunc,
	}
}
