package serix

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func (api *API) decode(ctx context.Context, b []byte, value reflect.Value, opts *options) (int, error) {
	valueI := value.Interface()
	if opts.validation {
		if bytesValidator, ok := valueI.(BytesValidator); ok {
			if err := bytesValidator.ValidateBytes(b); err != nil {
				return 0, errors.Wrap(err, "pre-deserialization bytes validation failed")
			}
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
		bytesRead, err = api.decodeBasedOnType(ctx, b, value, valueI, opts)
		if err != nil {
			return 0, errors.WithStack(err)
		}
	}
	if opts.validation {
		if validator, ok := valueI.(SyntacticValidator); ok {
			if err := validator.Validate(); err != nil {
				return 0, errors.Wrap(err, "post-deserialization syntactic validation failed")
			}
		}
	}

	return bytesRead, nil
}

func (api *API) decodeBasedOnType(ctx context.Context, b []byte, value reflect.Value, valueI interface{}, opts *options) (int, error) {
	//switch value.Kind() {
	//case reflect.Ptr:
	//	valueType := value.Type()
	//	if valueType == bigIntPtrType {
	//		bigIntDest := value.Addr().Interface().(**big.Int)
	//		deseri := serializer.NewDeserializer(b)
	//		return deseri.ReadUint256(bigIntDest, func(err error) error {
	//			return errors.Wrap(err, "failed to read math big int from deserializer")
	//		}).Done()
	//	}
	//
	//case reflect.Struct:
	//	valueType := value.Type()
	//	if valueType == timeType {
	//		timeDest := value.Addr().Interface().(*time.Time)
	//		deseri := serializer.NewDeserializer(b)
	//		return deseri.ReadTime(timeDest, func(err error) error {
	//			return errors.Wrap(err, "failed to write time to serializer")
	//		}).Done()
	//
	//	}
	//	return api.decodeStruct(ctx, b, value, valueI, valueType, opts)
	//case reflect.Slice:
	//	return api.decodeSlice(ctx, b, value, valueI, value.Type(), opts)
	//case reflect.Array:
	//	value = sliceFromArray(value)
	//	valueType := value.Type()
	//	if valueType.AssignableTo(bytesType) {
	//		seri := serializer.NewSerializer()
	//		return seri.WriteBytes(value.Bytes(), func(err error) error {
	//			return errors.Wrap(err, "failed to write array of bytes to serializer")
	//		}).Serialize()
	//	}
	//	return api.decodeSlice(ctx, b, value, valueI, valueType, opts)
	//case reflect.Interface:
	//	return api.decodeInterface(ctx, b, value, opts)
	//case reflect.String:
	//	if lptp, ok := valueI.(LengthPrefixTypeProvider); ok {
	//		seri := serializer.NewSerializer()
	//		return seri.WriteString(value.String(), lptp.LengthPrefixType(), func(err error) error {
	//			return errors.Wrap(err, "failed to write string value to serializer")
	//		}).Serialize()
	//	} else {
	//		return nil, errors.New(
	//			`in order to serialize "string" type in must implement LengthPrefixTypeProvider interface`,
	//		)
	//	}
	//case reflect.Bool:
	//	seri := serializer.NewSerializer()
	//	return seri.WriteBool(value.Bool(), func(err error) error {
	//		return errors.Wrap(err, "failed to write bool value to serializer")
	//	}).Serialize()
	//
	//case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	//	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	//	seri := serializer.NewSerializer()
	//	return seri.WriteNum(valueI, func(err error) error {
	//		return errors.Wrap(err, "failed to write number value to serializer")
	//	}).Serialize()
	//default:
	//	return nil, errors.Errorf("can't encode type %T, unsupported kind %s", valueI, value.Kind())
	//}
	return 0, nil

}

func (api *API) decodeStruct(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, opts *options) (int, error) {
	return 0, nil
}

func (api *API) decodeSlice(ctx context.Context, b []byte, value reflect.Value, valueI interface{},
	valueType reflect.Type, opts *options) (int, error) {
	return 0, nil
}

func (api *API) decodeInterface(ctx context.Context, b []byte, value reflect.Value, opts *options) (int, error) {
	return 0, nil
}
