// Package serix serializes and deserializes complex Go objects into/from bytes using reflection.
/*

Structs serialization/deserialization

In order for a field to be detected by serix it must have `serix` struct tag set with the position index: `serix:"0"`.
serix traverses all fields and handles them in the order specified in the struct tags.
Apart from the required position you can provide the following settings to serix via struct tags:
"optional" - means that field might be nil. Only valid for pointers or interfaces: `serix:"1,optional"`
"lengthPrefixType=uint32" - provide serializer.SeriLengthPrefixType for that field: `serix:"2,lengthPrefixType=unint32"`
"nest" - handle embedded/anonymous field as a nested field: `serix:"3,nest"`
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
	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/serializer/v2"
)

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

type validators struct {
	bytesValidator     reflect.Value
	syntacticValidator reflect.Value
}

type interfaceObjects struct {
	fromCodeToType map[uint32]reflect.Type
	fromTypeToCode map[reflect.Type]uint32
	typeDenotation serializer.TypeDenotationType
}

// DefaultAPI is the default instance of the API type.
var DefaultAPI = NewAPI()

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

var (
	bytesType     = reflect.TypeOf([]byte(nil))
	bigIntPtrType = reflect.TypeOf((*big.Int)(nil))
	timeType      = reflect.TypeOf(time.Time{})
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
	ctxType       = reflect.TypeOf((*context.Context)(nil)).Elem()
)

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

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules serializer.ArrayRules

// LengthPrefixType defines the type of the value denoting the length of a collection.
type LengthPrefixType serializer.SeriLengthPrefixType

const (
	// LengthPrefixTypeAsByte defines a collection length to be denoted by a byte.
	LengthPrefixTypeAsByte = LengthPrefixType(serializer.SeriLengthPrefixTypeAsByte)
	// LengthPrefixTypeAsUint16 defines a collection length to be denoted by a uint16.
	LengthPrefixTypeAsUint16 = LengthPrefixType(serializer.SeriLengthPrefixTypeAsUint16)
	// LengthPrefixTypeAsUint32 defines a collection length to be denoted by a uint32.
	LengthPrefixTypeAsUint32 = LengthPrefixType(serializer.SeriLengthPrefixTypeAsUint32)
)

// TypeSettings holds various settings for a particular type.
// Those settings determine how the object should be serialized/deserialized.
// There are three way to provide TypeSettings
// 1. Via global registry: API.RegisterTypeSettings().
// 2. Parse from struct tags.
// 3. Pass as an option to API.Encode/API.Decode methods.
// The type settings provided via struct tags or an option override the type settings from the registry.
// So the precedence is the following 1<2<3.
// See API.RegisterTypeSettings() and WithTypeSettings() for more detail.
type TypeSettings struct {
	lengthPrefixType *LengthPrefixType
	objectType       interface{}
	lexicalOrdering  *bool
	mapKey           *string
	arrayRules       *ArrayRules
}

// WithLengthPrefixType specifies LengthPrefixType.
func (ts TypeSettings) WithLengthPrefixType(lpt LengthPrefixType) TypeSettings {
	ts.lengthPrefixType = &lpt

	return ts
}

// LengthPrefixType returns LengthPrefixType.
func (ts TypeSettings) LengthPrefixType() (LengthPrefixType, bool) {
	if ts.lengthPrefixType == nil {
		return 0, false
	}

	return *ts.lengthPrefixType, true
}

// WithMapKey specifies the name for the map key.
func (ts TypeSettings) WithMapKey(name string) TypeSettings {
	ts.mapKey = &name

	return ts
}

// MapKey returns the map key name.
func (ts TypeSettings) MapKey() (string, bool) {
	if ts.mapKey == nil {
		return "", false
	}

	return *ts.mapKey, true
}

// MustMapKey must return a map key name.
func (ts TypeSettings) MustMapKey() string {
	if ts.mapKey == nil {
		panic("no map key set")
	}

	return *ts.mapKey
}

// WithMinLen specifies the min length for the object.
func (ts TypeSettings) WithMinLen(l uint) TypeSettings {
	if ts.arrayRules == nil {
		ts.arrayRules = new(ArrayRules)
	}
	ts.arrayRules.Min = l

	return ts
}

// MinLen returns min length for the object.
func (ts TypeSettings) MinLen() (uint, bool) {
	if ts.arrayRules == nil || ts.arrayRules.Min == 0 {
		return 0, false
	}

	return ts.arrayRules.Min, true
}

// WithMaxLen specifies the max length for the object.
func (ts TypeSettings) WithMaxLen(l uint) TypeSettings {
	if ts.arrayRules == nil {
		ts.arrayRules = new(ArrayRules)
	}
	ts.arrayRules.Max = l

	return ts
}

// MaxLen returns max length for the object.
func (ts TypeSettings) MaxLen() (uint, bool) {
	if ts.arrayRules == nil || ts.arrayRules.Max == 0 {
		return 0, false
	}

	return ts.arrayRules.Max, true
}

// MinMaxLen returns min/max lengths for the object.
// Returns 0 for either value if they are not set.
func (ts TypeSettings) MinMaxLen() (int, int) {
	var min, max int
	if ts.arrayRules != nil {
		min = int(ts.arrayRules.Min)
	}
	if ts.arrayRules != nil {
		max = int(ts.arrayRules.Max)
	}

	return min, max
}

// WithObjectType specifies the object type. It can be either uint8 or uint32 number.
// The object type holds two meanings: the actual code (number) and the serializer.TypeDenotationType like uint8 or uint32.
// serix uses object type to actually encode the number
// and to know its serializer.TypeDenotationType to be able to decode it.
func (ts TypeSettings) WithObjectType(t interface{}) TypeSettings {
	ts.objectType = t

	return ts
}

// ObjectType returns the object type as an uint8 or uint32 number.
func (ts TypeSettings) ObjectType() interface{} {
	return ts.objectType
}

// WithLexicalOrdering specifies whether the type must be lexically ordered during serialization.
func (ts TypeSettings) WithLexicalOrdering(val bool) TypeSettings {
	ts.lexicalOrdering = &val

	return ts
}

// LexicalOrdering returns lexical ordering flag.
func (ts TypeSettings) LexicalOrdering() (val bool, set bool) {
	if ts.lexicalOrdering == nil {
		return false, false
	}

	return *ts.lexicalOrdering, true
}

// WithArrayRules specifies serializer.ArrayRules.
func (ts TypeSettings) WithArrayRules(rules *ArrayRules) TypeSettings {
	ts.arrayRules = rules

	return ts
}

// ArrayRules returns serializer.ArrayRules.
func (ts TypeSettings) ArrayRules() *ArrayRules {
	return ts.arrayRules
}

func (ts TypeSettings) ensureOrdering() TypeSettings {
	newTS := ts.WithLexicalOrdering(true)
	arrayRules := newTS.ArrayRules()
	newArrayRules := new(ArrayRules)
	if arrayRules != nil {
		*newArrayRules = *arrayRules
	}
	newArrayRules.ValidationMode |= serializer.ArrayValidationModeLexicalOrdering

	return newTS.WithArrayRules(newArrayRules)
}

func (ts TypeSettings) merge(other TypeSettings) TypeSettings {
	if ts.lengthPrefixType == nil {
		ts.lengthPrefixType = other.lengthPrefixType
	}
	if ts.objectType == nil {
		ts.objectType = other.objectType
	}
	if ts.lexicalOrdering == nil {
		ts.lexicalOrdering = other.lexicalOrdering
	}
	if ts.arrayRules == nil {
		ts.arrayRules = other.arrayRules
	}
	if ts.mapKey == nil {
		ts.mapKey = other.mapKey
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

// Encode serializes the provided object obj into bytes.
// serix traverses the object recursively and serializes everything based on the type.
// If a type implements the custom Serializable interface serix delegates the serialization to that type.
// During the encoding process serix also performs the validation if such option was provided.
// Use the options list opts to customize the serialization behavior.
// To ensure deterministic serialization serix automatically applies lexical ordering for maps.
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
		return nil, errors.New("invalid value for destination")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	m, err := api.mapEncode(ctx, value, opt.ts, opt)
	if err != nil {
		return nil, err
	}

	return m.(*orderedmap.OrderedMap), nil
}

// Decode deserializes bytes b into the provided object obj.
// obj must be a non-nil pointer for serix to deserialize into it.
// serix traverses the object recursively and deserializes everything based on its type.
// If a type implements the custom Deserializable interface serix delegates the deserialization to that type.
// During the decoding process serix also performs the validation if such option was provided.
// Use the options list opts to customize the deserialization behavior.
func (api *API) Decode(ctx context.Context, b []byte, obj interface{}, opts ...Option) (int, error) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return 0, errors.New("invalid value for destination")
	}
	if value.Kind() != reflect.Ptr {
		return 0, errors.Errorf(
			"can't decode, the destination object must be a pointer, got: %T(%s)", obj, value.Kind(),
		)
	}
	if value.IsNil() {
		return 0, errors.Errorf("can't decode, the destination object %T must be a non-nil pointer", obj)
	}
	value = value.Elem()
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
		return errors.New("'obj' is a nil interface, it needs to be a valid type")
	}
	bytesValidatorValue, err := parseValidatorFunc(bytesValidatorFn)
	if err != nil {
		return errors.Wrapf(err, "failed to parse bytesValidatorFn")
	}
	syntacticValidatorValue, err := parseValidatorFunc(syntacticValidatorFn)
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
		if err := checkSyntacticValidatorSignature(objType, syntacticValidatorValue); err != nil {
			return errors.WithStack(err)
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
		return reflect.Value{}, errors.Errorf(
			"validator must be a function, got %T(%s)", validatorFn, funcValue.Kind(),
		)
	}
	funcType := funcValue.Type()
	if funcType.NumIn() != 2 {
		return reflect.Value{}, errors.Errorf("validator func must have two arguments")
	}
	firstArgType := funcType.In(0)
	if firstArgType != ctxType {
		return reflect.Value{}, errors.Errorf("validator func's first argument must be context")
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
	argumentType := funcType.In(1)
	if argumentType != bytesType {
		return errors.Errorf("bytesValidatorFn's argument must be bytes, got %s", argumentType)
	}

	return nil
}

func checkSyntacticValidatorSignature(objectType reflect.Type, funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != objectType {
		return errors.Errorf(
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
			return errors.Wrapf(err, "bytes validator returns an error for type %s", valueType)
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
			return errors.Wrapf(err, "syntactic validator returns an error for type %s", valueType)
		}
	}

	return nil
}

// RegisterTypeSettings registers settings for a particular type obj.
// It's better to provide obj as a value, not a pointer,
// that way serix will be able to get the type settings for both values and pointers during Encode/Decode via dereferencing
// The settings provided via registration are considered global and default ones,
// they can be overridden by type settings parsed from struct tags
// or by type settings provided via option to the Encode/Decode methods.
// See TypeSettings for more detail.
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
		return errors.New("'iType' is a nil interface, it needs to be a pointer to an interface")
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
			fromCodeToType: make(map[uint32]reflect.Type, len(objs)),
			fromTypeToCode: make(map[reflect.Type]uint32, len(objs)),
		}
	}

	for i, obj := range objs {
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

func (api *API) getTypeDenotationAndObjectCode(objType reflect.Type) (serializer.TypeDenotationType, uint32, error) {
	ts, exists := api.getTypeSettings(objType)
	if !exists {
		return 0, 0, errors.Errorf(
			"no type settings was found for object %s"+
				"you must register object with its type settings first",
			objType,
		)
	}
	objectType := ts.ObjectType()
	if objectType == nil {
		return 0, 0, errors.Errorf(
			"type settings for object %s doesn't contain object code",
			objType,
		)
	}
	objTypeDenotation, objectCode, err := getTypeDenotationAndCode(objectType)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return objTypeDenotation, objectCode, nil
}

func getTypeDenotationAndCode(objectType interface{}) (serializer.TypeDenotationType, uint32, error) {
	objCodeType := reflect.TypeOf(objectType)
	if objCodeType == nil {
		return 0, 0, errors.New("can't detect type denotation type: object code is nil interface")
	}
	var code uint32
	var objTypeDenotation serializer.TypeDenotationType
	switch objCodeType.Kind() {
	case reflect.Uint32:
		objTypeDenotation = serializer.TypeDenotationUint32
		code = objectType.(uint32)
	case reflect.Uint8:
		objTypeDenotation = serializer.TypeDenotationByte
		code = uint32(objectType.(uint8))
	default:
		return 0, 0, errors.Errorf("unsupported object code type: %s (%s), only uint32 and byte are supported",
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
	name             string
	isUnexported     bool
	index            int
	fType            reflect.Type
	isEmbeddedStruct bool
	settings         tagSettings
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
			return nil, errors.Errorf("struct field with duplicated position number %d", tSettings.position)
		}
		seenPositions[tSettings.position] = struct{}{}
		if tSettings.isOptional {
			if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
				return nil, errors.Errorf(
					"struct field %s is invalid: "+
						"'optional' setting can only be used with pointers or interfaces, got %s",
					field.Name, field.Type.Kind())
			}
			if isEmbeddedStruct {
				return nil, errors.Errorf(
					"struct field %s is invalid: 'optional' setting can't be used with embedded structs",
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
			isUnexported:     isUnexported,
			index:            i,
			fType:            field.Type,
			isEmbeddedStruct: isEmbeddedStruct,
			settings:         tSettings,
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
		case "optional":
			settings.isOptional = true
		case "nest":
			settings.nest = true
		case "omitempty":
			settings.omitEmpty = true
		case "mapKey":
			if len(keyValue) != 2 {
				return tagSettings{}, errors.Errorf("incorrect mapKey tag format: %s", currentPart)
			}
			settings.ts = settings.ts.WithMapKey(keyValue[1])
		case "minLen":
			if len(keyValue) != 2 {
				return tagSettings{}, errors.Errorf("incorrect minLen tag format: %s", currentPart)
			}
			minLen, err := strconv.ParseUint(keyValue[1], 10, 64)
			if err != nil {
				return tagSettings{}, errors.Wrapf(err, "failed to parse minLen %s", currentPart)
			}
			settings.ts = settings.ts.WithMinLen(uint(minLen))
		case "maxLen":
			if len(keyValue) != 2 {
				return tagSettings{}, errors.Errorf("incorrect maxLen tag format: %s", currentPart)
			}
			maxLen, err := strconv.ParseUint(keyValue[1], 10, 64)
			if err != nil {
				return tagSettings{}, errors.Wrapf(err, "failed to parse maxLen %s", currentPart)
			}
			settings.ts = settings.ts.WithMaxLen(uint(maxLen))
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

func parseLengthPrefixType(prefixTypeRaw string) (LengthPrefixType, error) {
	switch prefixTypeRaw {
	case "byte", "uint8":
		return LengthPrefixTypeAsByte, nil
	case "uint16":
		return LengthPrefixTypeAsUint16, nil
	case "uint32":
		return LengthPrefixTypeAsUint32, nil
	default:
		return LengthPrefixTypeAsByte, errors.Errorf("unknown length prefix type: %s", prefixTypeRaw)
	}
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

func mapStringKey(str string) string {
	return strings.ToLower(str[:1]) + str[1:]
}
