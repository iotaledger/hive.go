package serix

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"math/big"
	"reflect"
	"strconv"
)

func (api *API) mapDecode(ctx context.Context, mapVal any, value reflect.Value, ts TypeSettings, opts *options) error {
	valueType := value.Type()
	if opts.validation {
		// TODO: add map validator?
	}

	if err := api.mapDecodeBasedOnType(ctx, mapVal, value, valueType, ts, opts); err != nil {
		return errors.WithStack(err)
	}

	if opts.validation {
		// TODO: add post map validator?
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
			bigIntHexStr := mapVal.(string)
			bigIntBytes, err := hexutil.Decode(bigIntHexStr)
			if err != nil {
				return errors.Wrap(err, "failed to read big.Int from map")
			}
			value.Addr().Elem().Set(reflect.ValueOf(new(big.Int).SetBytes(bigIntBytes)))
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
				innerTs, ok := api.getTypeSettings(valueType)
				if !ok {
					return errors.Errorf("missing type settings for interface %s", valueType)
				}

				if innerTs.mapKey == nil {
					return fmt.Errorf("missing map key for slice")
				}

				fieldValStr := mapVal.(map[string]any)[*innerTs.mapKey].(string)
				byteSlice, err := DecodeHex(fieldValStr)
				if err != nil {
					return errors.Wrap(err, "failed to read byte slice from map")
				}
				copy(sliceValue.Bytes(), byteSlice)
				fillArrayFromSlice(value.Elem(), sliceValue)
				return nil
			}
			return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, ts, opts)
		}

	case reflect.Struct:
		return api.mapDecodeStruct(ctx, mapVal, value, valueType, ts, opts)
	case reflect.Slice:
		return api.mapDecodeSlice(ctx, mapVal, value, valueType, ts, opts)
	case reflect.Map:
		return api.mapDecodeMap(ctx, mapVal, value, valueType, ts, opts)
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
		return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.mapDecodeInterface(ctx, mapVal, value, valueType, ts, opts)
	case reflect.String:
		addrValue := value.Addr().Convert(reflect.TypeOf((*string)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))
		return nil
	case reflect.Bool:
		addrValue := value.Addr().Convert(reflect.TypeOf((*bool)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))
		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return api.mapDecodeInteger(value, valueType, mapVal, true)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return api.mapDecodeInteger(value, valueType, mapVal, false)
	case reflect.Float32:
		return api.mapDecodeFloat(value, valueType, mapVal, 32)
	case reflect.Float64:
		return api.mapDecodeFloat(value, valueType, mapVal, 64)
	default:
	}
	return errors.Errorf("can't map decode: unsupported type %s", valueType)
}

func (api *API) mapDecodeInteger(value reflect.Value, valueType reflect.Type, mapVal any, signed bool) error {
	addrValue := value.Addr()
	bitSize, _, addrTypeToConvert := getNumberTypeToConvert(valueType.Kind())
	addrValue = addrValue.Convert(addrTypeToConvert)

	if signed {
		num, err := strconv.ParseInt(mapVal.(string), 10, bitSize)
		if err != nil {
			return err
		}
		addrValue.Elem().SetInt(num)
		return nil
	}

	num, err := strconv.ParseUint(mapVal.(string), 10, bitSize)
	if err != nil {
		return err
	}
	addrValue.Elem().SetUint(num)
	return nil
}

func (api *API) mapDecodeFloat(value reflect.Value, valueType reflect.Type, mapVal any, bitSize int) error {
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
		// TODO
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
	structFields, err := parseStructType(valueType)
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
				// TODO: how can an embedded struct be of kind pointer?
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

		mapVal, has := m[sField.settings.ts.MustMapKey()]
		if !has {
			if sField.settings.isOptional {
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
	valueType reflect.Type, ts TypeSettings, opts *options) error {
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
	valueType reflect.Type, ts TypeSettings, opts *options) error {
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
