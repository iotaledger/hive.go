package serix

import (
	"context"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/iotaledger/hive.go/ierrors"
)

func (api *API) mapDecode(ctx context.Context, mapVal any, value reflect.Value, ts TypeSettings, opts *options) (err error) {
	var deserializable DeserializableJSON

	if _, ok := value.Interface().(DeserializableJSON); ok {
		if value.Kind() == reflect.Ptr && value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		//nolint:forcetypeassert // false positive
		deserializable = value.Interface().(DeserializableJSON)
	} else if value.CanAddr() {
		if addrDeserializable, ok := value.Addr().Interface().(DeserializableJSON); ok {
			deserializable = addrDeserializable
		}
	}

	if deserializable != nil {
		err = deserializable.DecodeJSON(mapVal)
		if err != nil {
			return ierrors.WithStack(err)
		}

		if contextAwareDeserializable, ok := deserializable.(ContextAwareDeserializable); ok {
			contextAwareDeserializable.SetDeserializationContext(ctx)
		}
	} else {
		if err = api.mapDecodeBasedOnType(ctx, mapVal, value, value.Type(), ts, opts); err != nil {
			return ierrors.WithStack(err)
		}
	}

	if opts.validation {
		if err := api.callSyntacticValidator(ctx, value, value.Type()); err != nil {
			return ierrors.Wrap(err, "post-serialization validation failed")
		}
	}

	return nil
}

func (api *API) mapDecodeBasedOnType(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	globalTS, _ := api.typeSettingsRegistry.GetByType(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueType == bigIntPtrType {
			bigIntHexStr, ok := mapVal.(string)
			if !ok {
				return ierrors.Errorf("non string value in map when decoding a big.Int, got %T instead", mapVal)
			}
			bigInt, err := DecodeUint256(bigIntHexStr)
			if err != nil {
				return ierrors.Wrap(err, "failed to read big.Int from map")
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

			if contextAwareDeserializable, ok := value.Interface().(ContextAwareDeserializable); ok {
				contextAwareDeserializable.SetDeserializationContext(ctx)
			}

			return api.mapDecodeStruct(ctx, mapVal, elemValue, elemType, ts, opts)
		}

		if elemType.Kind() == reflect.Interface {
			if value.IsNil() {
				value.Set(reflect.New(elemType))
			}
			elemValue := value.Elem()

			return api.mapDecodeInterface(ctx, mapVal, elemValue, elemType, ts, opts)
		}

		if elemType.Kind() == reflect.Array {
			if value.IsNil() {
				value.Set(reflect.New(elemType))
			}
			sliceValue := sliceFromArray(value.Elem())
			sliceValueType := sliceValue.Type()
			if sliceValueType.AssignableTo(bytesType) {
				innerTS, ok := api.typeSettingsRegistry.GetByType(valueType)
				if !ok {
					return ierrors.Errorf("missing type settings for interface %s", valueType)
				}

				fieldKey := keyDefaultSliceArray
				if innerTS.fieldKey != nil {
					fieldKey = *innerTS.fieldKey
				}

				//nolint:forcetypeassert
				fieldValStr := mapVal.(map[string]any)[fieldKey].(string)
				byteSlice, err := DecodeHex(fieldValStr)
				if err != nil {
					return ierrors.Wrap(err, "failed to read byte slice from map")
				}

				if opts.validation {
					if err := ts.checkMinMaxBoundsLength(len(byteSlice)); err != nil {
						return ierrors.Wrapf(err, "can't deserialize '%s' type", value.Kind())
					}
				}

				copy(sliceValue.Bytes(), byteSlice)
				fillArrayFromSlice(value.Elem(), sliceValue)

				return nil
			}

			return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, ts, opts)
		}

	case reflect.Struct:
		if contextAwareDeserializable, ok := value.Interface().(ContextAwareDeserializable); ok {
			contextAwareDeserializable.SetDeserializationContext(ctx)
		}

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
				return ierrors.Wrap(err, "failed to read byte slice from map")
			}
			copy(sliceValue.Bytes(), byteSlice)
			fillArrayFromSlice(value, sliceValue)

			return nil
		}

		return api.mapDecodeSlice(ctx, mapVal, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.mapDecodeInterface(ctx, mapVal, value, valueType, ts, opts)
	case reflect.String:
		str, ok := mapVal.(string)
		if !ok {
			return ierrors.New("non string value for string field")
		}

		if opts.validation {
			if err := ts.checkMinMaxBoundsLength(len(str)); err != nil {
				return ierrors.Wrapf(err, "can't deserialize '%s' type", value.Kind())
			}
			// check the string for UTF-8 validity
			if !utf8.ValidString(str) {
				return ErrNonUTF8String
			}
		}

		addrValue := value.Addr().Convert(reflect.TypeOf((*string)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))

		return nil
	case reflect.Bool:
		addrValue := value.Addr().Convert(reflect.TypeOf((*bool)(nil)))
		addrValue.Elem().Set(reflect.ValueOf(mapVal))

		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32:
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		return api.mapDecodeNum(value, valueType, float64NumParser(mapVal.(float64), value.Kind(), true))
	case reflect.Int64:
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		return api.mapDecodeNum(value, valueType, strNumParser(mapVal.(string), 64, true))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		return api.mapDecodeNum(value, valueType, float64NumParser(mapVal.(float64), value.Kind(), false))
	case reflect.Uint64:
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		return api.mapDecodeNum(value, valueType, strNumParser(mapVal.(string), 64, false))
	case reflect.Float32, reflect.Float64:
		return api.mapDecodeFloat(value, valueType, mapVal)
	default:
	}

	return ierrors.Errorf("can't map decode: unsupported type %s", valueType)
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
				return nil, ierrors.Errorf("can not map decode kind %s to signed integer", ty)
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
			return nil, ierrors.Errorf("can not map decode kind %s to unsigned integer", ty)
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
		return ierrors.Errorf("interface %s hasn't been registered", valueType)
	}

	m, ok := mapVal.(map[string]any)
	if !ok {
		return ierrors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	objectCodeAny, has := m[keyType]
	if !has {
		return ierrors.Errorf("no object type defined in map for interface %s", valueType)
	}
	//nolint:forcetypeassert // false positive
	objectCode := uint32(objectCodeAny.(float64))

	objectType, exists := iObjects.GetObjectTypeByCode(objectCode)
	if !exists || objectType == nil {
		return ierrors.Wrapf(ErrInterfaceUnderlyingTypeNotRegistered, "object code: %d, interface: %s", objectCode, valueType)
	}

	objectValue := reflect.New(objectType).Elem()
	if err := api.mapDecode(ctx, m, objectValue, ts, opts); err != nil {
		return ierrors.WithStack(err)
	}
	value.Set(objectValue)

	return nil
}

func (api *API) mapDecodeStruct(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	if valueType == timeType {
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		strVal := mapVal.(string)
		nanoTime, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return ierrors.Wrapf(err, "unable to parse time %s map value", strVal)
		}

		value.Set(reflect.ValueOf(time.Unix(0, int64(nanoTime)).UTC()))

		return nil
	}

	m, ok := mapVal.(map[string]any)
	if !ok {
		return ierrors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	if objectType := ts.ObjectType(); objectType != nil {
		_, objectCode, err := getTypeDenotationAndCode(objectType)
		if err != nil {
			return ierrors.WithStack(err)
		}
		mapObjectCode, has := m[keyType]
		if !has {
			return ierrors.Wrap(err, "missing type key in struct")
		}
		if castedMapObjectCode, ok := mapObjectCode.(float64); !ok || uint32(castedMapObjectCode) != objectCode {
			return ierrors.Errorf("map type key (%d) not equal registered object code (%d)", mapObjectCode, objectCode)
		}
	}

	if err := api.mapDecodeStructFields(ctx, m, value, valueType, opts); err != nil {
		return ierrors.WithStack(err)
	}

	return nil
}

func (api *API) mapDecodeStructFields(
	ctx context.Context, m map[string]any, structVal reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := api.getStructFields(valueType)
	if err != nil {
		return ierrors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := structVal.Field(sField.index)
		if sField.isEmbedded && !sField.settings.inlined {
			fieldType := sField.fType
			if fieldType.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					if sField.isUnexported {
						return ierrors.Errorf(
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
				return ierrors.Wrapf(err, "can't deserialize embedded struct %s", sField.name)
			}

			continue
		}

		if sField.settings.inlined {
			if err := api.mapDecode(ctx, m, fieldValue, sField.settings.ts, opts); err != nil {
				return ierrors.Wrapf(err, "failed to deserialize inlined struct field %s", sField.name)
			}

			continue
		}

		fieldKey := FieldKeyString(sField.name)
		if sField.settings.ts.fieldKey != nil {
			fieldKey = sField.settings.ts.MustFieldKey()
		}

		mapVal, has := m[fieldKey]
		if !has {
			if sField.settings.isOptional || sField.settings.omitEmpty {
				// initialize an empty slice if the kind is slice and its a nil pointer
				if fieldValue.Kind() == reflect.Slice && fieldValue.IsNil() {
					fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), 0, 0))
				}

				continue
			}

			return ierrors.Errorf("missing map entry for field %s", sField.name)
		}

		if err := api.mapDecode(ctx, mapVal, fieldValue, sField.settings.ts, opts); err != nil {
			return ierrors.Wrapf(err, "failed to deserialize struct field %s", sField.name)
		}
	}

	return nil
}

func (api *API) mapDecodeSlice(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	if valueType.AssignableTo(bytesType) {
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		fieldValStr := mapVal.(string)
		byteSlice, err := DecodeHex(fieldValStr)
		if err != nil {
			return ierrors.Wrap(err, "failed to read byte slice from map")
		}

		if opts.validation {
			if err := ts.checkMinMaxBoundsLength(len(byteSlice)); err != nil {
				return ierrors.Wrapf(err, "can't deserialize '%s' type", value.Kind())
			}
		}

		addrValue := value.Addr().Convert(reflect.TypeOf((*[]byte)(nil)))
		addrValue.Elem().SetBytes(byteSlice)

		return nil
	}

	refVal := reflect.ValueOf(mapVal)
	for i := range refVal.Len() {
		elemValue := reflect.New(valueType.Elem()).Elem()
		if err := api.mapDecode(ctx, refVal.Index(i).Interface(), elemValue, TypeSettings{}, opts); err != nil {
			return ierrors.WithStack(err)
		}
		value.Set(reflect.Append(value, elemValue))
	}

	if opts.validation {
		if err := ts.checkMinMaxBounds(value); err != nil {
			return ierrors.Wrapf(err, "can't serialize '%s' type", value.Kind())
		}
	}

	// check if the slice is a nil pointer to the slice type (in case the sliceLength is zero and the slice was not initialized before)
	if value.IsNil() {
		// initialize a new empty slice
		value.Set(reflect.MakeSlice(valueType, 0, 0))
	}

	if opts.validation {
		if err := api.checkArrayMustOccur(value, ts); err != nil {
			return ierrors.Wrapf(err, "can't deserialize '%s' type", value.Kind())
		}
	}

	return nil
}

func (api *API) mapDecodeMap(ctx context.Context, mapVal any, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) error {
	m, ok := mapVal.(map[string]any)
	if !ok {
		return ierrors.Errorf("non map[string]any in struct map decode, got %T instead", mapVal)
	}

	if value.IsNil() {
		value.Set(reflect.MakeMap(valueType))
	}

	var typeSettingsSet bool
	var keyTypeSettings, valueTypeSettings TypeSettings
	for k, v := range m {
		keyValue := reflect.New(valueType.Key()).Elem()
		elemValue := reflect.New(valueType.Elem()).Elem()

		if !typeSettingsSet {
			keyTypeSettings = api.typeSettingsRegistry.GetByValue(keyValue)
			valueTypeSettings = api.typeSettingsRegistry.GetByValue(elemValue)
			typeSettingsSet = true
		}

		if err := api.mapDecode(ctx, k, keyValue, keyTypeSettings, opts); err != nil {
			return ierrors.Wrapf(err, "failed to map decode map key of type %s", keyValue.Type())
		}

		if value.MapIndex(keyValue).IsValid() {
			// map entry already exists
			return ierrors.Wrapf(ErrMapValidationViolatesUniqueness, "map entry with key %v already exists", keyValue.Interface())
		}

		if err := api.mapDecode(ctx, v, elemValue, valueTypeSettings, opts); err != nil {
			return ierrors.Wrapf(err, "failed to map decode map element of type %s", elemValue.Type())
		}

		value.SetMapIndex(keyValue, elemValue)
	}

	if opts.validation {
		if err := ts.checkMinMaxBounds(value); err != nil {
			return err
		}
	}

	return nil
}
