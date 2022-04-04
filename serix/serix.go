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

type API struct {
	interfacesRegistryMutex sync.RWMutex
	interfacesRegistry      map[reflect.Type]*interfaceObjects

	typeSettingsRegistryMutex sync.RWMutex
	typeSettingsRegistry      map[reflect.Type]TypeSettings

	validatorsRegistryMutex sync.RWMutex
	validatorsRegistry      map[reflect.Type]validators
}

type validators struct {
	bytesValidator     reflect.Value
	syntacticValidator reflect.Value
}

type interfaceObjects struct {
	fromCodeToType map[interface{}]reflect.Type
	fromTypeToCode map[reflect.Type]interface{}
	typeDenotation serializer.TypeDenotationType
}

var DefaultAPI = NewAPI()

func NewAPI() *API {
	api := &API{
		interfacesRegistry:   map[reflect.Type]*interfaceObjects{},
		typeSettingsRegistry: map[reflect.Type]TypeSettings{},
		validatorsRegistry:   map[reflect.Type]validators{},
	}
	return api
}

var (
	bytesType     = reflect.TypeOf([]byte(nil))
	bigIntPtrType = reflect.TypeOf((*big.Int)(nil))
	timeType      = reflect.TypeOf(time.Time{})
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
	boolType      = reflect.TypeOf(false)
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

func (ts TypeSettings) toMode(opts *options) serializer.DeSerializationMode {
	mode := opts.toMode()
	lexicalOrdering, set := ts.LexicalOrdering()
	if set && lexicalOrdering {
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

func (api *API) RegisterValidators(obj interface{}, bytesValidatorFn interface{}, syntacticValidatorFn interface{}) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return errors.New("'obj' is a nil interface, it's need to be a valid type")
	}
	bytesValidatorValue, err := parseValidatorFunc(objType, bytesValidatorFn)
	if err != nil {
		return errors.Wrapf(err, "failed to parse bytesValidatorFn")
	}
	syntacticValidatorValue, err := parseValidatorFunc(objType, syntacticValidatorFn)
	if err != nil {
		return errors.Wrapf(err, "failed to parse syntacticValidatorFn")
	}
	vldtrs := validators{}
	if bytesValidatorValue.IsValid() {
		if err := checkBytesValidatorSignature(bytesValidatorValue); err != nil {
			return errors.WithStack(err)
		}
		vldtrs.bytesValidator = bytesValidatorValue
	}
	if syntacticValidatorValue.IsValid() {
		if err := checkSyntacticValidatorSignature(syntacticValidatorValue); err != nil {
			return errors.WithStack(err)
		}
		vldtrs.syntacticValidator = syntacticValidatorValue
	}
	api.validatorsRegistryMutex.Lock()
	defer api.validatorsRegistryMutex.Unlock()
	api.validatorsRegistry[objType] = vldtrs
	return nil
}

func parseValidatorFunc(obType reflect.Type, validatorFn interface{}) (reflect.Value, error) {
	if validatorFn == nil {
		return reflect.Value{}, nil
	}
	funcValue := reflect.ValueOf(validatorFn)
	if !funcValue.IsValid() || funcValue.IsZero() {
		return reflect.Value{}, nil
	}
	if funcValue.Kind() != reflect.Func {
		return reflect.Value{}, errors.Errorf(
			"validator must be a function, got %T(%s)", validatorFn, funcValue.Kind(),
		)
	}
	funcType := funcValue.Type()
	if funcType.NumIn() == 0 {
		return reflect.Value{}, errors.Errorf("validator func must have at least one argument")
	}
	firstArgumentType := funcType.In(0)
	if firstArgumentType != obType {
		return reflect.Value{}, errors.Errorf(
			"validator func's first argument must have the same type as the object it's been registered for, "+
				"argumentType=%s, objectType=%s", firstArgumentType, obType,
		)
	}
	if funcType.NumOut() != 1 {
		return reflect.Value{}, errors.Errorf("validator func must have one return value, got %d", funcType.NumOut())
	}
	returnType := funcType.Out(0)
	if returnType != errorType {
		return reflect.Value{}, errors.Errorf("validator func must have 'error' return type, got %s", returnType)
	}
	return funcValue, nil
}

func checkBytesValidatorSignature(funcValue reflect.Value) error {
	funcType := funcValue.Type()
	if funcType.NumIn() != 2 {
		return errors.Errorf("bytesValidatorFn must have two arguments, got %d", funcType.NumIn())
	}
	secondArgumentType := funcType.In(1)
	if secondArgumentType != bytesType {
		return errors.Errorf("bytesValidatorFn's second argument must be bytes, got %s", secondArgumentType)
	}
	return nil
}

func checkSyntacticValidatorSignature(funcValue reflect.Value) error {
	funcType := funcValue.Type()
	if funcType.NumIn() != 1 {
		return errors.Errorf("syntacticValidatorFn must have one argument, got %d", funcType.NumIn())
	}
	return nil
}

func (api *API) callBytesValidator(value reflect.Value, valueType reflect.Type, bytes []byte) error {
	api.validatorsRegistryMutex.RLock()
	defer api.validatorsRegistryMutex.RUnlock()
	bytesValidator := api.validatorsRegistry[valueType].bytesValidator
	if !bytesValidator.IsValid() {
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			value = value.Elem()
			bytesValidator = api.validatorsRegistry[valueType].bytesValidator
		}
	}
	if bytesValidator.IsValid() {
		if err := bytesValidator.Call([]reflect.Value{value, reflect.ValueOf(bytes)})[0].Interface().(error); err != nil {
			return errors.Wrapf(err, "bytes validator returns an error for type %s", valueType)
		}
	}
	return nil
}

func (api *API) callSyntacticValidator(value reflect.Value, valueType reflect.Type) error {
	api.validatorsRegistryMutex.RLock()
	defer api.validatorsRegistryMutex.RUnlock()
	syntacticValidator := api.validatorsRegistry[valueType].syntacticValidator
	if !syntacticValidator.IsValid() {
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			value = value.Elem()
			syntacticValidator = api.validatorsRegistry[valueType].syntacticValidator
		}
	}
	if syntacticValidator.IsValid() {
		if err := syntacticValidator.Call([]reflect.Value{value})[0].Interface().(error); err != nil {
			return errors.Wrapf(err, "syntactic validator returns an error for type %s", valueType)
		}
	}
	return nil
}

func (api *API) RegisterTypeSettings(obj interface{}, ts TypeSettings) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return errors.New("'obj' is a nil interface, it's need to be a valid type")
	}
	api.typeSettingsRegistryMutex.Lock()
	defer api.typeSettingsRegistryMutex.Unlock()
	api.typeSettingsRegistry[objType] = ts
	return nil
}

func (api *API) getTypeSettings(objType reflect.Type) (TypeSettings, bool) {
	api.typeSettingsRegistryMutex.RLock()
	defer api.typeSettingsRegistryMutex.RUnlock()
	ts, ok := api.typeSettingsRegistry[objType]
	if ok {
		return ts, true
	}
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		ts, ok = api.typeSettingsRegistry[objType]
		return ts, ok
	}
	return TypeSettings{}, false
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

	iRegistry, exists := api.interfacesRegistry[iTypeReflect]
	if !exists {
		iRegistry = &interfaceObjects{
			fromCodeToType: make(map[interface{}]reflect.Type, len(objs)),
			fromTypeToCode: make(map[reflect.Type]interface{}, len(objs)),
		}
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
		} else if iRegistry.typeDenotation != objTypeDenotation {
			firstObj := objs[0]
			return errors.Errorf(
				"all registered objects must have the same type denotation: object %T has %s and object %T has %s",
				firstObj, iRegistry.typeDenotation, obj, objTypeDenotation,
			)
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
	ts, exists := api.getTypeSettings(objType)
	if !exists {
		return 0, nil, errors.Errorf(
			"no type settings was found for object %s"+
				"you must register object with its type settings first",
			objType,
		)
	}
	objectCode := ts.ObjectCode()
	if objectCode == nil {
		return 0, nil, errors.Errorf(
			"type settings for object %s doesn't contain object code",
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

func isUnderlyingStruct(t reflect.Type) bool {
	t = deRefPointer(t)
	return t.Kind() == reflect.Struct
}

type iterableMeta struct {
	sizeMethodIndex    int
	forEachMethodIndex int
	iterFuncType       reflect.Type
}

func getOrderedMapMeta(t reflect.Type) (iterableMeta, bool) {
	if t.Kind() == reflect.Interface {
		return iterableMeta{}, false
	}
	forEachMethod, ok := t.MethodByName("ForEach")
	if !ok {
		return iterableMeta{}, false
	}
	sizeMethod, ok := t.MethodByName("Size")
	if !ok {
		return iterableMeta{}, false
	}
	sizeType := sizeMethod.Type
	if sizeType.NumIn() != 1 {
		return iterableMeta{}, false
	}
	if sizeType.NumOut() != 1 {
		return iterableMeta{}, false
	}
	if sizeType.Out(0) != reflect.TypeOf(int(0)) {
		return iterableMeta{}, false
	}
	forEachType := forEachMethod.Type
	if forEachType.NumIn() != 2 {
		return iterableMeta{}, false
	}
	if forEachType.NumOut() != 1 {
		return iterableMeta{}, false
	}
	if forEachType.Out(0) != reflect.TypeOf(true) {
		return iterableMeta{}, false
	}
	iterFuncType := forEachType.In(1)
	if iterFuncType.Kind() != reflect.Func {
		return iterableMeta{}, false
	}
	if iterFuncType.NumIn() != 2 {
		return iterableMeta{}, false
	}
	if iterFuncType.NumOut() != 1 {
		return iterableMeta{}, false
	}
	if iterFuncType.Out(0) != reflect.TypeOf(true) {
		return iterableMeta{}, false
	}
	im := iterableMeta{
		sizeMethodIndex:    sizeMethod.Index,
		forEachMethodIndex: forEachMethod.Index,
		iterFuncType:       iterFuncType,
	}
	return im, true
}

func deRefPointer(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
