package serix

import (
	"reflect"
	"strings"

	"github.com/iotaledger/hive.go/ierrors"
)

// checks whether the given value has the concept of a length.
func hasLength(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Array:
	case reflect.Map:
	case reflect.Slice:
	case reflect.String:
	default:
		return false
	}

	return true
}

func sliceFromArray(arrValue reflect.Value) reflect.Value {
	arrType := arrValue.Type()
	sliceType := reflect.SliceOf(arrType.Elem())
	sliceValue := reflect.MakeSlice(sliceType, arrType.Len(), arrType.Len())
	reflect.Copy(sliceValue, arrValue)

	return sliceValue
}

func fillArrayFromSlice(arrayValue, sliceValue reflect.Value) {
	for i := range sliceValue.Len() {
		arrayValue.Index(i).Set(sliceValue.Index(i))
	}
}

func isUnderlyingStruct(t reflect.Type) bool {
	t = DeRefPointer(t)

	return t.Kind() == reflect.Struct
}

func isUnderlyingInterface(t reflect.Type) bool {
	t = DeRefPointer(t)

	return t.Kind() == reflect.Interface
}

// DeRefPointer dereferences the given type if it's a pointer.
func DeRefPointer(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func checkDecodeDestination(obj any, value reflect.Value) error {
	if !value.IsValid() {
		return ierrors.New("invalid value for destination")
	}
	if value.Kind() != reflect.Ptr {
		return ierrors.Errorf(
			"can't decode, the destination object must be a pointer, got: %T(%s)", obj, value.Kind(),
		)
	}
	if value.IsNil() {
		return ierrors.Errorf("can't decode, the destination object %T must be a non-nil pointer", obj)
	}

	return nil
}

func getNumberTypeToConvert(kind reflect.Kind) (int, reflect.Type, reflect.Type) {
	var numberType reflect.Type
	var bitSize int
	switch kind {
	case reflect.Int8:
		numberType = reflect.TypeOf(int8(0))
		bitSize = 8
	case reflect.Int16:
		numberType = reflect.TypeOf(int16(0))
		bitSize = 16
	case reflect.Int32:
		numberType = reflect.TypeOf(int32(0))
		bitSize = 32
	case reflect.Int64:
		numberType = reflect.TypeOf(int64(0))
		bitSize = 64
	case reflect.Uint8:
		numberType = reflect.TypeOf(uint8(0))
		bitSize = 8
	case reflect.Uint16:
		numberType = reflect.TypeOf(uint16(0))
		bitSize = 16
	case reflect.Uint32:
		numberType = reflect.TypeOf(uint32(0))
		bitSize = 32
	case reflect.Uint64:
		numberType = reflect.TypeOf(uint64(0))
		bitSize = 64
	case reflect.Float32:
		numberType = reflect.TypeOf(float32(0))
		bitSize = 32
	case reflect.Float64:
		numberType = reflect.TypeOf(float64(0))
		bitSize = 64
	default:
		return -1, nil, nil
	}

	return bitSize, numberType, reflect.PointerTo(numberType)
}

// FieldKeyString converts the given string to camelCase.
// Special keywords like ID or URL are converted to only first letter upper case.
func FieldKeyString(str string) string {
	for _, keyword := range []string{"ID", "NFT", "URL", "HRP"} {
		if !strings.Contains(str, keyword) {
			continue
		}

		// first keyword letter upper case, rest lower case
		str = strings.ReplaceAll(str, keyword, string(keyword[0])+strings.ToLower(keyword)[1:])
	}

	// first letter lower case
	return strings.ToLower(str[:1]) + str[1:]
}
