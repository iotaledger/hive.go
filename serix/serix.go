package serix

import (
	"context"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/pkg/errors"
)

type Serializable interface {
	Serialize(ctx context.Context, opts ...Option) ([]byte, error)
}

type Deserializable interface {
	Deserialize(ctx context.Context, b []byte, opts ...Option) (int, error)
}

type BytesValidator interface {
	ValidateBytes(context.Context, []byte) error
}

type SyntacticValidator interface {
	Validate(context.Context) error
}

type ArrayRulesProvider interface {
	ArrayRules() *serializer.ArrayRules
}

type ObjectTypeProvider interface {
	ObjectType() uint32
}

type LengthPrefixTypeProvider interface {
	LengthPrefixType() serializer.SeriLengthPrefixType
}

type TypeDenotationTypeProvider interface {
	TypeDenotation() serializer.TypeDenotationType
}

type API struct {
	typesRegistryMutex sync.RWMutex
	typesRegistry      map[interface{}]map[uint32]interface{}
}

func NewAPI() *API {
	api := &API{
		typesRegistry: map[interface{}]map[uint32]interface{}{},
	}
	return api
}

var (
	serializerType         = reflect.TypeOf((*Serializable)(nil)).Elem()
	deserializerType       = reflect.TypeOf((*Deserializable)(nil)).Elem()
	bytesValidatorType     = reflect.TypeOf((*BytesValidator)(nil)).Elem()
	syntacticValidatorType = reflect.TypeOf((*SyntacticValidator)(nil)).Elem()
	bytesType              = reflect.TypeOf([]byte(nil))
	byteType               = reflect.TypeOf(byte(0))
)

type options struct {
	noValidation    bool
	lexicalOrdering bool
}

type Option func(o *options)

func WithNoValidation() Option {
	return func(o *options) {
		o.noValidation = true
	}
}

func WithLexicalOrdering() Option {
	return func(o *options) {
		o.lexicalOrdering = true
	}
}

func (api *API) Encode(ctx context.Context, obj interface{}, opts ...Option) ([]byte, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, nil
	}
	return api.encode(ctx, value, opts...)
}

func (api *API) encode(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	valueI := value.Interface()
	if validator, ok := valueI.(SyntacticValidator); ok {
		if err := validator.Validate(ctx); err != nil {
			return nil, errors.Wrap(err, "pre-serialization syntactic validation failed")
		}
	}
	var bytes []byte
	if serializable, ok := valueI.(Serializable); ok {
		var err error
		bytes, err = serializable.Serialize(ctx, opts...)
		if err != nil {
			return nil, errors.Wrap(err, "object failed to serialize itself")
		}
	} else {
		var err error
		bytes, err = api.encodeBasedOnType(ctx, value, opts...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if bytesValidator, ok := valueI.(BytesValidator); ok {
		if err := bytesValidator.ValidateBytes(ctx, bytes); err != nil {
			return nil, errors.Wrap(err, "post-serialization bytes validation failed")
		}
	}

	return bytes, nil
}

func (api *API) encodeBasedOnType(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	value = deReferencePointer(value)
	valueI := value.Interface()
	switch value.Kind() {
	case reflect.Struct:
		if valueBigInt, ok := valueI.(*big.Int); ok {
			seri := serializer.NewSerializer()
			return seri.WriteUint256(valueBigInt, func(err error) error {
				return errors.Wrap(err, "failed to write math big int to serializer")
			}).Serialize()
		} else if valueTime, ok := valueI.(time.Time); ok {
			seri := serializer.NewSerializer()
			return seri.WriteTime(valueTime, func(err error) error {
				return errors.Wrap(err, "failed to write time to serializer")
			}).Serialize()
		}
		return api.encodeStruct(ctx, value, opts...)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, opts...)
	case reflect.Array:
		return api.encodeSlice(ctx, sliceFromArray(value), opts...)
	case reflect.Interface:
		return nil, nil
	case reflect.String:
		if lptp, ok := valueI.(LengthPrefixTypeProvider); ok {
			seri := serializer.NewSerializer()
			return seri.WriteString(value.String(), lptp.LengthPrefixType(), func(err error) error {
				return errors.Wrap(err, "failed to write string value to serializer")
			}).Serialize()
		} else {
			return nil, errors.New(
				`in order to serialize "string" type in must implement LengthPrefixTypeProvider interface`,
			)
		}
	case reflect.Bool:
		seri := serializer.NewSerializer()
		return seri.WriteBool(value.Bool(), func(err error) error {
			return errors.Wrap(err, "failed to write bool value to serializer")
		}).Serialize()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		seri := serializer.NewSerializer()
		return seri.WriteNum(valueI, func(err error) error {
			return errors.Wrap(err, "failed to write number value to serializer")
		}).Serialize()
	}
	return nil, nil

}

func sliceFromArray(arrValue reflect.Value) reflect.Value {
	arrType := arrValue.Type()
	sliceType := reflect.SliceOf(arrType.Elem())
	sliceValue := reflect.MakeSlice(sliceType, arrType.Len(), arrType.Len())
	reflect.Copy(sliceValue, arrValue)
	return sliceValue
}

func (api *API) encodeStruct(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	valueType := value.Type()
	structFields := make([]*structField, 0, value.NumField())
	for i := 0; i < value.NumField(); i++ {
		fieldReflect := valueType.Field(i)
		if fieldReflect.PkgPath != "" {
			continue
		}
		tag, ok := fieldReflect.Tag.Lookup("seri")
		if !ok {
			continue
		}
		tSettings, err := parseStructTag(tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse struct tag %s for field %s", tag, fieldReflect.Name)
		}
		structFields = append(structFields, &structField{
			settings: tSettings,
			value:    value.Field(i),
		})
	}
	sort.Slice(structFields, func(i, j int) bool {
		return structFields[i].settings.position < structFields[j].settings.position
	})
	if len(structFields) == 0 {
		return nil, nil
	}
	//s := serializer.NewSerializer()
	//for _, sField := range structFields {
	//
	//}
	return nil, nil
}

type structField struct {
	settings *tagSettings
	value    reflect.Value
}

type tagSettings struct {
	position    int
	isPayload   bool
	isVarLength bool
}

func parseStructTag(tag string) (*tagSettings, error) {
	if tag == "" {
		return nil, errors.New("struct tag is empty")
	}
	parts := strings.Split(tag, ",")
	positionPart := parts[0]
	position, err := strconv.Atoi(positionPart)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse position number from the first part of the tag")
	}
	settings := &tagSettings{}
	settings.position = position
	parts = parts[1:]
	seenParts := map[string]struct{}{}
	for _, currentPart := range parts {
		if _, ok := seenParts[currentPart]; ok {
			return nil, errors.Errorf("duplicated tag part: %s", currentPart)
		}
		switch currentPart {
		case "payload":
			settings.isPayload = true
		case "varLength":
			settings.isVarLength = true
		default:
			return nil, errors.Errorf("unknown tag part: %s", currentPart)
		}
		seenParts[currentPart] = struct{}{}
	}
	return settings, nil
}

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	if value.Type().AssignableTo(bytesType) {
		seri := serializer.NewSerializer()
		seri.WriteBytes(value.Bytes(), func(err error) error {
			return errors.Wrap(err, "failed to write bytes to serializer")
		})
		return seri.Serialize()
	}
	return nil, nil
}

func (api *API) encodeMap(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	return nil, nil
}

func (api *API) Decode(ctx context.Context, b []byte, obj interface{}, opts ...Option) error {
	return nil
}

func (api *API) EncodeJSON(obj interface{}) ([]byte, error) {
	return nil, nil
}

func (api *API) DecodeJSON(b []byte, obj interface{}) error {
	return nil
}

func (api *API) RegisterObjects(iType interface{}, objs ...ObjectTypeProvider) *API {
	mapping := make(map[uint32]interface{}, len(objs))
	for _, obj := range objs {
		mapping[obj.ObjectType()] = obj
	}
	api.typesRegistryMutex.Lock()
	defer api.typesRegistryMutex.Unlock()
	api.typesRegistry[iType] = mapping
	return api
}
