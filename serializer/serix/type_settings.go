package serix

import (
	"reflect"
	"sync"

	hiveorderedmap "github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrUnknownLengthPrefixType gets returned when an unknown LengthPrefixType is used.
	ErrUnknownLengthPrefixType = ierrors.New("unknown length prefix type")
)

// LengthPrefixType defines the type of the value denoting the length of a collection.
type LengthPrefixType serializer.SeriLengthPrefixType

const (
	// LengthPrefixTypeAsByte defines a collection length to be denoted by a byte.
	LengthPrefixTypeAsByte = LengthPrefixType(serializer.SeriLengthPrefixTypeAsByte)
	// LengthPrefixTypeAsUint16 defines a collection length to be denoted by a uint16.
	LengthPrefixTypeAsUint16 = LengthPrefixType(serializer.SeriLengthPrefixTypeAsUint16)
	// LengthPrefixTypeAsUint32 defines a collection length to be denoted by a uint32.
	LengthPrefixTypeAsUint32 = LengthPrefixType(serializer.SeriLengthPrefixTypeAsUint32)
	// LengthPrefixTypeAsUint64 defines a collection length to be denoted by a uint64.
	LengthPrefixTypeAsUint64 = LengthPrefixType(serializer.SeriLengthPrefixTypeAsUint64)
)

func LengthPrefixTypeSize(t LengthPrefixType) (int, error) {
	switch t {
	case LengthPrefixTypeAsByte:
		return 1, nil
	case LengthPrefixTypeAsUint16:
		return 2, nil
	case LengthPrefixTypeAsUint32:
		return 4, nil
	case LengthPrefixTypeAsUint64:
		return 8, nil
	default:
		return 0, ErrUnknownLengthPrefixType
	}
}

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules serializer.ArrayRules

// TypeSettings holds various settings for a particular type.
// Those settings determine how the object should be serialized/deserialized.
// There are three ways to provide TypeSettings
// 1. Via global registry: API.RegisterTypeSettings().
// 2. Parse from struct tags.
// 3. Pass as an option to API.Encode/API.Decode methods.
// The type settings provided via struct tags or an option override the type settings from the registry.
// So the precedence is the following 1<2<3.
// See API.RegisterTypeSettings() and WithTypeSettings() for more detail.
type TypeSettings struct {
	// fieldKey defines the key for the field used in json serialization.
	fieldKey *string
	// description defines the description of the object.
	description string
	// objectType defines the object type. It can be either uint8 or uint32 number.
	objectType interface{}
	// lengthPrefixType defines the type of the value denoting the length of a collection.
	lengthPrefixType *LengthPrefixType
	// lexicalOrdering defines whether the collection must be lexically ordered during serialization.
	lexicalOrdering *bool
	// arrayRules defines rules around a to be deserialized array.
	arrayRules *ArrayRules
}

func NewTypeSettings() TypeSettings {
	return TypeSettings{}
}

// WithFieldKey specifies the key for the field.
func (ts TypeSettings) WithFieldKey(fieldKey string) TypeSettings {
	ts.fieldKey = &fieldKey

	return ts
}

// FieldKey returns the key for the field.
func (ts TypeSettings) FieldKey() (string, bool) {
	if ts.fieldKey == nil {
		return "", false
	}

	return *ts.fieldKey, true
}

// MustFieldKey must return a key for the field.
func (ts TypeSettings) MustFieldKey() string {
	if ts.fieldKey == nil {
		panic("no field key set")
	}

	return *ts.fieldKey
}

// WithDescription specifies the description of the object.
func (ts TypeSettings) WithDescription(description string) TypeSettings {
	ts.description = description

	return ts
}

// Description returns the description of the object.
func (ts TypeSettings) Description() string {
	return ts.description
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
	if ts.fieldKey == nil {
		ts.fieldKey = other.fieldKey
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

// checkMinMaxBoundsLength checks whether the given length is within its defined bounds.
func (ts TypeSettings) checkMinMaxBoundsLength(length int) error {
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
func (ts TypeSettings) checkMinMaxBounds(v reflect.Value) error {
	if has := hasLength(v); !has {
		return nil
	}

	if err := ts.checkMinMaxBoundsLength(v.Len()); err != nil {
		return ierrors.Wrapf(err, "can't serialize '%s' type", v.Kind())
	}

	return nil
}

type TypeSettingsRegistry struct {
	// the registered type settings for the known objects
	registryMutex sync.RWMutex
	registry      *hiveorderedmap.OrderedMap[reflect.Type, TypeSettings]
}

func NewTypeSettingsRegistry() *TypeSettingsRegistry {
	return &TypeSettingsRegistry{
		registry: hiveorderedmap.New[reflect.Type, TypeSettings](),
	}
}

func (r *TypeSettingsRegistry) Has(objType reflect.Type) bool {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	return r.registry.Has(objType)
}

func (r *TypeSettingsRegistry) GetByType(objType reflect.Type) (TypeSettings, bool) {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	ts, ok := r.registry.Get(objType)
	if ok {
		return ts, true
	}

	// if there is no type settings for the given type, and the type is a pointer,
	// try to get the type settings for the pointer type.
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		ts, ok = r.registry.Get(objType)

		return ts, ok
	}

	return TypeSettings{}, false
}

func (r *TypeSettingsRegistry) GetByValue(objValue reflect.Value, optTS ...TypeSettings) TypeSettings {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	for {
		if ts, ok := r.registry.Get(objValue.Type()); ok {
			if len(optTS) > 0 {
				return optTS[0].merge(ts)
			}

			return ts
		}

		// resolve indirections
		switch objValue.Kind() {
		case reflect.Ptr, reflect.Interface:
			objValue = objValue.Elem()

		default:
			if len(optTS) > 0 {
				return optTS[0]
			}

			return TypeSettings{}
		}
	}
}

func (r *TypeSettingsRegistry) ForEach(consumer func(objType reflect.Type, ts TypeSettings) bool) {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	r.registry.ForEach(func(objType reflect.Type, ts TypeSettings) bool {
		return consumer(objType, ts)
	})
}

// RegisterTypeSettings registers settings for a particular type obj.
func (r *TypeSettingsRegistry) RegisterTypeSettings(obj interface{}, ts TypeSettings) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return ierrors.New("'obj' is a nil interface, it needs to be a valid type")
	}

	r.registryMutex.Lock()
	defer r.registryMutex.Unlock()

	if r.registry.Has(objType) {
		return ierrors.Errorf("type settings for object %s are already registered", objType)
	}

	r.registry.Set(objType, ts)

	return nil
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

func (r *TypeSettingsRegistry) getTypeDenotationAndObjectCode(objType reflect.Type) (serializer.TypeDenotationType, uint32, error) {
	ts, exists := r.GetByType(objType)
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

type ObjectMetadata struct {
	Type           reflect.Type
	TypeDenotation serializer.TypeDenotationType
	Code           uint32
}

func (r *TypeSettingsRegistry) GetObjectMetadata(obj any) (*ObjectMetadata, error) {
	objType := reflect.TypeOf(obj)

	objTypeDenotation, objCode, err := r.getTypeDenotationAndObjectCode(objType)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to get type denotation for object %T", obj)
	}

	return &ObjectMetadata{
		Type:           objType,
		TypeDenotation: objTypeDenotation,
		Code:           objCode,
	}, nil
}
