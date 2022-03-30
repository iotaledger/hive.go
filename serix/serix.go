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

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/serializer/v2"
)

type Serializable interface {
	Encode() ([]byte, error)
}

type Deserializable interface {
	Decode(b []byte) (int, error)
}

type BytesValidator interface {
	ValidateBytes([]byte) error
}

type SyntacticValidator interface {
	Validate() error
}

type API struct {
	interfacesRegistryMutex sync.RWMutex
	interfacesRegistry      map[reflect.Type]*interfaceObjects

	typeSettingsRegistryMutex sync.RWMutex
	typeSettingsRegistry      map[reflect.Type]TypeSettings
}

type interfaceObjects struct {
	fromCodeToType map[interface{}]reflect.Type
	fromTypeToCode map[reflect.Type]interface{}
	typeDenotation serializer.TypeDenotationType
}

func NewAPI() *API {
	api := &API{
		interfacesRegistry:   map[reflect.Type]*interfaceObjects{},
		typeSettingsRegistry: map[reflect.Type]TypeSettings{},
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

func WithTypeSettings(ts TypeSettings) Option {
	return func(o *options) {
		o.ts = ts
	}
}

type options struct {
	validation bool
	ts         TypeSettings
}

func (o *options) toMode() serializer.DeSerializationMode {
	mode := serializer.DeSeriModeNoValidation
	if o.validation {
		mode |= serializer.DeSeriModePerformValidation
	}
	return mode
}

type TypeSettings struct {
	lengthPrefixType *serializer.SeriLengthPrefixType
	objectCode       interface{}
	lexicalOrdering  *bool
	arrayRules       *serializer.ArrayRules
}

func (ts TypeSettings) WithLengthPrefixType(lpt serializer.SeriLengthPrefixType) TypeSettings {
	ts.lengthPrefixType = &lpt
	return ts
}

func (ts TypeSettings) LengthPrefixType() (serializer.SeriLengthPrefixType, bool) {
	if ts.lengthPrefixType == nil {
		return 0, false
	}
	return *ts.lengthPrefixType, true
}

func (ts TypeSettings) WithObjectCode(code interface{}) TypeSettings {
	ts.objectCode = code
	return ts
}

func (ts TypeSettings) ObjectCode() interface{} {
	return ts.objectCode
}

func (ts TypeSettings) WithLexicalOrdering(val bool) TypeSettings {
	ts.lexicalOrdering = &val
	return ts
}

func (ts TypeSettings) LexicalOrdering() (val bool, set bool) {
	if ts.lexicalOrdering == nil {
		return false, false
	}
	return *ts.lexicalOrdering, true
}

func (ts TypeSettings) WithArrayRules(rules *serializer.ArrayRules) TypeSettings {
	ts.arrayRules = rules
	return ts
}

func (ts TypeSettings) ArrayRules() *serializer.ArrayRules {
	return ts.arrayRules
}

func (ts TypeSettings) merge(other TypeSettings) TypeSettings {
	if ts.lengthPrefixType == nil {
		ts.lengthPrefixType = other.lengthPrefixType
	}
	if ts.objectCode == nil {
		ts.objectCode = other.objectCode
	}
	if ts.lexicalOrdering == nil {
		ts.lexicalOrdering = other.lexicalOrdering
	}
	if ts.arrayRules == nil {
		ts.arrayRules = other.arrayRules
	}
	return ts
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
	return api.encode(ctx, value, opt.ts, opt)
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

func (api *API) RegisterTypeSettings(obj interface{}, ts TypeSettings) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return errors.New("'obj' is a nil interface, it's need to be a valid non-interface type")
	}
	if isUnderlyingInterface(objType) {
		return errors.New("'obj' is a pointer to an interface, it's need to be a valid non-interface type")
	}
	api.typeSettingsRegistryMutex.Lock()
	defer api.typeSettingsRegistryMutex.Unlock()
	api.typeSettingsRegistry[objType] = ts
	return nil
}

func (api *API) getTypeSettings(objType reflect.Type) TypeSettings {
	api.typeSettingsRegistryMutex.RLock()
	defer api.typeSettingsRegistryMutex.RUnlock()
	return api.typeSettingsRegistry[objType]
}

func (api *API) RegisterInterfaceObjects(iType interface{}, objs ...interface{}) error {
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
	iRegistry := &interfaceObjects{
		fromCodeToType: make(map[interface{}]reflect.Type, len(objs)),
		fromTypeToCode: make(map[reflect.Type]interface{}, len(objs)),
	}

	for i := range objs {
		obj := objs[i]
		objType := reflect.TypeOf(obj)
		objTypeDenotation, objCode, err := api.getTypeDenotationAndObjectCode(objType)
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
		iRegistry.fromCodeToType[objCode] = objType
		iRegistry.fromTypeToCode[objType] = objCode
	}
	api.interfacesRegistryMutex.Lock()
	defer api.interfacesRegistryMutex.Unlock()
	api.interfacesRegistry[iTypeReflect] = iRegistry
	return nil
}

func (api *API) getTypeDenotationAndObjectCode(objType reflect.Type) (serializer.TypeDenotationType, interface{}, error) {
	ts := api.getTypeSettings(objType)
	objectCode := ts.ObjectCode()
	if objectCode == nil {
		return 0, nil, errors.Errorf(
			"type settings for object %s doesn't contain object code, "+
				"you must register object with its type settings first",
			objType,
		)
	}
	objTypeDenotation, err := getTypeDenotationType(objectCode)
	if err != nil {
		return 0, nil, errors.WithStack(err)
	}
	return objTypeDenotation, objectCode, nil
}

func getTypeDenotationType(objectCode interface{}) (serializer.TypeDenotationType, error) {
	objCodeType := reflect.TypeOf(objectCode)
	if objCodeType == nil {
		return 0, errors.New("can't detect type denotation type: object code is nil interface")
	}
	var objTypeDenotation serializer.TypeDenotationType
	switch objCodeType.Kind() {
	case reflect.Uint32:
		objTypeDenotation = serializer.TypeDenotationUint32
	case reflect.Uint8:
		objTypeDenotation = serializer.TypeDenotationByte
	default:
		return 0, errors.Errorf("unsupported object code type: %s (%s), only uint32 and byte are supported",
			objCodeType, objCodeType.Kind())
	}

	return objTypeDenotation, nil
}

func (api *API) getInterfaceObjects(iType reflect.Type) *interfaceObjects {
	api.interfacesRegistryMutex.RLock()
	defer api.interfacesRegistryMutex.RUnlock()
	return api.interfacesRegistry[iType]
}

type structField struct {
	name             string
	index            int
	fType            reflect.Type
	isEmbeddedStruct bool
	settings         tagSettings
}

type tagSettings struct {
	position  int
	isPayload bool
	nest      bool
	ts        TypeSettings
}

func parseStructType(structType reflect.Type) ([]structField, error) {
	structFields := make([]structField, 0, structType.NumField())
	seenPositions := make(map[int]struct{})
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		isUnexported := field.PkgPath != ""
		isEmbedded := field.Anonymous
		isStruct := isUnderlyingStruct(field.Type)
		isEmbeddedStruct := isEmbedded && isStruct
		if isUnexported && !isEmbeddedStruct {
			continue
		}
		tag, ok := field.Tag.Lookup("serix")
		if !ok {
			continue
		}
		tSettings, err := parseStructTag(tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse struct tag %s for field %s", tag, field.Name)
		}
		if _, exists := seenPositions[tSettings.position]; exists {
			return nil, errors.Errorf("struct field with dupicated position number %d", tSettings.position)
		}
		seenPositions[tSettings.position] = struct{}{}
		if tSettings.isPayload {
			if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
				return nil, errors.Errorf(
					"struct field %s is invalid: "+
						"'payload' setting can only be used with pointers or interfaces, got %s",
					field.Name, field.Type.Kind())
			}
			if isEmbeddedStruct {
				return nil, errors.Errorf(
					"struct field %s is invalid: 'payload' setting can't be used with embedded structs",
					field.Name)
			}
		}
		if tSettings.nest && isUnexported {
			return nil, errors.Errorf(
				"struct field %s is invalid: 'nest' setting can't be used with unexported types",
				field.Name)
		}
		structFields = append(structFields, structField{
			name:             field.Name,
			index:            i,
			fType:            field.Type,
			isEmbeddedStruct: isEmbeddedStruct,
			settings:         tSettings,
		})
	}
	sort.Slice(structFields, func(i, j int) bool {
		return structFields[i].settings.position < structFields[j].settings.position
	})
	return structFields, nil
}

func isUnderlyingStruct(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

func isUnderlyingInterface(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Interface
}

func parseStructTag(tag string) (tagSettings, error) {
	if tag == "" {
		return tagSettings{}, errors.New("struct tag is empty")
	}
	parts := strings.Split(tag, ",")
	positionPart := parts[0]
	position, err := strconv.Atoi(positionPart)
	if err != nil {
		return tagSettings{}, errors.Wrap(err, "failed to parse position number from the first part of the tag")
	}
	settings := tagSettings{}
	settings.position = position
	parts = parts[1:]
	seenParts := map[string]struct{}{}
	for _, currentPart := range parts {
		if _, ok := seenParts[currentPart]; ok {
			return tagSettings{}, errors.Errorf("duplicated tag part: %s", currentPart)
		}
		keyValue := strings.Split(currentPart, "=")
		partName := keyValue[0]
		switch partName {
		case "payload":
			settings.isPayload = true
		case "nest":
			settings.nest = true
		case "lengthPrefixType":
			if len(keyValue) != 2 {
				return tagSettings{}, errors.Errorf("incorrect lengthPrefixType tag format: %s", currentPart)
			}
			lengthPrefixType, err := parseLengthPrefixType(keyValue[1])
			if err != nil {
				return tagSettings{}, errors.Wrapf(err, "failed to parse lengthPrefixType %s", currentPart)
			}
			settings.ts = settings.ts.WithLengthPrefixType(lengthPrefixType)
		default:
			return tagSettings{}, errors.Errorf("unknown tag part: %s", currentPart)
		}
		seenParts[partName] = struct{}{}
	}
	return settings, nil
}

func parseLengthPrefixType(prefixTypeRaw string) (serializer.SeriLengthPrefixType, error) {
	switch prefixTypeRaw {
	case "byte", "uint8":
		return serializer.SeriLengthPrefixTypeAsByte, nil
	case "uint16":
		return serializer.SeriLengthPrefixTypeAsUint16, nil
	case "uint32":
		return serializer.SeriLengthPrefixTypeAsUint32, nil
	default:
		return serializer.SeriLengthPrefixTypeAsByte, errors.Errorf("unknown length prefix type: %s", prefixTypeRaw)
	}
}

func sliceFromArray(arrValue reflect.Value) reflect.Value {
	arrType := arrValue.Type()
	sliceType := reflect.SliceOf(arrType.Elem())
	sliceValue := reflect.MakeSlice(sliceType, arrType.Len(), arrType.Len())
	reflect.Copy(sliceValue, arrValue)
	return sliceValue
}

type keyValuePair struct {
	Key   interface{} `serix:"0"`
	Value interface{} `serix:"1"`
}
