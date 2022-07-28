package reflect

import (
	"reflect"
)

// copyFuncSignature returns an interface of a new function with the same function signature as 'fnc'.
func copyFuncSignature(fnc interface{}, new func(args []reflect.Value) (results []reflect.Value)) interface{} {

	x := reflect.TypeOf(fnc)
	if x.Kind() != reflect.Func {
		panic("must pass a function to copyFunc")
	}

	in := make([]reflect.Type, x.NumIn())
	for i := 0; i < x.NumIn(); i++ {
		in[i] = x.In(i)
	}

	out := make([]reflect.Type, x.NumOut())
	for o := 0; o < x.NumOut(); o++ {
		out[o] = x.Out(o)
	}

	return reflect.MakeFunc(reflect.FuncOf(in, out, x.IsVariadic()), new).Interface()
}

// FuncPreCallback returns an interface to a function that calls 'callback' before calling 'fnc'.
func FuncPreCallback(fnc interface{}, callback func()) interface{} {
	return copyFuncSignature(fnc, func(args []reflect.Value) (results []reflect.Value) {
		callback()
		results = reflect.ValueOf(fnc).Call(args)
		return results
	})
}

// FuncPostCallback returns an interface to a function that calls 'callback' after calling 'fnc'.
func FuncPostCallback(fnc interface{}, callback func()) interface{} {
	return copyFuncSignature(fnc, func(args []reflect.Value) (results []reflect.Value) {
		results = reflect.ValueOf(fnc).Call(args)
		callback()
		return results
	})
}
