// Package serix serializes and deserializes complex Go objects into/from bytes using reflection.
/*

Structs serialization/deserialization

In order for a field to be detected by serix it must have `serix:""` struct tag.
The first part in the tag is the key used for json serialization.
If the name is empty, serix uses the field name in camel case.
	Exceptions:
		- "ID" => "Id"
		- "NFT" => "Nft"
		- "URL" => "Url"
		- "HRP" => "Hrp"

Examples:
	- `serix:""
	- `serix:"example"
	- `serix:","`

serix traverses all fields and handles them in the order specified in the struct.
You can provide the following settings to serix via struct tags:

	- "optional": means the field might be nil. Only valid for pointers or interfaces.
				  It will be prepended with the serialized size of the field.
		`serix:"example,optional"`

	- "inlined": handle embedded/anonymous field as a nested field
		`serix:"example,inlined"`

	- "omitempty": omit the field in json serialization if it's empty
		`serix:"example,omitempty"`

	- "maxByteSize": maximum serialized byte size for that field
		`serix:"example,maxByteSize=100"`

	- "lenPrefix": provide serializer.SeriLengthPrefixType for that field (string, slice, map)
		`serix:"example,lenPrefix=uint32"`

	- "minLen": minimum length for that field (string, slice, map)
		`serix:"example,minLen=2"`

	- "maxLen": maximum length for that field (string, slice, map)
		`serix:"example,maxLen=5"`

See serix_text.go for more detail.
*/
package serix

import (
	"context"
	"encoding/json"
	"math/big"
	"reflect"
	"time"

	// we need to use this orderedmap implementation for serialization instead of our own,
	// because the generic orderedmap in hive.go doesn't support marshaling to json.
	// this orderedmap implementation uses map[string]any as underlying datastructure,
	// which is a must instead of map[K]V, otherwise we can't correctly sort nested maps during unmarshaling.
	"github.com/iancoleman/orderedmap"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrMapValidationViolatesUniqueness gets returned if the map elements are not unique.
	ErrMapValidationViolatesUniqueness = ierrors.New("map elements must be unique")
	// ErrNonUTF8String gets returned when a non UTF-8 string is being encoded/decoded.
	ErrNonUTF8String = ierrors.New("non UTF-8 string value")
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
	// the registered interfaces and their known objects
	interfacesRegistry *InterfacesRegistry

	// the registered type settings for the known objects
	typeSettingsRegistry *TypeSettingsRegistry

	// the registered validators for the known objects
	validatorsRegistry *validatorsRegistry

	// the cache for the struct fields
	structFieldsCache *structFieldsCache
}

// NewAPI creates a new instance of the API type.
func NewAPI() *API {
	api := &API{
		interfacesRegistry:   NewInterfacesRegistry(),
		typeSettingsRegistry: NewTypeSettingsRegistry(),
		validatorsRegistry:   newValidatorsRegistry(),
		structFieldsCache:    newStructFieldsCache(),
	}

	return api
}

func (api *API) getInterfaceObjects(iType reflect.Type) *InterfaceObjects {
	iObj, exists := api.interfacesRegistry.Get(iType)
	if !exists {
		return nil
	}

	return iObj
}

// Checks the "Must Occur" array rules in the given slice.
func (api *API) checkArrayMustOccur(slice reflect.Value, ts TypeSettings) error {
	if slice.Kind() != reflect.Slice {
		return ierrors.Errorf("must occur can only be checked for a slice, got value of kind %v", slice.Kind())
	}

	if ts.arrayRules == nil || len(ts.arrayRules.MustOccur) == 0 {
		return nil
	}

	mustOccurPrefixes := make(serializer.TypePrefixes, len(ts.arrayRules.MustOccur))
	for key, value := range ts.arrayRules.MustOccur {
		mustOccurPrefixes[key] = value
	}

	sliceLen := slice.Len()
	for i := range sliceLen {
		elemValue := slice.Index(i)

		// Get the type prefix of the element by retrieving the type settings.
		if elemValue.Kind() == reflect.Ptr || elemValue.Kind() == reflect.Interface {
			elemValue = reflect.Indirect(elemValue.Elem())
		}

		elemTypeSettings, exists := api.typeSettingsRegistry.GetByType(elemValue.Type())
		if !exists {
			return ierrors.Errorf("missing type settings for %s; needed to check Must Occur rules", elemValue)
		}
		_, typePrefix, err := getTypeDenotationAndCode(elemTypeSettings.objectType)
		if err != nil {
			return ierrors.WithStack(err)
		}
		delete(mustOccurPrefixes, typePrefix)
	}

	if len(mustOccurPrefixes) != 0 {
		return ierrors.Wrapf(serializer.ErrArrayValidationTypesNotOccurred, "expected type prefixes that did not occur: %v", mustOccurPrefixes)
	}

	return nil
}

func (api *API) callSyntacticValidator(ctx context.Context, value reflect.Value, valueType reflect.Type) error {
	vldtrs, exists := api.validatorsRegistry.Get(valueType)

	// if the type doesn't exist in the registry, or the validator is not valid,
	// try to get the validator for the dereferenced pointer type
	if !exists || !vldtrs.syntacticValidator.IsValid() {
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			value = value.Elem()
			vldtrs, exists = api.validatorsRegistry.Get(valueType)
		}
	}

	if exists && vldtrs.syntacticValidator.IsValid() {
		if err, _ := vldtrs.syntacticValidator.Call(
			[]reflect.Value{reflect.ValueOf(ctx), value},
		)[0].Interface().(error); err != nil {
			return ierrors.Errorf("syntactic validator returned an error for type %s: %w", valueType, err)
		}
	}

	return nil
}

func (api *API) getStructFields(structType reflect.Type) ([]structField, error) {
	structFields, exists := api.structFieldsCache.Get(structType)
	if exists {
		return structFields, nil
	}

	structFields, err := parseStructFields(structType)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to parse struct type %s", structType)
	}
	api.structFieldsCache.Set(structType, structFields)

	return structFields, nil
}

// RegisterValidator registers a syntactic validator function that serix will call during the Encode and Decode processes.
//
// A syntactic validator validates the Go object and its data.
//
// For Encode, it is called for the original Go object before serix serializes the object into bytes.
// For Decode, it is called after serix builds the Go object from bytes.
//
// The validation is called for every registered type during the recursive traversal.
// It's an early stop process, if some validator returns an error serix stops the Encode/Decode and pops up the error.
//
// obj is an instance of the type you want to provide the validator for.
// Note that it's better to pass the obj as a value, not as a pointer
// because that way serix would be able to dereference pointers during Encode/Decode
// and detect the validators for both pointers and values
//
// syntacticValidatorFn is a function that accepts context.Context, and an object with the same type as obj.
//
// Example:
//
// syntacticValidator := func (ctx context.Context, t time.Time) error { ... }
//
// api.RegisterValidator(time.Time{}, syntacticValidator).
func (api *API) RegisterValidator(obj any, syntacticValidatorFn interface{}) error {
	return api.validatorsRegistry.RegisterValidator(obj, syntacticValidatorFn)
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
	return api.interfacesRegistry.RegisterInterfaceObjects(api.typeSettingsRegistry, iType, objs...)
}

func (api *API) ForEachRegisteredInterfaceObjects(consumer func(objType reflect.Type, interfaceObjects *InterfaceObjects) bool) {
	api.interfacesRegistry.ForEach(func(objType reflect.Type, interfaceObjects *InterfaceObjects) bool {
		return consumer(objType, interfaceObjects)
	})
}

// RegisterTypeSettings registers settings for a particular type obj.
// It's better to provide obj as a value, not a pointer,
// that way serix will be able to get the type settings for both values and pointers during Encode/Decode via de-referencing
// The settings provided via registration are considered global and default ones,
// they can be overridden by type settings parsed from struct tags
// or by type settings provided via option to the Encode/Decode methods.
// See TypeSettings for more detail.
func (api *API) RegisterTypeSettings(obj interface{}, ts TypeSettings) error {
	return api.typeSettingsRegistry.RegisterTypeSettings(obj, ts)
}

func (api *API) ForEachRegisteredTypeSetting(consumer func(objType reflect.Type, ts TypeSettings) bool) {
	api.typeSettingsRegistry.ForEach(func(objType reflect.Type, ts TypeSettings) bool {
		return consumer(objType, ts)
	})
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
