package stringify

import (
	"fmt"
	"reflect"
	"unsafe"
)

func IsInterfaceNil(param interface{}) bool {
	return param == nil || (*[2]uintptr)(unsafe.Pointer(&param))[1] == 0
}

func Interface(value interface{}) string {
	if IsInterfaceNil(value) {
		return "<nil>"
	}

	switch typeCastedValue := value.(type) {
	case bool:
		return Bool(typeCastedValue)
	case string:
		return String(typeCastedValue)
	case []byte:
		return SliceOfBytes(typeCastedValue)
	case int:
		return Int(int64(typeCastedValue))
	case int8:
		return Int(int64(typeCastedValue))
	case int16:
		return Int(int64(typeCastedValue))
	case int32:
		return Int(int64(typeCastedValue))
	case int64:
		return Int(typeCastedValue)
	case uint8:
		return UInt(uint64(typeCastedValue))
	case uint16:
		return UInt(uint64(typeCastedValue))
	case uint32:
		return UInt(uint64(typeCastedValue))
	case uint64:
		return UInt(typeCastedValue)
	case float64:
		return Float64(typeCastedValue)
	case float32:
		return Float32(typeCastedValue)
	case reflect.Value:
		switch typeCastedValue.Kind() {
		case reflect.Slice:
			return sliceReflect(typeCastedValue)
		case reflect.Array:
			return sliceReflect(typeCastedValue)
		case reflect.String:
			return Interface(typeCastedValue.String())
		case reflect.Int:
			return Interface(typeCastedValue.Int())
		case reflect.Int8:
			return Interface(typeCastedValue.Int())
		case reflect.Int16:
			return Interface(typeCastedValue.Int())
		case reflect.Int32:
			return Interface(typeCastedValue.Int())
		case reflect.Int64:
			return Interface(typeCastedValue.Int())
		case reflect.Uint8:
			return Interface(typeCastedValue.Uint())
		case reflect.Uint16:
			return Interface(typeCastedValue.Uint())
		case reflect.Uint32:
			return Interface(typeCastedValue.Uint())
		case reflect.Uint64:
			return Interface(typeCastedValue.Uint())
		case reflect.Ptr:
			return Interface(typeCastedValue.Interface())
		case reflect.Struct:
			return fmt.Sprint(value)
		default:
			panic("undefined reflect type: " + typeCastedValue.Kind().String())
		}
	case fmt.Stringer:
		return typeCastedValue.String()
	default:
		value := reflect.ValueOf(value)
		switch value.Kind() {
		case reflect.Slice:
			return sliceReflect(value)
		case reflect.Map:
			return mapReflect(value)
		default:
			panic("undefined type: " + value.Kind().String())
		}
	}
}
