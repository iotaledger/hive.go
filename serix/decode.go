package serix

import (
	"context"
	"reflect"

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
	//globalTS, _ := api.getTypeSettings(valueType)
	//ts = ts.merge(globalTS)
	//switch value.Kind() {
	//case reflect.Ptr:
	//	if valueType == bigIntPtrType {
	//		addrValue := value.Addr()
	//		deseri := serializer.NewDeserializer(b)
	//		deseri.ReadUint256(addrValue.Interface().(**big.Int), func(err error) error {
	//			return errors.Wrap(err, "failed to read big.Int from deserializer")
	//		})
	//		return deseri.Done()
	//	}
	//	if omm, ok := parseOrderedMapMeta(valueType); ok {
	//		return api.encodeOrderedMap(ctx, value, valueType, omm, ts, opts)
	//	}
	//	elemValue := reflect.Indirect(value)
	//	if !elemValue.IsValid() {
	//		return nil, errors.Errorf("unexpected nil pointer for type %T", valueI)
	//	}
	//	if elemValue.Kind() == reflect.Struct {
	//		return api.encodeStruct(ctx, elemValue, elemValue.Interface(), elemValue.Type(), ts, opts)
	//	}
	//
	//case reflect.Struct:
	//	return api.encodeStruct(ctx, value, valueI, valueType, ts, opts)
	//case reflect.Slice:
	//	return api.encodeSlice(ctx, value, valueType, ts, opts)
	//case reflect.Map:
	//	return api.encodeMap(ctx, value, valueType, ts, opts)
	//case reflect.Array:
	//	sliceValue := sliceFromArray(value)
	//	sliceValueType := sliceValue.Type()
	//	if sliceValueType.AssignableTo(bytesType) {
	//		seri := serializer.NewSerializer()
	//		return seri.WriteBytes(sliceValue.Bytes(), func(err error) error {
	//			return errors.Wrap(err, "failed to write array of bytes to serializer")
	//		}).Serialize()
	//	}
	//	return api.encodeSlice(ctx, sliceValue, sliceValueType, ts, opts)
	//case reflect.Interface:
	//	return api.encodeInterface(ctx, value, valueType, ts, opts)
	//case reflect.String:
	//	lengthPrefixType, set := ts.LengthPrefixType()
	//	if !set {
	//		return nil, errors.Errorf("in order to serialize 'string' type no LengthPrefixType was provided")
	//	}
	//	seri := serializer.NewSerializer()
	//	return seri.WriteString(value.String(), lengthPrefixType, func(err error) error {
	//		return errors.Wrap(err, "failed to write string value to serializer")
	//	}).Serialize()
	//
	//case reflect.Bool:
	//	seri := serializer.NewSerializer()
	//	return seri.WriteBool(value.Bool(), func(err error) error {
	//		return errors.Wrap(err, "failed to write bool value to serializer")
	//	}).Serialize()
	//
	//case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	//	reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	//	reflect.Float32, reflect.Float64:
	//	seri := serializer.NewSerializer()
	//	return seri.WriteNum(valueI, func(err error) error {
	//		return errors.Wrap(err, "failed to write number value to serializer")
	//	}).Serialize()
	//default:
	//}
	//return nil, errors.Errorf("can't encode: unsupported type %T", valueI)
	return 0, nil
}

func (api *API) decodeStruct(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, opts *options) (int, error) {
	return 0, nil
}

func (api *API) decodeSlice(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, opts *options) (int, error) {
	//if valueType.AssignableTo(bytesType) {
	//	lengthPrefixType, set := ts.LengthPrefixType()
	//	if !set {
	//		return nil, errors.Errorf("no LengthPrefixType was provided for slice type %s", valueType)
	//	}
	//	seri := serializer.NewSerializer()
	//	seri.WriteVariableByteSlice(value.Bytes(), lengthPrefixType, func(err error) error {
	//		return errors.Wrap(err, "failed to write bytes to serializer")
	//	})
	//	return seri.Serialize()
	//}
	//deseri := serializer.NewDeserializer(b)
	//newValue:=reflect.New(valueType)
	//value.Set()
	//deserializeItem:= func(d *Deserializable) {[]}
	//deseri.ReadSequenceOfObjects()
	//reflect.Append()
	//sliceLen := value.Len()
	//data := make([][]byte, sliceLen)
	//for i := 0; i < sliceLen; i++ {
	//	elemValue := value.Index(i)
	//	elemBytes, err := api.encode(ctx, elemValue, TypeSettings{}, opts)
	//	if err != nil {
	//		return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
	//	}
	//	data[i] = elemBytes
	//}
	//return encodeSliceOfBytes(data, valueType, ts, opts)
	return 0, nil
}

func (api *API) decodeInterface(ctx context.Context, b []byte, value reflect.Value, opts *options) (int, error) {
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
