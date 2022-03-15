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
	ObjectCode() uint32
}

type LengthPrefixTypeProvider interface {
	LengthPrefixType() serializer.SeriLengthPrefixType
}

type TypeDenotationTypeProvider interface {
	TypeDenotation() serializer.TypeDenotationType
}

type API struct {
	typesRegistryMutex sync.RWMutex
	typesRegistry      map[reflect.Type]*objectsMapping
}

type objectsMapping struct {
	fromCodeToType map[uint32]reflect.Type
	fromTypeToCode map[reflect.Type]uint32
}

func NewAPI() *API {
	api := &API{
		typesRegistry: map[reflect.Type]*objectsMapping{},
	}
	return api
}

var (
	bytesType = reflect.TypeOf([]byte(nil))
)

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

func (api *API) Encode(ctx context.Context, obj interface{}, opts ...Option) ([]byte, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, nil
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	return api.encode(ctx, value, opt)
}

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
		bytes, err = serializable.Serialize()
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
		if valueBigInt, ok := valueI.(*big.Int); ok {
			seri := serializer.NewSerializer()
			return seri.WriteUint256(valueBigInt, func(err error) error {
				return errors.Wrap(err, "failed to write math big int to serializer")
			}).Serialize()
		}
		return api.encode(ctx, value.Elem(), opts)

	case reflect.Struct:
		if valueTime, ok := valueI.(time.Time); ok {
			seri := serializer.NewSerializer()
			return seri.WriteTime(valueTime, func(err error) error {
				return errors.Wrap(err, "failed to write time to serializer")
			}).Serialize()
		}
		return api.encodeStruct(ctx, value, opts)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, valueI, value.Type(), opts)
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
	default:
		return nil, errors.Errorf("can't encode type %T, unsupported kind %s", valueI, value.Kind())
	}
}

func sliceFromArray(arrValue reflect.Value) reflect.Value {
	arrType := arrValue.Type()
	sliceType := reflect.SliceOf(arrType.Elem())
	sliceValue := reflect.MakeSlice(sliceType, arrType.Len(), arrType.Len())
	reflect.Copy(sliceValue, arrValue)
	return sliceValue
}

func (api *API) encodeInterface(ctx context.Context, value reflect.Value, opts *options) ([]byte, error) {
	valueType := value.Type()
	elemValue := value.Elem()
	if !elemValue.IsValid() {
		return nil, errors.Errorf("can't serialize interface %s it must have underlying value", valueType)
	}
	mapping := api.getObjectsMapping(valueType)
	if mapping == nil {
		return nil, errors.Errorf("interface %s isn't registered", valueType)
	}
	elemType := elemValue.Type()
	if _, exists := mapping.fromTypeToCode[elemType]; !exists {
		return nil, errors.Errorf("underlying type %s hasn't been registered for interface type %s",
			elemType, valueType)
	}
	encodedBytes, err := api.encode(ctx, elemValue, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode interface element %s", elemType)
	}
	return encodedBytes, nil
}

func (api *API) getObjectsMapping(iType reflect.Type) *objectsMapping {
	api.typesRegistryMutex.RLock()
	defer api.typesRegistryMutex.RUnlock()
	return api.typesRegistry[iType]
}

func (api *API) encodeStruct(ctx context.Context, value reflect.Value, opts *options) ([]byte, error) {
	valueType := value.Type()
	structFields, err := parseStructType(valueType)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil, nil
	}
	s := serializer.NewSerializer()
	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
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
				return nil, errors.Wrapf(err, "failed to serialize payload struct field %s", sField.name)
			}
			s.WritePayloadLength(len(payloadBytes), func(err error) error {
				return errors.Wrapf(err,
					"failed to write payload length for struct field %s to serializer",
					sField.name,
				)
			})
			fieldBytes = payloadBytes
		}
		if fieldBytes == nil {
			var err error
			fieldBytes, err = api.encode(ctx, fieldValue, opts)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to serialize struct field %s", sField.name)
			}
		}
		s.WriteBytes(fieldBytes, func(err error) error {
			return errors.Wrapf(err,
				"failed to write serialized struct field bytes to serializer, field=%s",
				sField.name,
			)
		})
	}
	return nil, nil
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

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type, opts *options) ([]byte, error) {
	lptp, ok := valueI.(LengthPrefixTypeProvider)
	if !ok {
		return nil, errors.Errorf("slice type %s must implement LengthPrefixTypeProvider interface", valueType)
	}
	lengthPrefixType := lptp.LengthPrefixType()

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
		elemBytes, err := api.encode(ctx, elemValue, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
		}
		data[i] = elemBytes
	}
	var arrayRules *serializer.ArrayRules
	if ruler, ok := valueI.(ArrayRulesProvider); ok {
		arrayRules = ruler.ArrayRules()
	}
	seri := serializer.NewSerializer()
	return seri.WriteSliceOfByteSlices(data, opts.toMode(), lengthPrefixType, arrayRules, func(err error) error {
		return errors.Wrapf(err,
			"serializer failed to write slice of objects %s as slice of byte slices", valueType,
		)
	}).Serialize()
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
	mapping := &objectsMapping{
		fromCodeToType: make(map[uint32]reflect.Type, len(objs)),
		fromTypeToCode: make(map[reflect.Type]uint32, len(objs)),
	}
	for _, obj := range objs {
		objCode := obj.ObjectCode()
		objType := reflect.TypeOf(obj)
		mapping.fromCodeToType[objCode] = objType
		mapping.fromTypeToCode[objType] = objCode
	}
	api.typesRegistryMutex.Lock()
	defer api.typesRegistryMutex.Unlock()
	api.typesRegistry[iTypeReflect] = mapping
	return nil
}
