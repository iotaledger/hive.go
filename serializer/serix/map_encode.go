package serix

import (
	"context"
	"math/big"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
)

const (
	// the map key under which the object code is written.
	mapTypeKeyName = "type"
	// key used when no map key is defined for types which are slice/arrays of bytes.
	mapSliceArrayDefaultKey = "data"
)

var (
	// ErrNonUTF8String gets returned when a non UTF-8 string is being encoded/decoded.
	ErrNonUTF8String = errors.New("non UTF-8 string value")
)

func (api *API) mapEncode(ctx context.Context, value reflect.Value, ts TypeSettings, opts *options) (ele any, err error) {
	valueI := value.Interface()
	valueType := value.Type()
	if opts.validation {
		if err := api.callSyntacticValidator(ctx, value, valueType); err != nil {
			return nil, errors.Wrap(err, "pre-serialization validation failed")
		}
	}

	if serializable, ok := valueI.(SerializableJSON); ok {
		ele, err = serializable.EncodeJSON()
		if err != nil {
			return nil, errors.Wrap(err, "object failed to serialize itself")
		}
	} else {
		ele, err = api.mapEncodeBasedOnType(ctx, value, valueI, valueType, ts, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return ele, nil
}

func (api *API) mapEncodeBasedOnType(
	ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type, ts TypeSettings, opts *options,
) (any, error) {
	globalTS, _ := api.getTypeSettings(valueType)
	ts = ts.merge(globalTS)
	switch value.Kind() {
	case reflect.Ptr:
		if valueBigInt, ok := valueI.(*big.Int); ok {
			return EncodeUint256(valueBigInt), nil
		}

		elemValue := reflect.Indirect(value)
		if !elemValue.IsValid() {
			return nil, errors.Errorf("unexpected nil pointer for type %T", valueI)
		}
		if elemValue.Kind() == reflect.Struct {
			return api.mapEncodeStruct(ctx, elemValue, elemValue.Interface(), elemValue.Type(), ts, opts)
		}
		if elemValue.Kind() == reflect.Array {
			sliceValue := sliceFromArray(elemValue)
			sliceValueType := sliceValue.Type()

			ts, _ = api.getTypeSettings(valueType)

			return api.mapEncodeSlice(ctx, sliceValue, sliceValueType, ts, opts)
		}

	case reflect.Struct:
		return api.mapEncodeStruct(ctx, value, valueI, valueType, ts, opts)
	case reflect.Slice:
		return api.mapEncodeSlice(ctx, value, valueType, ts, opts)
	case reflect.Map:
		return api.mapEncodeMap(ctx, value, opts)
	case reflect.Array:
		sliceValue := sliceFromArray(value)
		sliceValueType := sliceValue.Type()

		return api.mapEncodeSlice(ctx, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.mapEncodeInterface(ctx, value, valueType, opts)
	case reflect.String:
		str := value.String()
		if !utf8.ValidString(str) {
			return nil, ErrNonUTF8String
		}

		return value.String(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Int8, reflect.Int16, reflect.Int32:
		return value.Int(), nil
	case reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return value.Uint(), nil
	case reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'E', -1, 64), nil
	default:
	}

	return nil, errors.Errorf("can't encode: unsupported type %T", valueI)
}

func (api *API) mapEncodeInterface(
	ctx context.Context, value reflect.Value, valueType reflect.Type, opts *options) (any, error) {
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

	elemTypeSettings, _ := api.getTypeSettings(elemType)

	ele, err := api.mapEncode(ctx, elemValue, elemTypeSettings, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode interface element %s", elemType)
	}

	return ele, nil
}

func (api *API) mapEncodeStruct(
	ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type, ts TypeSettings, opts *options,
) (any, error) {
	if valueTime, ok := valueI.(time.Time); ok {
		return valueTime.Format(time.RFC3339Nano), nil
	}

	obj := orderedmap.New()
	if ts.ObjectType() != nil {
		obj.Set(mapTypeKeyName, ts.ObjectType())
	}
	if err := api.mapEncodeStructFields(ctx, obj, value, valueType, opts); err != nil {
		return nil, errors.WithStack(err)
	}

	return obj, nil
}

func (api *API) mapEncodeStructFields(
	ctx context.Context, obj *orderedmap.OrderedMap, value reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := api.parseStructType(valueType)
	if err != nil {
		return errors.Wrapf(err, "can't parse struct type %s", valueType)
	}

	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
		if sField.isEmbeddedStruct && !sField.settings.nest {
			fieldType := sField.fType
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					continue
				}
				fieldValue = fieldValue.Elem()
				fieldType = fieldType.Elem()
			}
			if err := api.mapEncodeStructFields(ctx, obj, fieldValue, fieldType, opts); err != nil {
				return errors.Wrapf(err, "can't serialize embedded struct %s", sField.name)
			}

			continue
		}

		if sField.settings.omitEmpty && fieldValue.IsZero() {
			continue
		}

		var eleOut any
		if sField.settings.isOptional {
			if fieldValue.IsNil() {
				continue
			}
		}

		eleOut, err = api.mapEncode(ctx, fieldValue, sField.settings.ts, opts)
		if err != nil {
			return errors.Wrapf(err, "failed to serialize optional struct field %s", sField.name)
		}

		switch {
		case sField.settings.ts.mapKey != nil:
			obj.Set(*sField.settings.ts.mapKey, eleOut)
		default:
			obj.Set(mapStringKey(sField.name), eleOut)
		}
	}

	return nil
}

func (api *API) mapEncodeSlice(ctx context.Context, value reflect.Value, valueType reflect.Type,
	ts TypeSettings, opts *options) (any, error) {

	if ts.ObjectType() != nil {
		m := orderedmap.New()
		m.Set(mapTypeKeyName, ts.ObjectType())
		mapKey := mapSliceArrayDefaultKey
		if ts.mapKey != nil {
			mapKey = *ts.mapKey
		}
		m.Set(mapKey, EncodeHex(value.Bytes()))

		return m, nil
	}

	if valueType.AssignableTo(bytesType) {
		return EncodeHex(value.Bytes()), nil
	}

	sliceLen := value.Len()
	data := make([]any, sliceLen)
	for i := 0; i < sliceLen; i++ {
		elemValue := value.Index(i)
		elem, err := api.mapEncode(ctx, elemValue, TypeSettings{}, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
		}
		data[i] = elem
	}

	return data, nil
}

func (api *API) mapEncodeMap(ctx context.Context, value reflect.Value, opts *options) (*orderedmap.OrderedMap, error) {
	m := orderedmap.New()
	iter := value.MapRange()
	for i := 0; iter.Next(); i++ {
		key := iter.Key()
		elem := iter.Value()
		k, v, err := api.mapEncodeMapKVPair(ctx, key, elem, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		m.Set(k, v)
	}

	return m, nil
}

func (api *API) mapEncodeMapKVPair(ctx context.Context, key, val reflect.Value, opts *options) (string, any, error) {
	k, err := api.mapEncode(ctx, key, TypeSettings{}, opts)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to encode map key of type %s", key.Type())
	}
	v, err := api.mapEncode(ctx, val, TypeSettings{}, opts)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to encode map element of type %s", val.Type())
	}

	return k.(string), v, nil
}
