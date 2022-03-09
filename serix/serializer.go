package serix

import (
	"context"
	"reflect"
	"sync"

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
	switch value.Kind() {
	case reflect.Struct:
		return api.encodeStruct(ctx, value, opts...)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, opts...)
	case reflect.Map:
		return api.encodeMap(ctx, value, opts...)
	case reflect.Interface:
		return nil, nil

	}
	return nil, nil

}

func (api *API) encodeStruct(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
	return nil, nil
}

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, opts ...Option) ([]byte, error) {
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
