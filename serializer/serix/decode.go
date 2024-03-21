package serix

import (
	"context"
	"math/big"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

func (api *API) decode(ctx context.Context, b []byte, value reflect.Value, ts TypeSettings, opts *options) (int, error) {
	valueType := value.Type()
	var deserializable Deserializable
	var bytesRead int

	if _, ok := value.Interface().(Deserializable); ok {
		if value.Kind() == reflect.Ptr && value.IsNil() {
			value.Set(reflect.New(valueType.Elem()))
		}
		//nolint:forcetypeassert // false positive
		deserializable = value.Interface().(Deserializable)
	} else if value.CanAddr() {
		if addrDeserializable, ok := value.Addr().Interface().(Deserializable); ok {
			deserializable = addrDeserializable
		}
	}

	// Check for a pointer to a Deserializable interface
	if deserializable == nil && value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Ptr {
		if _, ok := value.Elem().Interface().(Deserializable); ok {
			valueElem := value.Elem()
			if valueElem.IsNil() {
				valueElem.Set(reflect.New(valueElem.Type().Elem()))
			}

			return api.decode(ctx, b, valueElem, ts, opts)
		}
	}

	if deserializable != nil {
		typeSettingValue := value
		if valueType.Kind() == reflect.Ptr {
			typeSettingValue = value.Elem()
		}
		globalTS, _ := api.typeSettingsRegistry.GetByType(typeSettingValue.Type())
		ts = ts.merge(globalTS)
		if objectType := ts.ObjectType(); objectType != nil {
			typeDen, objectCode, err := getTypeDenotationAndCode(objectType)
			if err != nil {
				return 0, ierrors.WithStack(err)
			}

			deseri := serializer.NewDeserializer(b)
			deseri.CheckTypePrefix(objectCode, typeDen, func(err error) error {
				return ierrors.Wrap(err, "failed to check object type")
			})
			b = deseri.RemainingBytes()
			prefixBytesRead, err := deseri.Done()
			if err != nil {
				return 0, ierrors.WithStack(err)
			}
			bytesRead += prefixBytesRead
		}

		var err error
		bytesDecoded, err := deserializable.Decode(b)
		bytesRead += bytesDecoded
		if err != nil {
			return 0, ierrors.Wrap(err, "object failed to deserialize itself")
		}

		if contextAwareDeserializable, ok := deserializable.(ContextAwareDeserializable); ok {
			contextAwareDeserializable.SetDeserializationContext(ctx)
		}
	} else {
		var err error
		bytesRead, err = api.decodeBasedOnType(ctx, b, value, valueType, ts, opts)
		if err != nil {
			return 0, ierrors.WithStack(err)
		}
	}

	if opts.validation {
		if err := api.callSyntacticValidator(ctx, value, valueType); err != nil {
			return 0, ierrors.Errorf("post-deserialization validation failed: %w", err)
		}
	}

	return bytesRead, nil
}

func (api *API) decodeBasedOnType(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	globalTS, _ := api.typeSettingsRegistry.GetByType(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueType == bigIntPtrType {
			addrValue := value.Addr()
			deseri := serializer.NewDeserializer(b)
			//nolint:forcetypeassert // false positive, we already checked the type via reflect
			deseri.ReadUint256(addrValue.Interface().(**big.Int), func(err error) error {
				return ierrors.Wrap(err, "failed to read big.Int from deserializer")
			})

			return deseri.Done()
		}
		elemType := valueType.Elem()

		if value.IsNil() {
			value.Set(reflect.New(elemType))
		}
		elemValue := value.Elem()

		switch elemType.Kind() {
		case reflect.Struct:
			if contextAwareDeserializable, ok := value.Interface().(ContextAwareDeserializable); ok {
				contextAwareDeserializable.SetDeserializationContext(ctx)
			}

			return api.decodeStruct(ctx, b, elemValue, elemType, ts, opts)
		case reflect.Slice:
			return api.decodeSlice(ctx, b, elemValue, elemType, ts, opts)
		case reflect.Ptr:
			return api.decodeBasedOnType(ctx, b, elemValue, elemType, ts, opts)
		case reflect.Interface:
			return api.decodeInterface(ctx, b, elemValue, elemType, ts, opts)
		case reflect.Map:
			return api.decodeMap(ctx, b, elemValue, elemType, ts, opts)
		case reflect.Array:
			return api.decodeArray(ctx, b, elemValue, ts, opts)
		default:
			return api.decodeBasedOnType(ctx, b, elemValue, elemType, ts, opts)
		}

	case reflect.Struct:
		if contextAwareDeserializable, ok := value.Interface().(ContextAwareDeserializable); ok {
			contextAwareDeserializable.SetDeserializationContext(ctx)
		}

		return api.decodeStruct(ctx, b, value, valueType, ts, opts)
	case reflect.Slice:
		return api.decodeSlice(ctx, b, value, valueType, ts, opts)
	case reflect.Map:
		return api.decodeMap(ctx, b, value, valueType, ts, opts)
	case reflect.Array:
		return api.decodeArray(ctx, b, value, ts, opts)
	case reflect.Interface:
		return api.decodeInterface(ctx, b, value, valueType, ts, opts)
	case reflect.String:
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return 0, ierrors.New("can't deserialize 'string' type: no LengthPrefixType was provided")
		}
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		addrValue = addrValue.Convert(reflect.TypeOf((*string)(nil)))

		var minLen, maxLen int
		if opts.validation {
			minLen, maxLen = ts.MinMaxLen()
		}

		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		deseri.ReadString(
			addrValue.Interface().(*string),
			serializer.SeriLengthPrefixType(lengthPrefixType),
			func(err error) error {
				err = ierrors.Wrap(err, "failed to read string value from the deserializer")

				switch {
				case ierrors.Is(err, serializer.ErrDeserializationLengthMinNotReached):
					return ierrors.Join(err, serializer.ErrArrayValidationMinElementsNotReached)
				case ierrors.Is(err, serializer.ErrDeserializationLengthMaxExceeded):
					return ierrors.Join(err, serializer.ErrArrayValidationMaxElementsExceeded)
				default:
					return err
				}
			}, minLen, maxLen)

		if opts.validation {
			// check the string for UTF-8 validity
			if !utf8.ValidString(value.String()) {
				return 0, ierrors.Errorf("can't deserialize 'string' type: %w", ErrNonUTF8String)
			}
		}

		return deseri.Done()

	case reflect.Bool:
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()

		addrValue = addrValue.Convert(reflect.TypeOf((*bool)(nil)))
		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		deseri.ReadBool(addrValue.Interface().(*bool), func(err error) error {
			return ierrors.Wrap(err, "failed to read bool value from the deserializer")
		})

		return deseri.Done()

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		_, _, addrTypeToConvert := getNumberTypeToConvert(valueType.Kind())
		addrValue = addrValue.Convert(addrTypeToConvert)
		deseri.ReadNum(addrValue.Interface(), func(err error) error {
			return ierrors.Wrap(err, "failed to read number value from the serializer")
		})

		return deseri.Done()
	default:
	}

	return 0, ierrors.Errorf("can't decode: unsupported type %s", valueType)
}

func (api *API) decodeInterface(
	ctx context.Context, b []byte, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) (int, error) {
	iObjects := api.getInterfaceObjects(valueType)
	if iObjects == nil {
		return 0, ierrors.Errorf("interface %s hasn't been registered", valueType)
	}
	deseri := serializer.NewDeserializer(b)
	objectCode, err := deseri.GetObjectType(iObjects.TypeDenotation())
	if err != nil {
		return 0, ierrors.WithStack(err)
	}

	objectType, exists := iObjects.GetObjectTypeByCode(objectCode)
	if !exists || objectType == nil {
		return 0, ierrors.Wrapf(ErrInterfaceUnderlyingTypeNotRegistered, "object code: %d, interface: %s", objectCode, valueType)
	}

	objectValue := reflect.New(objectType).Elem()
	bytesRead, err := api.decode(ctx, b, objectValue, ts, opts)
	if err != nil {
		return 0, ierrors.WithStack(err)
	}
	value.Set(objectValue)

	return bytesRead, nil
}

func (api *API) decodeStruct(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	if valueType == timeType {
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()

		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		deseri.ReadTime(addrValue.Interface().(*time.Time), func(err error) error {
			return ierrors.Wrap(err, "failed to read time from the deserializer")
		})

		return deseri.Done()
	}
	deseri := serializer.NewDeserializer(b)
	if objectType := ts.ObjectType(); objectType != nil {
		typeDen, objectCode, err := getTypeDenotationAndCode(objectType)
		if err != nil {
			return 0, ierrors.WithStack(err)
		}
		deseri.CheckTypePrefix(objectCode, typeDen, func(err error) error {
			return ierrors.Wrap(err, "failed to check object type")
		})
	}
	if err := api.decodeStructFields(ctx, deseri, value, valueType, opts); err != nil {
		return 0, ierrors.WithStack(err)
	}

	return deseri.Done()
}

func (api *API) decodeStructFields(
	ctx context.Context, deseri *serializer.Deserializer, value reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := api.getStructFields(valueType)
	if err != nil {
		return ierrors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
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
			if err := api.decodeStructFields(ctx, deseri, fieldValue, fieldType, opts); err != nil {
				return ierrors.Wrapf(err, "can't deserialize embedded struct %s", sField.name)
			}

			continue
		}
		var bytesRead int
		if sField.settings.isOptional {
			payloadLength, err := deseri.ReadPayloadLength()
			if err != nil {
				return ierrors.Wrap(err, "can't read payload length from the deserializer")
			}
			if payloadLength == 0 {
				continue
			}

			bytesRead, err = api.decode(ctx, deseri.RemainingBytes(), fieldValue, sField.settings.ts, opts)
			if err != nil {
				return ierrors.Wrapf(err, "failed to deserialize optional struct field %s", sField.name)
			}
			if bytesRead != int(payloadLength) {
				return ierrors.Wrapf(
					err,
					"optional object length isn't equal to the amount of bytes read; length=%d, bytesRead=%d",
					payloadLength, bytesRead,
				)
			}
		} else {
			bytesRead, err = api.decode(ctx, deseri.RemainingBytes(), fieldValue, sField.settings.ts, opts)
			if err != nil {
				return ierrors.Wrapf(err, "failed to deserialize struct field %s", sField.name)
			}
		}
		deseri.Skip(bytesRead, func(err error) error {
			return ierrors.Wrap(err, "failed to skip amount of bytes read for the struct field")
		})
	}

	return nil
}

func (api *API) decodeArray(ctx context.Context, b []byte, value reflect.Value, ts TypeSettings, opts *options) (int, error) {
	sliceValue := sliceFromArray(value)
	sliceValueType := sliceValue.Type()

	// check if it is an array of bytes
	if sliceValueType.AssignableTo(bytesType) {
		deseri := serializer.NewDeserializer(b)
		if objectType := ts.ObjectType(); objectType != nil {
			typeDen, objectCode, err := getTypeDenotationAndCode(objectType)
			if err != nil {
				return 0, ierrors.WithStack(err)
			}
			deseri.CheckTypePrefix(objectCode, typeDen, func(err error) error {
				return ierrors.Wrap(err, "failed to check object type")
			})
		}
		deseri.ReadBytesInPlace(sliceValue.Bytes(), func(err error) error {
			return ierrors.Wrap(err, "failed to read array of bytes from the deserializer")
		})
		fillArrayFromSlice(value, sliceValue)

		return deseri.Done()
	}

	// if it is an array of objects, handle the array like a slice
	return api.decodeSlice(ctx, b, sliceValue, sliceValueType, ts, opts)
}

func (api *API) decodeSlice(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	if valueType.AssignableTo(bytesType) {
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return 0, ierrors.Errorf("no LengthPrefixType was provided for slice type %s", valueType)
		}
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		addrValue = addrValue.Convert(reflect.TypeOf((*[]byte)(nil)))
		minLen, maxLen := ts.MinMaxLen()

		//nolint:forcetypeassert // false positive, we already checked the type via reflect
		deseri.ReadVariableByteSlice(
			addrValue.Interface().(*[]byte),
			serializer.SeriLengthPrefixType(lengthPrefixType),
			func(err error) error {
				err = ierrors.Wrap(err, "failed to read bytes from the deserializer")

				switch {
				case ierrors.Is(err, serializer.ErrDeserializationLengthMinNotReached):
					return ierrors.Join(err, serializer.ErrArrayValidationMinElementsNotReached)
				case ierrors.Is(err, serializer.ErrDeserializationLengthMaxExceeded):
					return ierrors.Join(err, serializer.ErrArrayValidationMaxElementsExceeded)
				default:
					return err
				}
			}, minLen, maxLen)

		return deseri.Done()
	}
	deserializeItem := func(b []byte) (bytesRead int, err error) {
		elemValue := reflect.New(valueType.Elem()).Elem()
		bytesRead, err = api.decode(ctx, b, elemValue, TypeSettings{}, opts)
		if err != nil {
			return 0, ierrors.WithStack(err)
		}
		value.Set(reflect.Append(value, elemValue))

		return bytesRead, nil
	}

	bytesRead, err := api.decodeSequence(b, deserializeItem, valueType, ts, opts)
	if err != nil {
		return bytesRead, err
	}

	// check if the slice is a nil pointer to the slice type (in case the sliceLength is zero and the slice was not initialized before)
	if value.IsNil() {
		// initialize a new empty slice
		value.Set(reflect.MakeSlice(valueType, 0, 0))
	}

	if opts.validation {
		if err := api.checkArrayMustOccur(value, ts); err != nil {
			return bytesRead, ierrors.Wrapf(err, "can't deserialize '%s' type", value.Kind())
		}
	}

	return bytesRead, nil
}

func (api *API) decodeMapKVPair(ctx context.Context, b []byte, key, val reflect.Value, opts *options) (int, error) {
	keyTypeSettings := api.typeSettingsRegistry.GetByValue(key)
	valueTypeSettings := api.typeSettingsRegistry.GetByValue(val)

	keyBytesRead, err := api.decode(ctx, b, key, keyTypeSettings, opts)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to decode map key of type %s", key.Type())
	}
	b = b[keyBytesRead:]
	elemBytesRead, err := api.decode(ctx, b, val, valueTypeSettings, opts)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to decode map element of type %s", val.Type())
	}

	return keyBytesRead + elemBytesRead, nil
}

func (api *API) decodeMap(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	if value.IsNil() {
		value.Set(reflect.MakeMap(valueType))
	}

	deserializeItem := func(b []byte) (bytesRead int, err error) {
		keyValue := reflect.New(valueType.Key()).Elem()
		elemValue := reflect.New(valueType.Elem()).Elem()
		bytesRead, err = api.decodeMapKVPair(ctx, b, keyValue, elemValue, opts)
		if err != nil {
			return 0, ierrors.WithStack(err)
		}

		if value.MapIndex(keyValue).IsValid() {
			// map entry already exists
			return 0, ierrors.Wrapf(ErrMapValidationViolatesUniqueness, "map entry with key %v already exists", keyValue.Interface())
		}

		value.SetMapIndex(keyValue, elemValue)

		return bytesRead, nil
	}
	ts = ts.ensureOrdering()

	consumedBytes, err := api.decodeSequence(b, deserializeItem, valueType, ts, opts)
	if err != nil {
		return consumedBytes, err
	}

	if opts.validation {
		if err := ts.checkMinMaxBounds(value); err != nil {
			return consumedBytes, err
		}
	}

	return consumedBytes, nil
}

func (api *API) decodeSequence(b []byte, deserializeItem serializer.DeserializeFunc, valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	lengthPrefixType, set := ts.LengthPrefixType()
	if !set {
		return 0, ierrors.Errorf("no LengthPrefixType was provided for type %s", valueType)
	}
	arrayRules := ts.ArrayRules()
	if arrayRules == nil {
		arrayRules = new(ArrayRules)
	}
	serializationMode := ts.toMode(opts)
	deseri := serializer.NewDeserializer(b)
	serializerArrayRules := serializer.ArrayRules(*arrayRules)
	serializerArrayRulesPtr := &serializerArrayRules
	deseri.ReadSequenceOfObjects(
		deserializeItem, serializationMode,
		serializer.SeriLengthPrefixType(lengthPrefixType),
		serializerArrayRulesPtr,
		func(err error) error {
			return ierrors.Wrapf(err, "failed to read sequence of objects %s from the deserialized", valueType)
		})

	return deseri.Done()
}
