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
	Serialize() ([]byte, error)
}

type Deserializable interface {
	Deserialize(b []byte) (int, error)
}

type BytesValidator interface {
	ValidateBytes([]byte) error
}

type SyntacticValidator interface {
	Validate() error
}

type ArrayRulesProvider interface {
	ArrayRules() *serializer.ArrayRules
}

type ObjectCodeProvider interface {
	ObjectCode() interface{}
}

type LengthPrefixTypeProvider interface {
	LengthPrefixType() serializer.SeriLengthPrefixType
}

type TypeDenotationTypeProvider interface {
	TypeDenotation() serializer.TypeDenotationType
}

type API struct {
	typesRegistryMutex sync.RWMutex
	typesRegistry      map[reflect.Type]*interfaceRegistry
}

type interfaceRegistry struct {
	fromCodeToType map[interface{}]reflect.Type
	fromTypeToCode map[reflect.Type]interface{}
	typeDenotation serializer.TypeDenotationType
}

func NewAPI() *API {
	api := &API{
		typesRegistry: map[reflect.Type]*interfaceRegistry{},
	}
	return api
}

var (
	bytesType     = reflect.TypeOf([]byte(nil))
	bigIntPtrType = reflect.TypeOf((*big.Int)(nil))
	timeType      = reflect.TypeOf(time.Time{})
)

type Option func(o *options)

func WithValidation() Option {
	return func(o *options) {
		o.validation = true
	}
}

func WithLexicalOrdering() Option {
	return func(o *options) {
		o.lexicalOrdering = true
	}
}

type options struct {
	validation      bool
	lexicalOrdering bool
}

func (o *options) toMode() serializer.DeSerializationMode {
	mode := serializer.DeSeriModeNoValidation
	if o.validation {
		mode |= serializer.DeSeriModePerformValidation
	}
	if o.lexicalOrdering {
		mode |= serializer.DeSeriModePerformLexicalOrdering
	}
	return mode
}

func (api *API) Encode(ctx context.Context, obj interface{}, opts ...Option) ([]byte, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, errors.New("invalid value for destination")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	return api.encode(ctx, value, opt)
}

func (api *API) Decode(ctx context.Context, b []byte, obj interface{}, opts ...Option) error {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return errors.New("invalid value for destination")
	}
	if value.Kind() != reflect.Ptr {
		return errors.Errorf(
			"can't decode, the destination object must be a pointer, got: %T(%s)", obj, value.Kind(),
		)
	}
	if value.IsNil() {
		return errors.Errorf("can't decode, the destination object %T must be a non-nil pointer", obj)
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	_, err := api.decode(ctx, b, value, opt)
	return errors.WithStack(err)
}

func (api *API) EncodeJSON(obj interface{}) ([]byte, error) {
	return nil, nil
}

func (api *API) DecodeJSON(b []byte, obj interface{}) error {
	return nil
}

func (api *API) RegisterObjects(iType interface{}, objs ...ObjectCodeProvider) error {
	ptrType := reflect.TypeOf(iType)
	if ptrType == nil {
		return errors.New("'iType' is a nil interface, it's need to be a pointer to an interface")
	}
	if ptrType.Kind() != reflect.Ptr {
		return errors.Errorf("'iType' parameter must be a pointer, got %s", ptrType.Kind())
	}
	iTypeReflect := ptrType.Elem()
	if iTypeReflect.Kind() != reflect.Interface {
		return errors.Errorf(
			"'iType' pointer must contain an interface, got %s", iTypeReflect.Kind())
	}
	if len(objs) == 0 {
		return nil
	}
	iRegistry := &interfaceRegistry{
		fromCodeToType: make(map[interface{}]reflect.Type, len(objs)),
		fromTypeToCode: make(map[reflect.Type]interface{}, len(objs)),
	}

	for i := range objs {
		obj := objs[i]
		objTypeDenotation, objCode, err := getTypeDenotationAndObjectCode(obj)
		if err != nil {
			return errors.Wrapf(err, "failed to get type denotation for object %T", obj)
		}
		if i == 0 {
			iRegistry.typeDenotation = objTypeDenotation
		} else {
			if iRegistry.typeDenotation != objTypeDenotation {
				firstObj := objs[0]
				return errors.Errorf(
					"all registered objects must have the same type denotation: object %T has %s and object %T has %s",
					firstObj, iRegistry.typeDenotation, obj, objTypeDenotation,
				)
			}
		}
		objType := reflect.TypeOf(obj)
		iRegistry.fromCodeToType[objCode] = objType
		iRegistry.fromTypeToCode[objType] = objCode
	}
	api.typesRegistryMutex.Lock()
	defer api.typesRegistryMutex.Unlock()
	api.typesRegistry[iTypeReflect] = iRegistry
	return nil
}

func getTypeDenotationAndObjectCode(obj ObjectCodeProvider) (serializer.TypeDenotationType, interface{}, error) {
	objCode := obj.ObjectCode()
	objCodeType := reflect.TypeOf(objCode)
	if objCodeType == nil {
		return 0, nil, errors.Errorf("can't detect object code type for object %T", obj)
	}
	var objTypeDenotation serializer.TypeDenotationType
	switch objCodeType.Kind() {
	case reflect.Uint32:
		objTypeDenotation = serializer.TypeDenotationUint32
	case reflect.Uint8:
		objTypeDenotation = serializer.TypeDenotationByte
	default:
		return 0, nil, errors.Errorf("unsupported object code type: %s (%s), only uint32 and byte are supported",
			objCodeType, objCodeType.Kind())
	}
	return objTypeDenotation, objCodeType, nil
}

func (api *API) getInterfaceRegistry(iType reflect.Type) *interfaceRegistry {
	api.typesRegistryMutex.RLock()
	defer api.typesRegistryMutex.RUnlock()
	return api.typesRegistry[iType]
}

type structField struct {
	name     string
	index    int
	fType    reflect.Type
	settings *tagSettings
}

type tagSettings struct {
	position  int
	isPayload bool
}

func parseStructType(structType reflect.Type) ([]*structField, error) {
	structFields := make([]*structField, 0, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag, ok := field.Tag.Lookup("seri")
		if !ok {
			continue
		}
		tSettings, err := parseStructTag(tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse struct tag %s for field %s", tag, field.Name)
		}
		if tSettings.isPayload {
			if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
				return nil, errors.Errorf(
					"struct field %s is invalid: "+
						"'payload' setting can only be used with pointers or interfaces, got %s",
					field.Name, field.Type.Kind())
			}

		}
		structFields = append(structFields, &structField{
			name:     field.Name,
			index:    i,
			fType:    field.Type,
			settings: tSettings,
		})
	}
	sort.Slice(structFields, func(i, j int) bool {
		return structFields[i].settings.position < structFields[j].settings.position
	})
	return structFields, nil
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
		default:
			return nil, errors.Errorf("unknown tag part: %s", currentPart)
		}
		seenParts[currentPart] = struct{}{}
	}
	return settings, nil
}

func sliceFromArray(arrValue reflect.Value) reflect.Value {
	arrType := arrValue.Type()
	sliceType := reflect.SliceOf(arrType.Elem())
	sliceValue := reflect.MakeSlice(sliceType, arrType.Len(), arrType.Len())
	reflect.Copy(sliceValue, arrValue)
	return sliceValue
}
