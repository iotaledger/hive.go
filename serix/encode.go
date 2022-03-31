package serix

import (
	"context"
	"math/big"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/datastructure/orderedmap"

	"github.com/iotaledger/hive.go/serializer/v2"
)

func (api *API) encode(ctx context.Context, value reflect.Value, ts TypeSettings, opts *options) ([]byte, error) {
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
		bytes, err = api.encodeBasedOnType(ctx, value, valueI, ts, opts)
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

func (api *API) encodeBasedOnType(
	ctx context.Context, value reflect.Value, valueI interface{}, ts TypeSettings, opts *options,
) ([]byte, error) {
	valueType := value.Type()
	globalTS := api.getTypeSettings(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueBigInt, ok := valueI.(*big.Int); ok {
			seri := serializer.NewSerializer()
			return seri.WriteUint256(valueBigInt, func(err error) error {
				return errors.Wrap(err, "failed to write math big int to serializer")
			}).Serialize()
		}
		if valueOrderedMap, ok := valueI.(*orderedmap.OrderedMap); ok {
			return api.encodeOrderedMap(ctx, valueOrderedMap, ts, opts)
		}
		if value.IsNil() {
			return nil, errors.Errorf("unexpected nil pointer for type %T", valueI)
		}
		return api.encode(ctx, value.Elem(), ts, opts)

	case reflect.Struct:
		if valueTime, ok := valueI.(time.Time); ok {
			seri := serializer.NewSerializer()
			return seri.WriteTime(valueTime, func(err error) error {
				return errors.Wrap(err, "failed to write time to serializer")
			}).Serialize()
		}
		return api.encodeStruct(ctx, value, valueType, ts, opts)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, valueType, ts, opts)
	case reflect.Map:
		return api.encodeMap(ctx, value, valueType, ts, opts)
	case reflect.Array:
		sliceValue := sliceFromArray(value)
		sliceValueType := sliceValue.Type()
		if sliceValueType.AssignableTo(bytesType) {
			seri := serializer.NewSerializer()
			return seri.WriteBytes(value.Bytes(), func(err error) error {
				return errors.Wrap(err, "failed to write array of bytes to serializer")
			}).Serialize()
		}
		return api.encodeSlice(ctx, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.encodeInterface(ctx, value, valueType, ts, opts)
	case reflect.String:
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return nil, errors.Errorf("in order to serialize 'string' type no LengthPrefixType was provided")
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
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		seri := serializer.NewSerializer()
		return seri.WriteNum(valueI, func(err error) error {
			return errors.Wrap(err, "failed to write number value to serializer")
		}).Serialize()
	default:
		return nil, errors.Errorf("can't encode type %T, unsupported kind %s", valueI, value.Kind())
	}
}

func (api *API) encodeInterface(
	ctx context.Context, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) ([]byte, error) {
	elemValue := value.Elem()
	if !elemValue.IsValid() {
		return nil, errors.Errorf("can't serialize interface %s it must have underlying value", valueType)
	}
	registry := api.getInterfaceObjects(valueType)
	if registry == nil {
		return nil, errors.Errorf("interface %s isn't registered", valueType)
	}
	elemType := elemValue.Type()
	if _, exists := registry.fromTypeToCode[elemType]; !exists {
		return nil, errors.Errorf("underlying type %s hasn't been registered for interface type %s",
			elemType, valueType)
	}
	encodedBytes, err := api.encode(ctx, elemValue, ts, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode interface element %s", elemType)
	}
	return encodedBytes, nil
}

func (api *API) encodeStruct(
	ctx context.Context, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) ([]byte, error) {
	s := serializer.NewSerializer()
	if objectCode := ts.ObjectCode(); objectCode != nil {
		s.WriteNum(objectCode, func(err error) error {
			return errors.Wrap(err, "failed to write object type code into serializer")
		})
	}
	if err := api.encodeStructFields(ctx, s, value, valueType, opts); err != nil {
		return nil, errors.WithStack(err)
	}
	return s.Serialize()
}

func (api *API) encodeStructFields(
	ctx context.Context, s *serializer.Serializer, value reflect.Value, valueType reflect.Type, opts *options,
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
			if err := api.encodeStructFields(ctx, s, fieldValue, sField.fType, opts); err != nil {
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
			payloadBytes, err := api.encode(ctx, fieldValue, sField.settings.ts, opts)
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
			b, err := api.encode(ctx, fieldValue, sField.settings.ts, opts)
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

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, valueType reflect.Type,
	ts TypeSettings, opts *options) ([]byte, error) {
	lengthPrefixType, set := ts.LengthPrefixType()
	if !set {
		return nil, errors.Errorf("no LengthPrefixType was provided for slice type %s", valueType)
	}

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
		elemBytes, err := api.encode(ctx, elemValue, TypeSettings{}, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
		}
		data[i] = elemBytes
	}
	arrayRules := ts.ArrayRules()
	if arrayRules == nil {
		arrayRules = new(serializer.ArrayRules)
	}
	seri := serializer.NewSerializer()
	serializationMode := opts.toMode()
	lexicalOrdering, set := ts.LexicalOrdering()
	if set && lexicalOrdering {
		serializationMode |= serializer.DeSeriModePerformLexicalOrdering
	}
	return seri.WriteSliceOfByteSlices(data, serializationMode, lengthPrefixType, arrayRules, func(err error) error {
		return errors.Wrapf(err,
			"serializer failed to write slice of objects %s as slice of byte slices", valueType,
		)
	}).Serialize()
}

func (api *API) encodeMap(ctx context.Context, value reflect.Value, valueType reflect.Type,
	ts TypeSettings, opts *options) ([]byte, error) {
	size := value.Len()
	keys := value.MapKeys()

	slice := make([]*keyValuePair, size)
	for i, key := range keys {
		mapValue := value.MapIndex(key)
		slice[i] = &keyValuePair{
			Key:   key.Interface(),
			Value: mapValue.Interface(),
		}
	}
	valueSlice := reflect.ValueOf(slice)
	ts = ts.WithLexicalOrdering(true)
	arrayRules := ts.ArrayRules()
	if arrayRules == nil {
		arrayRules = new(serializer.ArrayRules)
	}
	arrayRules.ValidationMode |= serializer.ArrayValidationModeLexicalOrdering
	ts = ts.WithArrayRules(arrayRules)
	b, err := api.encodeSlice(ctx, valueSlice, valueSlice.Type(), ts, opts)
	return b, errors.Wrapf(err, "failed to encode slice built from map %s", valueType)
}

func (api *API) encodeOrderedMap(ctx context.Context, om *orderedmap.OrderedMap, ts TypeSettings, opts *options) ([]byte, error) {
	size := om.Size()

	slice := make([]*keyValuePair, size)
	var i int
	om.ForEach(func(key, value interface{}) bool {
		slice[i] = &keyValuePair{
			Key:   key,
			Value: value,
		}
		i++
		return true
	})
	valueSlice := reflect.ValueOf(slice)
	b, err := api.encodeSlice(ctx, valueSlice, valueSlice.Type(), ts, opts)
	return b, errors.Wrap(err, "failed to encode slice built from an OrderedMap")
}
