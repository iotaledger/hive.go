// Package serix serializes and deserializes complex Go objects into/from bytes using reflection.
/*

Structs serialization/deserialization

In order for a field to be detected by serix it must have `serix:""` struct tag.
The first part in the tag is the key used for json serialization.
If the name is empty, serix uses the field name in camel case (Exceptions: "ID" => "Id", "URL" => "Url").

Examples:
	- `serix:""
	- `serix:"example"
	- `serix:","`

serix traverses all fields and handles them in the order specified in the struct.
You can provide the following settings to serix via struct tags:

	- "optional": means the field might be nil. Only valid for pointers or interfaces.
				  It will be prepended with the serialized size of the field.
		`serix:"example,optional"`

	- "nest": handle embedded/anonymous field as a nested field
		`serix:"example,nest"`

	- "omitempty": omit the field in json serialization if it's empty
		`serix:"example,omitempty"`

	- "lenPrefix": provide serializer.SeriLengthPrefixType for that field (string, slice, map)
		`serix:"example,lenPrefix=uint32"`

	- "minLen": minimum length for that field (string, slice, map)
		`serix:"example,minLen=2"`

	- "maxLen": maximum length for that field (string, slice, map)
		`serix:"example,maxLen=5"`

	- "mapMaxByteSize": maximum serialized byte size for that map
		`serix:"example,mapMaxByteSize=100"`

	- "mapKeyLenPrefix": provide serializer.SeriLengthPrefixType for the keys of that map
		`serix:"example,mapKeyLenPrefix=uint32"`

	- "mapKeyMinLen": minimum length for the keys of that map
		`serix:"example,mapKeyMinLen=2"`

	- "mapKeyMaxLen": maximum length for the keys of that map
		`serix:"example,mapKeyMaxLen=5"`

	- "mapValueLenPrefix": provide serializer.SeriLengthPrefixType for the values of that map
		`serix:"example,mapValueLenPrefix=uint32"`

	- "mapValueMinLen": minimum length for the values of that map
		`serix:"example,mapValueMinLen=2"`

	- "mapValueMaxLen": maximum length for the values of that map
		`serix:"example,mapValueMaxLen=5"`


See serix_text.go for more detail.
*/
package serix

import (
	"context"
	"encoding/json"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/orderedmap"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrMapValidationMaxBytesExceeded gets returned if the serialized byte size of the map is too big.
	ErrMapValidationMaxBytesExceeded = ierrors.New("max bytes size of the map exceeded")
	// ErrMapValidationViolatesUniqueness gets returned if the map elements are not unique.
	ErrMapValidationViolatesUniqueness = ierrors.New("map elements must be unique")
)

var (
	bytesType     = reflect.TypeOf([]byte(nil))
	bigIntPtrType = reflect.TypeOf((*big.Int)(nil))
	timeType      = reflect.TypeOf(time.Time{})
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
	ctxType       = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// DefaultAPI is the default instance of the API type.
var DefaultAPI = NewAPI()

// Serializable is a type that can serialize itself.
// Serix will call its .Encode() method instead of trying to serialize it in the default way.
// The behavior is totally the same as in the standard "encoding/json" package and json.Marshaler interface.
type Serializable interface {
	Encode() ([]byte, error)
}

// Deserializable is a type that can deserialize itself.
// Serix will call its .Decode() method instead of trying to deserialize it in the default way.
// The behavior is totally the same as in the standard "encoding/json" package and json.Unmarshaler interface.
type Deserializable interface {
	Decode(b []byte) (int, error)
}

// ContextAwareDeserializable is a type that is able to receive the serialization context.
type ContextAwareDeserializable interface {
	SetDeserializationContext(ctx context.Context)
}

// SerializableJSON is a type that can serialize itself to JSON format.
// Serix will call its .EncodeJSON() method instead of trying to serialize it in the default way.
// The behavior is totally the same as in the standard "encoding/json" package and json.Marshaler interface.
type SerializableJSON interface {
	EncodeJSON() (any, error)
}

// DeserializableJSON is a type that can deserialize itself from JSON format.
// Serix will call its .Decode() method instead of trying to deserialize it in the default way.
// The behavior is totally the same as in the standard "encoding/json" package and json.Unmarshaler interface.
type DeserializableJSON interface {
	DecodeJSON(b any) error
}

type validators struct {
	bytesValidator     reflect.Value
	syntacticValidator reflect.Value
}

type interfaceObjects struct {
	fromCodeToType map[uint32]reflect.Type
	fromTypeToCode map[reflect.Type]uint32
	typeDenotation serializer.TypeDenotationType
}

// Option is an option for Encode/Decode methods.
type Option func(o *options)

// WithValidation returns an Option that tells serix to perform validation.
func WithValidation() Option {
	return func(o *options) {
		o.validation = true
	}
}

// WithTypeSettings returns an option that sets TypeSettings.
// TypeSettings provided via option override global TypeSettings from the registry.
// See TypeSettings for details.
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

// API is the main object of the package that provides the methods for client to use.
// It holds all the settings and configuration. It also stores the cache.
// Most often you will need a single object of API for the whole program.
// You register all type settings and interfaces on the program start or in init() function.
// Instead of creating a new API object you can also use the default singleton API object: DefaultAPI.
type API struct {
	interfacesRegistryMutex sync.RWMutex
	interfacesRegistry      map[reflect.Type]*interfaceObjects

	typeSettingsRegistryMutex sync.RWMutex
	typeSettingsRegistry      map[reflect.Type]TypeSettings

	validatorsRegistryMutex sync.RWMutex
	validatorsRegistry      map[reflect.Type]validators

	typeCacheMutex sync.RWMutex
	typeCache      map[reflect.Type][]structField
}

// NewAPI creates a new instance of the API type.
func NewAPI() *API {
	api := &API{
		interfacesRegistry:   map[reflect.Type]*interfaceObjects{},
		typeSettingsRegistry: map[reflect.Type]TypeSettings{},
		validatorsRegistry:   map[reflect.Type]validators{},
		typeCache:            map[reflect.Type][]structField{},
	}

	return api
}

// checks whether the given value has the concept of a length.
func hasLength(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Array:
	case reflect.Map:
	case reflect.Slice:
	case reflect.String:
	default:
		return false
	}

	return true
}

// checkMinMaxBoundsLength checks whether the given length is within its defined bounds.
func (api *API) checkMinMaxBoundsLength(length int, ts TypeSettings) error {
	if minLen, ok := ts.MinLen(); ok {
		if uint(length) < minLen {
			return ierrors.Wrapf(serializer.ErrArrayValidationMinElementsNotReached, "min length %d not reached (len %d)", minLen, length)
		}
	}
	if maxLen, ok := ts.MaxLen(); ok {
		if uint(length) > maxLen {
			return ierrors.Wrapf(serializer.ErrArrayValidationMaxElementsExceeded, "max length %d exceeded (len %d)", maxLen, length)
		}
	}

	return nil
}

// checkMinMaxBounds checks whether the given value is within its defined bounds in case it has a length.
func (api *API) checkMinMaxBounds(v reflect.Value, ts TypeSettings) error {
	if has := hasLength(v); !has {
		return nil
	}

	if err := api.checkMinMaxBoundsLength(v.Len(), ts); err != nil {
		return ierrors.Wrapf(err, "can't serialize '%s' type", v.Kind())
	}

	return nil
}

// checkMapMaxByteSize checks whether the given map is within its defined max byte size in case it has defined map rules.
func (api *API) checkMapMaxByteSize(byteSize int, ts TypeSettings) error {
	if ts.mapRules != nil && ts.mapRules.MaxByteSize > 0 && byteSize > int(ts.mapRules.MaxByteSize) {
		return ierrors.Wrapf(ErrMapValidationMaxBytesExceeded, "map (len %d) exceeds max bytes of %d ", byteSize, ts.mapRules.MaxByteSize)
	}

	return nil
}

func (api *API) checkMapSerializedSize(ctx context.Context, value reflect.Value, ts TypeSettings, opts *options) error {
	if ts.mapRules == nil || ts.mapRules.MaxByteSize == 0 {
		return nil
	}

	if value.Kind() != reflect.Map {
		return ierrors.Errorf("can't get map serialized size: value is not a map, got %s", value.Kind())
	}

	bytes, err := api.encode(ctx, value, ts, opts)
	if err != nil {
		return ierrors.Wrapf(err, "can't get map serialized size: failed to encode map")
	}

	return api.checkMapMaxByteSize(len(bytes), ts)
}

// Encode serializes the provided object obj into bytes.
// serix traverses the object recursively and serializes everything based on the type.
// If a type implements the custom Serializable interface serix delegates the serialization to that type.
// During the encoding process serix also performs the validation if such option was provided.
// Use the options list opts to customize the serialization behavior.
// To ensure deterministic serialization serix automatically applies lexical ordering for maps.
func (api *API) Encode(ctx context.Context, obj interface{}, opts ...Option) ([]byte, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, ierrors.New("invalid value for destination")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return api.encode(ctx, value, opt.ts, opt)
}

// JSONEncode serializes the provided object obj into its JSON representation.
func (api *API) JSONEncode(ctx context.Context, obj any, opts ...Option) ([]byte, error) {
	orderedMap, err := api.MapEncode(ctx, obj, opts...)
	if err != nil {
		return nil, err
	}

	return json.Marshal(orderedMap)
}

// MapEncode serializes the provided object obj into an ordered map.
// serix traverses the object recursively and serializes everything based on the type.
// Use the options list opts to customize the serialization behavior.
func (api *API) MapEncode(ctx context.Context, obj interface{}, opts ...Option) (*orderedmap.OrderedMap, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, ierrors.New("invalid value for destination")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	m, err := api.mapEncode(ctx, value, opt.ts, opt)
	if err != nil {
		return nil, err
	}

	mCasted, ok := m.(*orderedmap.OrderedMap)
	if !ok {
		return nil, ierrors.New("failed to cast to *orderedmap.OrderedMap")
	}

	return mCasted, nil
}

// Decode deserializes bytes b into the provided object obj.
// obj must be a non-nil pointer for serix to deserialize into it.
// serix traverses the object recursively and deserializes everything based on its type.
// If a type implements the custom Deserializable interface serix delegates the deserialization to that type.
// During the decoding process serix also performs the validation if such option was provided.
// Use the options list opts to customize the deserialization behavior.
func (api *API) Decode(ctx context.Context, b []byte, obj interface{}, opts ...Option) (int, error) {
	value := reflect.ValueOf(obj)
	if err := checkDecodeDestination(obj, value); err != nil {
		return 0, err
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return api.decode(ctx, b, value, opt.ts, opt)
}

// JSONDecode deserializes json data into the provided object obj.
func (api *API) JSONDecode(ctx context.Context, data []byte, obj interface{}, opts ...Option) error {
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return api.MapDecode(ctx, m, obj, opts...)
}

// MapDecode deserializes generic map m into the provided object obj.
// obj must be a non-nil pointer for serix to deserialize into it.
// serix traverses the object recursively and deserializes everything based on its type.
// Use the options list opts to customize the deserialization behavior.
func (api *API) MapDecode(ctx context.Context, m map[string]any, obj interface{}, opts ...Option) error {
	value := reflect.ValueOf(obj)
	if err := checkDecodeDestination(obj, value); err != nil {
		return err
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return api.mapDecode(ctx, m, value, opt.ts, opt)
}

func checkDecodeDestination(obj any, value reflect.Value) error {
	if !value.IsValid() {
		return ierrors.New("invalid value for destination")
	}
	if value.Kind() != reflect.Ptr {
		return ierrors.Errorf(
			"can't decode, the destination object must be a pointer, got: %T(%s)", obj, value.Kind(),
		)
	}
	if value.IsNil() {
		return ierrors.Errorf("can't decode, the destination object %T must be a non-nil pointer", obj)
	}

	return nil
}

// RegisterValidators registers validator functions that serix will call during the Encode and Decode processes.
// There are two types of validator functions:
//
// 1. Syntactic validators, they validate the Go object and its data.
// For Encode they are called for the original Go object before serix serializes the object into bytes.
// For Decode they are called after serix builds the Go object from bytes.
//
// 2. Bytes validators, they validate the corresponding bytes representation of an object.
// For Encode they are called after serix serializes Go object into bytes
// For Decode they are called for the bytes before serix deserializes them into a Go object.
//
// The validation is called for every registered type during the recursive traversal.
// It's an early stop process, if some validator returns an error serix stops the Encode/Decode and pops up the error.
//
// obj is an instance of the type you want to provide the validator for.
// Note that it's better to pass the obj as a value, not as a pointer
// because that way serix would be able to dereference pointers during Encode/Decode
// and detect the validators for both pointers and values
// bytesValidatorFn is a function that accepts context.Context, []byte and returns an error.
// syntacticValidatorFn is a function that accepts context.Context, and an object with the same type as obj.
// Every validator func is optional, just provide nil.
// Example:
// bytesValidator := func(ctx context.Context, b []byte) error { ... }
// syntacticValidator := func (ctx context.Context, t time.Time) error { ... }
// api.RegisterValidators(time.Time{}, bytesValidator, syntacticValidator)
//
// See TestMain() in serix_test.go for more examples.
func (api *API) RegisterValidators(obj any, bytesValidatorFn func(context.Context, []byte) error, syntacticValidatorFn interface{}) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return ierrors.New("'obj' is a nil interface, it needs to be a valid type")
	}
	bytesValidatorValue, err := parseValidatorFunc(bytesValidatorFn)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse bytesValidatorFn")
	}
	syntacticValidatorValue, err := parseValidatorFunc(syntacticValidatorFn)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse syntacticValidatorFn")
	}
	vldtrs := validators{}
	if bytesValidatorValue.IsValid() {
		if err := checkBytesValidatorSignature(bytesValidatorValue); err != nil {
			return ierrors.WithStack(err)
		}
		vldtrs.bytesValidator = bytesValidatorValue
	}
	if syntacticValidatorValue.IsValid() {
		if err := checkSyntacticValidatorSignature(objType, syntacticValidatorValue); err != nil {
			return ierrors.WithStack(err)
		}
		vldtrs.syntacticValidator = syntacticValidatorValue
	}
	api.validatorsRegistryMutex.Lock()
	defer api.validatorsRegistryMutex.Unlock()
	api.validatorsRegistry[objType] = vldtrs

	return nil
}

func parseValidatorFunc(validatorFn interface{}) (reflect.Value, error) {
	if validatorFn == nil {
		return reflect.Value{}, nil
	}
	funcValue := reflect.ValueOf(validatorFn)
	if !funcValue.IsValid() || funcValue.IsZero() {
		return reflect.Value{}, nil
	}
	if funcValue.Kind() != reflect.Func {
		return reflect.Value{}, ierrors.Errorf(
			"validator must be a function, got %T(%s)", validatorFn, funcValue.Kind(),
		)
	}
	funcType := funcValue.Type()
	if funcType.NumIn() != 2 {
		return reflect.Value{}, ierrors.New("validator func must have two arguments")
	}
	firstArgType := funcType.In(0)
	if firstArgType != ctxType {
		return reflect.Value{}, ierrors.New("validator func's first argument must be context")
	}
	if funcType.NumOut() != 1 {
		return reflect.Value{}, ierrors.Errorf("validator func must have one return value, got %d", funcType.NumOut())
	}
	returnType := funcType.Out(0)
	if returnType != errorType {
		return reflect.Value{}, ierrors.Errorf("validator func must have 'error' return type, got %s", returnType)
	}

	return funcValue, nil
}

func checkBytesValidatorSignature(funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != bytesType {
		return ierrors.Errorf("bytesValidatorFn's argument must be bytes, got %s", argumentType)
	}

	return nil
}

func checkSyntacticValidatorSignature(objectType reflect.Type, funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != objectType {
		return ierrors.Errorf(
			"syntacticValidatorFn's argument must have the same type as the object it was registered for, "+
				"objectType=%s, argumentType=%s",
			objectType, argumentType,
		)
	}

	return nil
}

func (api *API) callBytesValidator(ctx context.Context, valueType reflect.Type, bytes []byte) error {
	api.validatorsRegistryMutex.RLock()
	defer api.validatorsRegistryMutex.RUnlock()
	bytesValidator := api.validatorsRegistry[valueType].bytesValidator
	if !bytesValidator.IsValid() {
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			bytesValidator = api.validatorsRegistry[valueType].bytesValidator
		}
	}
	if bytesValidator.IsValid() {
		if err, _ := bytesValidator.Call(
			[]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(bytes)},
		)[0].Interface().(error); err != nil {
			return ierrors.Wrapf(err, "bytes validator returns an error for type %s", valueType)
		}
	}

	return nil
}

func (api *API) callSyntacticValidator(ctx context.Context, value reflect.Value, valueType reflect.Type) error {
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
		if err, _ := syntacticValidator.Call(
			[]reflect.Value{reflect.ValueOf(ctx), value},
		)[0].Interface().(error); err != nil {
			return ierrors.Wrapf(err, "syntactic validator returns an error for type %s", valueType)
		}
	}

	return nil
}

// RegisterTypeSettings registers settings for a particular type obj.
// It's better to provide obj as a value, not a pointer,
// that way serix will be able to get the type settings for both values and pointers during Encode/Decode via de-referencing
// The settings provided via registration are considered global and default ones,
// they can be overridden by type settings parsed from struct tags
// or by type settings provided via option to the Encode/Decode methods.
// See TypeSettings for more detail.
func (api *API) RegisterTypeSettings(obj interface{}, ts TypeSettings) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return ierrors.New("'obj' is a nil interface, it's need to be a valid type")
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

// RegisterInterfaceObjects tells serix that when it encounters iType during serialization/deserialization
// it actually might be one of the objs types.
// Those objs type must provide their ObjectTypes beforehand via API.RegisterTypeSettings().
// serix needs object types to be able to figure out what concrete object to instantiate during the deserialization
// based on its object type code.
// In order for reflection to grasp the actual interface type, iType must be provided as a pointer to an interface:
// api.RegisterInterfaceObjects((*Interface)(nil), (*InterfaceImpl)(nil))
// See TestMain() in serix_test.go for more detail.
func (api *API) RegisterInterfaceObjects(iType interface{}, objs ...interface{}) error {
	ptrType := reflect.TypeOf(iType)
	if ptrType == nil {
		return ierrors.New("'iType' is a nil interface, it needs to be a pointer to an interface")
	}
	if ptrType.Kind() != reflect.Ptr {
		return ierrors.Errorf("'iType' parameter must be a pointer, got %s", ptrType.Kind())
	}
	iTypeReflect := ptrType.Elem()
	if iTypeReflect.Kind() != reflect.Interface {
		return ierrors.Errorf(
			"'iType' pointer must contain an interface, got %s", iTypeReflect.Kind())
	}
	if len(objs) == 0 {
		return nil
	}

	iRegistry, exists := api.interfacesRegistry[iTypeReflect]
	if !exists {
		iRegistry = &interfaceObjects{
			fromCodeToType: make(map[uint32]reflect.Type, len(objs)),
			fromTypeToCode: make(map[reflect.Type]uint32, len(objs)),
		}
	}

	for i, obj := range objs {
		objType := reflect.TypeOf(obj)
		objTypeDenotation, objCode, err := api.getTypeDenotationAndObjectCode(objType)
		if err != nil {
			return ierrors.Wrapf(err, "failed to get type denotation for object %T", obj)
		}
		if i == 0 {
			iRegistry.typeDenotation = objTypeDenotation
		} else if iRegistry.typeDenotation != objTypeDenotation {
			firstObj := objs[0]

			return ierrors.Errorf(
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

func (api *API) getTypeDenotationAndObjectCode(objType reflect.Type) (serializer.TypeDenotationType, uint32, error) {
	ts, exists := api.getTypeSettings(objType)
	if !exists {
		return 0, 0, ierrors.Errorf(
			"no type settings was found for object %s"+
				"you must register object with its type settings first",
			objType,
		)
	}
	objectType := ts.ObjectType()
	if objectType == nil {
		return 0, 0, ierrors.Errorf(
			"type settings for object %s doesn't contain object code",
			objType,
		)
	}
	objTypeDenotation, objectCode, err := getTypeDenotationAndCode(objectType)
	if err != nil {
		return 0, 0, ierrors.WithStack(err)
	}

	return objTypeDenotation, objectCode, nil
}

func getTypeDenotationAndCode(objectType interface{}) (serializer.TypeDenotationType, uint32, error) {
	objCodeType := reflect.TypeOf(objectType)
	if objCodeType == nil {
		return 0, 0, ierrors.New("can't detect type denotation type: object code is nil interface")
	}
	var code uint32
	var objTypeDenotation serializer.TypeDenotationType
	switch objCodeType.Kind() {
	case reflect.Uint32:
		objTypeDenotation = serializer.TypeDenotationUint32

		//nolint:forcetypeassert // false positive, we already checked the type via reflect.Kind
		code = objectType.(uint32)
	case reflect.Uint8:
		objTypeDenotation = serializer.TypeDenotationByte

		//nolint:forcetypeassert // false positive, we already checked the type via reflect.Kind
		code = uint32(objectType.(uint8))
	default:
		return 0, 0, ierrors.Errorf("unsupported object code type: %s (%s), only uint32 and byte are supported",
			objCodeType, objCodeType.Kind())
	}

	return objTypeDenotation, code, nil
}

func (api *API) getInterfaceObjects(iType reflect.Type) *interfaceObjects {
	api.interfacesRegistryMutex.RLock()
	defer api.interfacesRegistryMutex.RUnlock()

	return api.interfacesRegistry[iType]
}

type structField struct {
	name         string
	isUnexported bool
	index        int
	fType        reflect.Type
	isEmbedded   bool
	settings     tagSettings
}

type tagSettings struct {
	position   int
	isOptional bool
	nest       bool
	omitEmpty  bool
	ts         TypeSettings
}

func (api *API) parseStructType(structType reflect.Type) ([]structField, error) {
	api.typeCacheMutex.RLock()
	structFields, exists := api.typeCache[structType]
	api.typeCacheMutex.RUnlock()
	if exists {
		return structFields, nil
	}

	structFields = make([]structField, 0, structType.NumField())

	serixPosition := 0
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		isUnexported := field.PkgPath != ""
		isEmbedded := field.Anonymous
		isStruct := isUnderlyingStruct(field.Type)
		isInterface := isUnderlyingInterface(field.Type)
		isEmbeddedStruct := isEmbedded && isStruct
		isEmbeddedInterface := isEmbedded && isInterface

		if isUnexported && !isEmbeddedStruct && !isEmbeddedInterface {
			continue
		}

		tag, ok := field.Tag.Lookup("serix")
		if !ok {
			continue
		}

		tSettings, err := parseSerixSettings(tag, serixPosition)
		if err != nil {
			return nil, ierrors.Wrapf(err, "failed to parse serix struct tag for field %s", field.Name)
		}
		serixPosition++

		if tSettings.isOptional {
			if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: "+
						"'optional' setting can only be used with pointers or interfaces, got %s",
					field.Name, field.Type.Kind())
			}

			if isEmbeddedStruct {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: 'optional' setting can't be used with embedded structs",
					field.Name)
			}

			if isEmbeddedInterface {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: 'optional' setting can't be used with embedded interfaces",
					field.Name)
			}
		}

		if tSettings.nest && isUnexported {
			return nil, ierrors.Errorf(
				"struct field %s is invalid: 'nest' setting can't be used with unexported types",
				field.Name)
		}

		if !tSettings.nest && isEmbeddedInterface {
			return nil, ierrors.Errorf(
				"struct field %s is invalid: 'nest' setting needs to be used for embedded interfaces",
				field.Name)
		}

		structFields = append(structFields, structField{
			name:         field.Name,
			isUnexported: isUnexported,
			index:        i,
			fType:        field.Type,
			isEmbedded:   isEmbeddedStruct || isEmbeddedInterface,
			settings:     tSettings,
		})
	}
	sort.Slice(structFields, func(i, j int) bool {
		return structFields[i].settings.position < structFields[j].settings.position
	})

	api.typeCacheMutex.Lock()
	defer api.typeCacheMutex.Unlock()

	api.typeCache[structType] = structFields

	return structFields, nil
}

func parseStructTagValue(name string, keyValue []string, currentPart string) (string, error) {
	if len(keyValue) != 2 {
		return "", ierrors.Errorf("incorrect %s tag format: %s", name, currentPart)
	}

	return keyValue[1], nil
}

func parseStructTagValueUint(name string, keyValue []string, currentPart string) (uint, error) {
	value, err := parseStructTagValue(name, keyValue, currentPart)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to parse %s %s", name, currentPart)
	}

	return uint(result), nil
}

func parseLengthPrefixType(prefixTypeRaw string) (LengthPrefixType, error) {
	switch prefixTypeRaw {
	case "byte", "uint8":
		return LengthPrefixTypeAsByte, nil
	case "uint16":
		return LengthPrefixTypeAsUint16, nil
	case "uint32":
		return LengthPrefixTypeAsUint32, nil
	default:
		return LengthPrefixTypeAsByte, ierrors.Wrapf(ErrUnknownLengthPrefixType, "%s", prefixTypeRaw)
	}
}

func parseStructTagValuePrefixType(name string, keyValue []string, currentPart string) (LengthPrefixType, error) {
	value, err := parseStructTagValue(name, keyValue, currentPart)
	if err != nil {
		return 0, err
	}

	lengthPrefixType, err := parseLengthPrefixType(value)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to parse %s %s", name, currentPart)
	}

	return lengthPrefixType, nil
}

func parseSerixSettings(tag string, serixPosition int) (tagSettings, error) {
	settings := tagSettings{}
	settings.position = serixPosition

	if tag == "" {
		// empty struct tags are allowed
		return settings, nil
	}

	parts := strings.Split(tag, ",")
	keyPart := parts[0]

	if strings.ContainsAny(keyPart, "=") {
		return tagSettings{}, ierrors.Errorf("incorrect struct tag format: %s, must start with the field key or \",\"", tag)
	}

	settings.ts = settings.ts.WithFieldKey(keyPart)

	parts = parts[1:]
	seenParts := map[string]struct{}{}
	for _, currentPart := range parts {
		if _, ok := seenParts[currentPart]; ok {
			return tagSettings{}, ierrors.Errorf("duplicated tag part: %s", currentPart)
		}
		keyValue := strings.Split(currentPart, "=")
		partName := keyValue[0]

		switch partName {
		case "optional":
			settings.isOptional = true

		case "nest":
			settings.nest = true

		case "omitempty":
			settings.omitEmpty = true

		case "lenPrefix":
			value, err := parseStructTagValuePrefixType("lenPrefix", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithLengthPrefixType(value)

		case "minLen":
			value, err := parseStructTagValueUint("minLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMinLen(value)

		case "maxLen":
			value, err := parseStructTagValueUint("maxLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMaxLen(value)

		case "mapMaxByteSize":
			value, err := parseStructTagValueUint("mapMaxByteSize", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapMaxByteSize(value)

		case "mapKeyLenPrefix":
			value, err := parseStructTagValuePrefixType("mapKeyLenPrefix", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapKeyLengthPrefixType(value)

		case "mapKeyMinLen":
			value, err := parseStructTagValueUint("mapKeyMinLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapKeyMinLen(value)

		case "mapKeyMaxLen":
			value, err := parseStructTagValueUint("mapKeyMaxLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapKeyMaxLen(value)

		case "mapValueLenPrefix":
			value, err := parseStructTagValuePrefixType("mapValueLenPrefix", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapValueLengthPrefixType(value)

		case "mapValueMinLen":
			value, err := parseStructTagValueUint("mapValueMinLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapValueMinLen(value)

		case "mapValueMaxLen":
			value, err := parseStructTagValueUint("mapValueMaxLen", keyValue, currentPart)
			if err != nil {
				return tagSettings{}, err
			}
			settings.ts = settings.ts.WithMapValueMaxLen(value)

		default:
			return tagSettings{}, ierrors.Errorf("unknown tag part: %s", currentPart)
		}

		seenParts[partName] = struct{}{}
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

func fillArrayFromSlice(arrayValue, sliceValue reflect.Value) {
	for i := 0; i < sliceValue.Len(); i++ {
		arrayValue.Index(i).Set(sliceValue.Index(i))
	}
}

func isUnderlyingStruct(t reflect.Type) bool {
	t = deRefPointer(t)

	return t.Kind() == reflect.Struct
}

func isUnderlyingInterface(t reflect.Type) bool {
	t = deRefPointer(t)

	return t.Kind() == reflect.Interface
}

func deRefPointer(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func getNumberTypeToConvert(kind reflect.Kind) (int, reflect.Type, reflect.Type) {
	var numberType reflect.Type
	var bitSize int
	switch kind {
	case reflect.Int8:
		numberType = reflect.TypeOf(int8(0))
		bitSize = 8
	case reflect.Int16:
		numberType = reflect.TypeOf(int16(0))
		bitSize = 16
	case reflect.Int32:
		numberType = reflect.TypeOf(int32(0))
		bitSize = 32
	case reflect.Int64:
		numberType = reflect.TypeOf(int64(0))
		bitSize = 64
	case reflect.Uint8:
		numberType = reflect.TypeOf(uint8(0))
		bitSize = 8
	case reflect.Uint16:
		numberType = reflect.TypeOf(uint16(0))
		bitSize = 16
	case reflect.Uint32:
		numberType = reflect.TypeOf(uint32(0))
		bitSize = 32
	case reflect.Uint64:
		numberType = reflect.TypeOf(uint64(0))
		bitSize = 64
	case reflect.Float32:
		numberType = reflect.TypeOf(float32(0))
		bitSize = 32
	case reflect.Float64:
		numberType = reflect.TypeOf(float64(0))
		bitSize = 64
	default:
		return -1, nil, nil
	}

	return bitSize, numberType, reflect.PointerTo(numberType)
}

// FieldKeyString converts the given string to camelCase.
// Special keywords like ID or URL are converted to only first letter upper case.
func FieldKeyString(str string) string {
	for _, keyword := range []string{"ID", "URL"} {
		if !strings.Contains(str, keyword) {
			continue
		}

		// first keyword letter upper case, rest lower case
		str = strings.ReplaceAll(str, keyword, string(keyword[0])+strings.ToLower(keyword)[1:])
	}

	// first letter lower case
	return strings.ToLower(str[:1]) + str[1:]
}
