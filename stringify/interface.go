package stringify

import (
	"fmt"
	"reflect"
	"strconv"
)

func Interface(value interface{}) string {
	switch typeCastedValue := value.(type) {
	case bool:
		return Bool(typeCastedValue)
	case string:
		return String(typeCastedValue)
	case []byte:
		return SliceOfBytes(typeCastedValue)
	case int:
		return Int(typeCastedValue)
	case uint64:
		return strconv.FormatUint(typeCastedValue, 10)
	case reflect.Value:
		switch typeCastedValue.Kind() {
		case reflect.Slice:
			return sliceReflect(typeCastedValue)
		case reflect.Array:
			return sliceReflect(typeCastedValue)
		case reflect.String:
			return String(typeCastedValue.String())
		case reflect.Int:
			return Int(int(typeCastedValue.Int()))
		case reflect.Uint8:
			return Int(int(typeCastedValue.Uint()))
		case reflect.Ptr:
			return Interface(typeCastedValue.Interface())
		case reflect.Struct:
			return fmt.Sprint(value)
		default:
			panic("undefined reflect type: " + typeCastedValue.Kind().String())
		}
	case fmt.Stringer:
		return value.(fmt.Stringer).String()
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
