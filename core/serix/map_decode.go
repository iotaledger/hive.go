package serix

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
)

func (api *API) mapDecode(ctx context.Context, mapVal any, value reflect.Value, ts TypeSettings, opts *options) (err error) {
	var deserializable DeserializableJSON

	if _, ok := value.Interface().(DeserializableJSON); ok {
		if value.Kind() == reflect.Ptr && value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		deserializable = value.Interface().(DeserializableJSON)
	} else if value.CanAddr() {
		if addrDeserializable, ok := value.Addr().Interface().(DeserializableJSON); ok {
			deserializable = addrDeserializable
		}
	}

	if deserializable != nil {
		err = deserializable.DecodeJSON(mapVal)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err = api.mapDecodeBasedOnType(ctx, mapVal, value, value.Type(), ts, opts); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (api *API) mapDecodeBasedOnType(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	globalTS, _ := api.getTypeSettings(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueType == bigIntPtrType {
			bigIntHexStr, ok := mapVal.(string)
			if !ok {
				return fmt.Errorf("non string value in map when decoding a big.Int, got %T instead", mapVal)
			}
			bigInt, err := DecodeUint256(bigIntHexStr)
			if err != nil {
				return errors.Wrap(err, "failed to read big.Int from map")
			}
			value.Addr().Elem().Set(reflect.ValueOf(bigInt))

			return nil
		}

		elemType := valueType.Elem()
		if elemType.Kind() == reflect.Struct {
			if value.IsNil() {
				value.Set(reflect.New(elemType))
			}
			elemValue := value.Elem()

			return api.mapDecodeStruct(ctx, mapVal, elemValue, elemType, ts, opts)
		}

		if elemType.Kind() == reflect.Array {
			if value.IsNil() {
				value.Set(reflect.New(elemType))
			}
			sliceValue := sliceFromArray(value.Elem())
			sliceValueType := sliceValue.Type()
			if sliceValueType.AssignableTo(bytesType) {
				innerTS, ok := api.getTypeSettings(valueType)
				if !ok {
					return errors.Errorf("missing type settings for interface %s", valueType)
				}

				mapKey := mapSliceArrayDefaultKey
				if innerTS.mapKey != nil {
					mapKey = *innerTS.mapKey
				}

				fieldValStr := mapVal.(map[string]any)[mapKey].(string)
				byteSlice, err := DecodeHex(fieldValStr)
				if err != nil {
					return errors.Wrap(err, "failed to read byte slice from map")
				}
				copy(sliceValue.Bytes(), byteSlice)
				fillArrayFromSlice(value.Elem(), sliceValue)

				return nil
			}

			return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, opts)
		}

	case reflect.Struct:
		return api.mapDecodeStruct(ctx, mapVal, value, valueType, ts, opts)
	case reflect.Slice:
		return api.mapDecodeSlice(ctx, mapVal, value, valueType, opts)
	case reflect.Map:
		return api.mapDecodeMap(ctx, mapVal, value, valueType, opts)
	case reflect.Array:
		sliceValue := sliceFromArray(value)
		sliceValueType := sliceValue.Type()
		if sliceValueType.AssignableTo(bytesType) {
			byteSlice, err := DecodeHex(mapVal.(string))
			if err != nil {
				return errors.Wrap(err, "failed to read byte slice from map")
			}
			copy(sliceValue.Bytes(), byteSlice)
			fillArrayFromSlice(value, sliceValue)

			return nil
		}

		return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, opts)
	case reflect.Interface:
		return api.mapDecodeInterface(ctx, mapVal, value, valueType, ts, opts)
	case reflect.String:
		str, ok := mapVal.(string)
		if !ok {
			return fmt.Errorf("non string value for string field")
		}
		if !utf8.ValidString(str) {
			return ErrNonUTF8String
		}
		addrValue := value.Addr().Convert(reflect.TypeOf((*string)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))

		return nil
	case reflect.Bool:
		addrValue := value.Addr().Convert(reflect.TypeOf((*bool)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))

		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32:
		return api.mapDecodeNum(value, valueType, float64NumParser(mapVal.(float64), value.Kind(), true))
	case reflect.Int64:
		return api.mapDecodeNum(value, valueType, strNumParser(mapVal.(string), 64, true))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return api.mapDecodeNum(value, valueType, float64NumParser(mapVal.(float64), value.Kind(), false))
	case reflect.Uint64:
		return api.mapDecodeNum(value, valueType, strNumParser(mapVal.(string), 64, false))
	case reflect.Float32, reflect.Float64:
		return api.mapDecodeFloat(value, valueType, mapVal)
	default:
	}

	return errors.Errorf("can't map decode: unsupported type %s", valueType)
}

// num parse func returns a num or an error.
type numParseFunc func() (any, error)

func float64NumParser(v float64, ty reflect.Kind, signed bool) numParseFunc {
	return func() (any, error) {
		if signed {
			switch ty {
			case reflect.Int8:
				return int8(v), nil
			case reflect.Int16:
				return int16(v), nil
			case reflect.Int32:
				return int32(v), nil
			default:
				return nil, errors.Errorf("can not map decode kind %s to signed integer", ty)
			}
		}
		switch ty {
		case reflect.Uint8:
			return uint8(v), nil
		case reflect.Uint16:
			return uint16(v), nil
		case reflect.Uint32:
			return uint32(v), nil
		default:
			return nil, errors.Errorf("can not map decode kind %s to unsigned integer", ty)
		}
	}
}

func strNumParser(str string, bitSize int, signed bool) numParseFunc {
	return func() (any, error) {
		if signed {
			return strconv.ParseInt(str, 10, bitSize)
		}

		return strconv.ParseUint(str, 10, bitSize)
	}
}

func (api *API) mapDecodeNum(value reflect.Value, valueType reflect.Type, parser numParseFunc) error {
	addrValue := value.Addr()
	_, _, addrTypeToConvert := getNumberTypeToConvert(valueType.Kind())
	addrValue = addrValue.Convert(addrTypeToConvert)

	num, err := parser()
	if err != nil {
		return err
	}

	addrValue.Elem().Set(reflect.ValueOf(num))

	return nil
}

func (api *API) mapDecodeFloat(value reflect.Value, valueType reflect.Type, mapVal any) error {
	addrValue := value.Addr()
	bitSize, _, addrTypeToConvert := getNumberTypeToConvert(valueType.Kind())
	addrValue = addrValue.Convert(addrTypeToConvert)

	f, err := strconv.ParseFloat(mapVal.(string), bitSize)
	if err != nil {
		return err
	}
	addrValue.Elem().SetFloat(f)

	return nil
}

func (api *API) mapDecodeInterface(
	ctx context.Context, mapVal any, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) error {
	iObjects := api.getInterfaceObjects(valueType)
	if iObjects == nil {
		return errors.Errorf("interface %s hasn't been registered", valueType)
	}

	m, ok := mapVal.(map[string]any)
	if !ok {
		return errors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	objectCodeAny, has := m[mapTypeKeyName]
	if !has {
		return errors.Errorf("no object type defined in map for interface %s", valueType)
	}
	objectCode := uint32(objectCodeAny.(float64))

	objectType := iObjects.fromCodeToType[objectCode]
	if objectType == nil {
		return errors.Errorf("no object type with code %d was found for interface %s", objectCode, valueType)
	}

	objectValue := reflect.New(objectType).Elem()
	if err := api.mapDecode(ctx, m, objectValue, ts, opts); err != nil {
		return errors.WithStack(err)
	}
	value.Set(objectValue)

	return nil
}

func (api *API) mapDecodeStruct(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	if valueType == timeType {
		strVal := mapVal.(string)
		parsedTime, err := time.Parse(time.RFC3339Nano, strVal)
		if err != nil {
			return fmt.Errorf("unable to parse time %s map value: %w", strVal, err)
		}
		value.Set(reflect.ValueOf(parsedTime))

		return nil
	}

	m, ok := mapVal.(map[string]any)
	if !ok {
		return errors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	if objectType := ts.ObjectType(); objectType != nil {
		_, objectCode, err := getTypeDenotationAndCode(objectType)
		if err != nil {
			return errors.WithStack(err)
		}
		mapObjectCode, has := m[mapTypeKeyName]
		if !has {
			return errors.Wrap(err, "missing type key in struct")
		}
		if uint32(mapObjectCode.(float64)) != objectCode {
			return errors.Errorf("map type key (%d) not equal registered object code (%d)", mapObjectCode, objectCode)
		}
	}

	if err := api.mapDecodeStructFields(ctx, m, value, valueType, opts); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (api *API) mapDecodeStructFields(
	ctx context.Context, m map[string]any, structVal reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := api.parseStructType(valueType)
	if err != nil {
		return errors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := structVal.Field(sField.index)
		if sField.isEmbeddedStruct && !sField.settings.nest {
			fieldType := sField.fType
			if fieldType.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					if sField.isUnexported {
						return errors.Errorf(
							"embedded field %s is a nil pointer, can't initialize because it's unexported",
							sField.name,
						)
					}
					fieldValue.Set(reflect.New(fieldType.Elem()))
				}
				fieldValue = fieldValue.Elem()
				fieldType = fieldType.Elem()
			}
			if err := api.mapDecodeStructFields(ctx, m, fieldValue, fieldType, opts); err != nil {
				return errors.Wrapf(err, "can't deserialize embedded struct %s", sField.name)
			}

			continue
		}

		key := mapStringKey(sField.name)
		if sField.settings.ts.mapKey != nil {
			key = sField.settings.ts.MustMapKey()
		}

		mapVal, has := m[key]
		if !has {
			if sField.settings.isOptional || sField.settings.omitEmpty {
				continue
			}

			return errors.Wrapf(err, "missing map entry for field %s", sField.name)
		}

		if err := api.mapDecode(ctx, mapVal, fieldValue, sField.settings.ts, opts); err != nil {
			return errors.Wrapf(err, "failed to deserialize struct field %s", sField.name)
		}
	}

	return nil
}

func (api *API) mapDecodeSlice(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, opts *options) error {
	if valueType.AssignableTo(bytesType) {
		fieldValStr := mapVal.(string)
		byteSlice, err := DecodeHex(fieldValStr)
		if err != nil {
			return errors.Wrap(err, "failed to read byte slice from map")
		}

		addrValue := value.Addr().Convert(reflect.TypeOf((*[]byte)(nil)))
		addrValue.Elem().SetBytes(byteSlice)

		return nil
	}

	refVal := reflect.ValueOf(mapVal)
	for i := 0; i < refVal.Len(); i++ {
		elemValue := reflect.New(valueType.Elem()).Elem()
		if err := api.mapDecode(ctx, refVal.Index(i).Interface(), elemValue, TypeSettings{}, opts); err != nil {
			return errors.WithStack(err)
		}
		value.Set(reflect.Append(value, elemValue))
	}

	return nil
}

func (api *API) mapDecodeMap(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, opts *options) error {
	m, ok := mapVal.(map[string]any)
	if !ok {
		return errors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	if value.IsNil() {
		value.Set(reflect.MakeMap(valueType))
	}

	for k, v := range m {
		key := reflect.New(valueType.Key()).Elem()
		val := reflect.New(valueType.Elem()).Elem()

		if err := api.mapDecode(ctx, k, key, TypeSettings{}, opts); err != nil {
			return errors.Wrapf(err, "failed to map decode map key of type %s", key.Type())
		}

		if err := api.mapDecode(ctx, v, val, TypeSettings{}, opts); err != nil {
			return errors.Wrapf(err, "failed to map decode map element of type %s", val.Type())
		}

		value.SetMapIndex(key, val)
	}

	return nil
}
