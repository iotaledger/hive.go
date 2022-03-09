package serix

import "reflect"

func deReferencePointer(value reflect.Value) reflect.Value {
	value.Type()
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value
}
