package serix

import (
	"context"
	"math/big"
	"reflect"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/pkg/errors"
)

func (api *API) decode(ctx context.Context, b []byte, value reflect.Value, ts TypeSettings, opts *options) (int, error) {
	valueI := value.Interface()
	valueType := value.Type()
	if opts.validation {
		if err := api.callBytesValidator(valueType, b); err != nil {
			return 0, errors.Wrap(err, "pre-deserialization validation failed")
		}
	}
	var bytesRead int
	if deserializable, ok := valueI.(Deserializable); ok {
		var err error
		bytesRead, err = deserializable.Decode(b)
		if err != nil {
			return 0, errors.Wrap(err, "object failed to deserialize itself")
		}
	} else {
		var err error
		bytesRead, err = api.decodeBasedOnType(ctx, b, value, valueI, valueType, ts, opts)
		if err != nil {
			return 0, errors.WithStack(err)
		}
	}
	if opts.validation {
		if err := api.callSyntacticValidator(value, valueType); err != nil {
			return 0, errors.Wrap(err, "post-deserialization validation failed")
		}
	}

	return bytesRead, nil
}

func (api *API) decodeBasedOnType(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	globalTS, _ := api.getTypeSettings(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueType == bigIntPtrType {
			addrValue := value.Addr()
			deseri := serializer.NewDeserializer(b)
			deseri.ReadUint256(addrValue.Interface().(**big.Int), func(err error) error {
				return errors.Wrap(err, "failed to read big.Int from deserializer")
			})
			return deseri.Done()
		}
		elemType := valueType.Elem()
		if elemType.Kind() == reflect.Struct {
			if value.IsNil() {
				value.Set(reflect.New(elemType))
			}
			elemValue := value.Elem()
			return api.decodeStruct(ctx, b, elemValue, elemValue.Interface(), elemType, ts, opts)
		}

	case reflect.Struct:
		return api.decodeStruct(ctx, b, value, valueI, valueType, ts, opts)
	case reflect.Slice:
		return api.decodeSlice(ctx, b, value, valueType, ts, opts)
	case reflect.Map:
		return api.decodeMap(ctx, b, value, valueType, ts, opts)
	case reflect.Array:
		sliceValue := sliceFromArray(value)
		sliceValueType := sliceValue.Type()
		if sliceValueType.AssignableTo(bytesType) {
			deseri := serializer.NewDeserializer(b)
			deseri.ReadBytesInPlace(sliceValue.Bytes(), func(err error) error {
				return errors.Wrap(err, "failed to ready array of bytes from the deserializer")
			})
			fillArrayFromSlice(value, sliceValue)
			return deseri.Done()
		}
		return api.decodeSlice(ctx, b, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.decodeInterface(ctx, b, value, valueType, ts, opts)
	case reflect.String:
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return 0, errors.Errorf("can't deserialize 'string' type: no LengthPrefixType was provided")
		}
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		deseri.ReadString(addrValue.Interface().(*string), lengthPrefixType, func(err error) error {
			return errors.Wrap(err, "failed to read string value from the deserializer")
		})
		return deseri.Done()

	case reflect.Bool:
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		deseri.ReadBool(addrValue.Interface().(*bool), func(err error) error {
			return errors.Wrap(err, "failed to read bool value from the deserializer")
		})
		return deseri.Done()

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		deseri.ReadNum(addrValue.Interface(), func(err error) error {
			return errors.Wrap(err, "failed to read number value from the serializer")
		})
		return deseri.Done()
	default:
	}
	return 0, errors.Errorf("can't decode: unsupported type %T", valueI)
}

func (api *API) decodeInterface(
	ctx context.Context, b []byte, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) (int, error) {
	return 0, nil
}

func (api *API) decodeStruct(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	if valueType == timeType {
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		deseri.ReadTime(addrValue.Interface().(*time.Time), func(err error) error {
			return errors.Wrap(err, "failed to read time from the deserializer")
		})
		return deseri.Done()
	}
	deseri := serializer.NewDeserializer(b)
	if objectCode := ts.ObjectCode(); objectCode != nil {
		typeDen, codeNumber, err := getTypeDenotationAndNumber(objectCode)
		if err != nil {
			return 0, errors.WithStack(err)
		}
		deseri.CheckTypePrefix(codeNumber, typeDen, func(err error) error {
			return errors.Wrap(err, "failed to check object type")
		})
	}
	if err := api.decodeStructFields(ctx, deseri, value, valueType, opts); err != nil {
		return 0, errors.WithStack(err)
	}
	return deseri.Done()
}

func (api *API) decodeStructFields(
	ctx context.Context, deseri *serializer.Deserializer, value reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := parseStructType(valueType)
	if err != nil {
		return errors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
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
			if err := api.decodeStructFields(ctx, deseri, fieldValue, fieldType, opts); err != nil {
				return errors.Wrapf(err, "can't deserialize embedded struct %s", sField.name)
			}
			continue
		}
		var bytesRead int
		if sField.settings.isOptional {
			payloadLength, err := deseri.ReadPayloadLength()
			if err != nil {
				return errors.Wrap(err, "can't read payload length from the deserializer")
			}
			if payloadLength == 0 {
				continue
			}

			bytesRead, err = api.decode(ctx, deseri.RemainingBytes(), fieldValue, sField.settings.ts, opts)
			if err != nil {
				return errors.Wrapf(err, "failed to deserialize optional struct field %s", sField.name)
			}
			if bytesRead != int(payloadLength) {
				return errors.Wrapf(
					err,
					"optional object length isn't equal to the amount of bytes read; length=%d, bytesRead=%d",
					payloadLength, bytesRead,
				)
			}
		} else {
			bytesRead, err = api.decode(ctx, deseri.RemainingBytes(), fieldValue, sField.settings.ts, opts)
			if err != nil {
				return errors.Wrapf(err, "failed to deserialize struct field %s", sField.name)
			}
		}
		deseri.Skip(bytesRead, func(err error) error {
			return errors.Wrap(err, "failed to skip amount of bytes read for the struct field")
		})
	}
	return nil
}

func (api *API) decodeSlice(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	if valueType.AssignableTo(bytesType) {
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return 0, errors.Errorf("no LengthPrefixType was provided for slice type %s", valueType)
		}
		deseri := serializer.NewDeserializer(b)
		addrValue := value.Addr()
		deseri.ReadVariableByteSlice(addrValue.Interface().(*[]byte), lengthPrefixType, func(err error) error {
			return errors.Wrap(err, "failed to read bytes from the deserializer")
		})
		return deseri.Done()
	}
	deseri := serializer.NewDeserializer(b)
	deserializeItem := func(b []byte) (bytesRead int, ty uint32, err error) {
		elemValue := reflect.New(valueType.Elem()).Elem()
		bytesRead, err = api.decode(ctx, b, elemValue, TypeSettings{}, opts)
		if err != nil {
			return 0, 0, errors.WithStack(err)
		}
		value.Set(reflect.Append(value, elemValue))
		return bytesRead, 0, nil
	}
	lengthPrefixType, set := ts.LengthPrefixType()
	if !set {
		return 0, errors.Errorf("no LengthPrefixType was provided for type %s", valueType)
	}
	arrayRules := ts.ArrayRules()
	if arrayRules == nil {
		arrayRules = new(serializer.ArrayRules)
	}
	serializationMode := ts.toMode(opts)
	deseri.ReadSequenceOfObjects(deserializeItem, serializationMode, lengthPrefixType, arrayRules, func(err error) error {
		return errors.Wrapf(err, "failed to read slice of objects %s from the deserialized", valueType)
	})
	return deseri.Done()
}

func (api *API) decodeMap(ctx context.Context, b []byte, value reflect.Value,
	valueType reflect.Type, ts TypeSettings, opts *options) (int, error) {
	return 0, nil
}

func (api *API) decodeMapKVPair(ctx context.Context, b []byte, key, val reflect.Value, opts *options) (int, error) {
	keyBytesRead, err := api.decode(ctx, b, key, TypeSettings{}, opts)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to decode map key of type %s", key.Type())
	}
	b = b[keyBytesRead:]
	elemBytesRead, err := api.decode(ctx, b, val, TypeSettings{}, opts)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to decode map element of type %s", val.Type())
	}
	return keyBytesRead + elemBytesRead, nil
}
