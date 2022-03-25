package serix

import (
	"context"
	"math/big"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/serializer/v2"
)

func (api *API) encode(ctx context.Context, value reflect.Value, opts *options) ([]byte, error) {
	valueI := value.Interface()
	if opts.validation {
		if validator, ok := valueI.(SyntacticValidator); ok {
			if err := validator.Validate(); err != nil {
				return nil, errors.Wrap(err, "pre-serialization syntactic validation failed")
			}
		}
	}
	var bytes []byte
	if serializable, ok := valueI.(Serializable); ok {
		var err error
		bytes, err = serializable.Encode()
		if err != nil {
			return nil, errors.Wrap(err, "object failed to serialize itself")
		}
	} else {
		var err error
		bytes, err = api.encodeBasedOnType(ctx, value, valueI, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if opts.validation {
		if bytesValidator, ok := valueI.(BytesValidator); ok {
			if err := bytesValidator.ValidateBytes(bytes); err != nil {
				return nil, errors.Wrap(err, "post-serialization bytes validation failed")
			}
		}
	}

	return bytes, nil
}

func (api *API) encodeBasedOnType(ctx context.Context, value reflect.Value, valueI interface{}, opts *options) ([]byte, error) {
	switch value.Kind() {
	case reflect.Ptr:
		seri := serializer.NewSerializer()
		if valueBigInt, ok := valueI.(*big.Int); ok {
			return seri.WriteUint256(valueBigInt, func(err error) error {
				return errors.Wrap(err, "failed to write math big int to serializer")
			}).Serialize()
		}
		if value.IsNil() {
			return nil, errors.Errorf("unexpected nil pointer for type %T", valueI)
		}
		api.writeObjectCode(valueI, seri)
		valueBytes, err := api.encode(ctx, value.Elem(), opts)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize pointer value")
		}
		seri.WriteBytes(valueBytes, func(err error) error {
			return errors.Wrapf(err, "failed to write serialized pointer value bytes to serializer")
		})
		return seri.Serialize()

	case reflect.Struct:
		if valueTime, ok := valueI.(time.Time); ok {
			seri := serializer.NewSerializer()
			return seri.WriteTime(valueTime, func(err error) error {
				return errors.Wrap(err, "failed to write time to serializer")
			}).Serialize()
		}
		return api.encodeStruct(ctx, value, valueI, opts)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, valueI, value.Type(), opts)
	case reflect.Map:
		return api.encodeMap(ctx, value, valueI, value.Type(), opts)
	case reflect.Array:
		value = sliceFromArray(value)
		valueType := value.Type()
		if valueType.AssignableTo(bytesType) {
			seri := serializer.NewSerializer()
			return seri.WriteBytes(value.Bytes(), func(err error) error {
				return errors.Wrap(err, "failed to write array of bytes to serializer")
			}).Serialize()
		}
		return api.encodeSlice(ctx, value, valueI, valueType, opts)
	case reflect.Interface:
		return api.encodeInterface(ctx, value, opts)
	case reflect.String:
		lengthPrefixType, err := determineLengthPrefixType(valueI, opts, value.Type())
		if err != nil {
			return nil, err
		}
		seri := serializer.NewSerializer()
		return seri.WriteString(value.String(), lengthPrefixType, func(err error) error {
			return errors.Wrap(err, "failed to write string value to serializer")
		}).Serialize()

	case reflect.Bool:
		seri := serializer.NewSerializer()
		return seri.WriteBool(value.Bool(), func(err error) error {
			return errors.Wrap(err, "failed to write bool value to serializer")
		}).Serialize()

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		seri := serializer.NewSerializer()
		return seri.WriteNum(valueI, func(err error) error {
			return errors.Wrap(err, "failed to write number value to serializer")
		}).Serialize()
	default:
		return nil, errors.Errorf("can't encode type %T, unsupported kind %s", valueI, value.Kind())
	}
}

func (api *API) encodeInterface(ctx context.Context, value reflect.Value, opts *options) ([]byte, error) {
	valueType := value.Type()
	elemValue := value.Elem()
	if !elemValue.IsValid() {
		return nil, errors.Errorf("can't serialize interface %s it must have underlying value", valueType)
	}
	registry := api.getInterfaceRegistry(valueType)
	if registry == nil {
		return nil, errors.Errorf("interface %s isn't registered", valueType)
	}
	elemType := elemValue.Type()
	if _, exists := registry.fromTypeToCode[elemType]; !exists {
		return nil, errors.Errorf("underlying type %s hasn't been registered for interface type %s",
			elemType, valueType)
	}
	encodedBytes, err := api.encode(ctx, elemValue, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode interface element %s", elemType)
	}
	return encodedBytes, nil
}

func (api *API) encodeStruct(ctx context.Context, value reflect.Value, valueI interface{}, opts *options) ([]byte, error) {
	s := serializer.NewSerializer()
	api.writeObjectCode(valueI, s)
	if err := api.encodeStructFields(ctx, s, value, opts); err != nil {
		return nil, errors.WithStack(err)
	}
	return s.Serialize()
}

func (api *API) encodeStructFields(ctx context.Context, s *serializer.Serializer, value reflect.Value, opts *options) error {
	valueType := value.Type()
	structFields, err := parseStructType(valueType)
	if err != nil {
		return errors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
		if sField.isEmbeddedStruct {
			if err := api.encodeStructFields(ctx, s, fieldValue, opts); err != nil {
				return errors.Wrapf(err, "can't serialize embedded struct %s", sField.name)
			}
			continue
		}
		var fieldBytes []byte
		if sField.settings.isPayload {
			if fieldValue.IsNil() {
				s.WritePayloadLength(0, func(err error) error {
					return errors.Wrapf(err,
						"failed to write zero payload length for struct field %s to serializer",
						sField.name,
					)
				})
				continue
			}
			payloadBytes, err := api.encode(ctx, fieldValue, opts)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize payload struct field %s", sField.name)
			}
			s.WritePayloadLength(len(payloadBytes), func(err error) error {
				return errors.Wrapf(err,
					"failed to write payload length for struct field %s to serializer",
					sField.name,
				)
			})
			fieldBytes = payloadBytes
		} else {
			newOpts := &options{}
			if sField.settings.isLengthPrefixType {
				newOpts.lengthPrefixType = sField.settings.lengthPrefixType
				newOpts.isLengthPrefixType = true
			}

			b, err := api.encode(ctx, fieldValue, newOpts)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize struct field %s", sField.name)
			}
			fieldBytes = b
		}
		s.WriteBytes(fieldBytes, func(err error) error {
			return errors.Wrapf(err,
				"failed to write serialized struct field bytes to serializer, field=%s",
				sField.name,
			)
		})
	}
	return nil

}

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type,
	opts *options) ([]byte, error) {

	lengthPrefixType, err := determineLengthPrefixType(valueI, opts, valueType)
	if err != nil {
		return nil, err
	}
	serializationMode := opts.toMode()

	if valueType.AssignableTo(bytesType) {
		seri := serializer.NewSerializer()
		seri.WriteVariableByteSlice(value.Bytes(), lengthPrefixType, func(err error) error {
			return errors.Wrap(err, "failed to write bytes to serializer")
		})
		return seri.Serialize()
	}
	sliceLen := value.Len()
	data := make([][]byte, sliceLen)
	for i := 0; i < sliceLen; i++ {
		elemValue := value.Index(i)
		elemBytes, err := api.encode(ctx, elemValue, new(options))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
		}
		data[i] = elemBytes
	}
	arrayRules := &serializer.ArrayRules{}
	if ruler, ok := valueI.(ArrayRulesProvider); ok {
		arrayRules = ruler.ArrayRules()
	}
	seri := serializer.NewSerializer()
	return seri.WriteSliceOfByteSlices(data, serializationMode, lengthPrefixType, arrayRules, func(err error) error {
		return errors.Wrapf(err,
			"serializer failed to write slice of objects %s as slice of byte slices", valueType,
		)
	}).Serialize()
}

func (api *API) encodeMap(ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type,
	opts *options) ([]byte, error) {

	lengthPrefixType, err := determineLengthPrefixType(valueI, opts, valueType)
	if err != nil {
		return nil, err
	}
	serializationMode := opts.toMode()

	mapLen := value.Len()
	mapKeys := value.MapKeys()

	// each entry in data slice contains concatenated bytes of key + value
	data := make([][]byte, mapLen)
	for i, mapK := range mapKeys {
		mapV := value.MapIndex(mapK)

		mapKBytes, err := api.encode(ctx, mapK, new(options))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode map key %s of map %s", mapK.String(), valueType)
		}

		mapVBytes, err := api.encode(ctx, mapV, new(options))

		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode map value with key %s of map %s", mapK.String(), valueType)
		}
		data[i] = byteutils.ConcatBytes(mapKBytes, mapVBytes)
	}
	arrayRules := &serializer.ArrayRules{}
	if ruler, ok := valueI.(ArrayRulesProvider); ok {
		arrayRules = ruler.ArrayRules()
	}
	seri := serializer.NewSerializer()
	return seri.WriteSliceOfByteSlices(data, serializationMode, lengthPrefixType, arrayRules, func(err error) error {
		return errors.Wrapf(err,
			"serializer failed to write map %s as slice of byte slices", valueType,
		)
	}).Serialize()
}

func determineLengthPrefixType(valueI interface{}, opts *options, valueType reflect.Type) (serializer.SeriLengthPrefixType, error) {
	lptp, ok := valueI.(LengthPrefixTypeProvider)
	if !ok && !opts.isLengthPrefixType {
		return 0, errors.Errorf("type %s must implement LengthPrefixTypeProvider interface or have a 'lengthPrefixType' struct tag", valueType)
	}

	var lengthPrefixType serializer.SeriLengthPrefixType
	if ok {
		lengthPrefixType = lptp.LengthPrefixType()
	}
	if opts.isLengthPrefixType {
		lengthPrefixType = opts.lengthPrefixType
	}
	return lengthPrefixType, nil
}

func (api *API) writeObjectCode(valueI interface{}, s *serializer.Serializer) {
	if codeProvider, ok := valueI.(ObjectCodeProvider); ok {
		objectCode := codeProvider.ObjectCode()
		s.WriteNum(objectCode, func(err error) error {
			return errors.Wrap(err, "failed to write object type code into serializer")
		})
	}
}
