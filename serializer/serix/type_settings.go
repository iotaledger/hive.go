package serix

import (
	"github.com/iotaledger/hive.go/serializer/v2"
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
)

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules serializer.ArrayRules

// MapElementRules defines rules around to be deserialized map elements (key or value).
// MinLength and MaxLength at 0 define an unbounded map element.
type MapElementRules struct {
	LengthPrefixType *LengthPrefixType
	MinLength        uint
	MaxLength        uint
}

func (m *MapElementRules) ToTypeSettings() TypeSettings {
	return TypeSettings{
		lengthPrefixType: m.LengthPrefixType,
		arrayRules: &ArrayRules{
			Min: m.MinLength,
			Max: m.MaxLength,
		},
	}
}

// MapRules defines rules around a to be deserialized map.
type MapRules struct {
	// MinEntries defines the min entries for the map.
	MinEntries uint
	// MaxEntries defines the max entries for the map. 0 means unbounded.
	MaxEntries uint
	// MaxByteSize defines the max serialized byte size for the map. 0 means unbounded.
	MaxByteSize uint

	// KeyRules define the rules applied to the keys of the map.
	KeyRules *MapElementRules
	// ValueRules define the rules applied to the values of the map.
	ValueRules *MapElementRules
}

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
	lengthPrefixType *LengthPrefixType
	objectType       interface{}
	lexicalOrdering  *bool
	mapKey           *string
	arrayRules       *ArrayRules
	mapRules         *MapRules
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

// WithMapRules specifies the map rules.
func (ts TypeSettings) WithMapRules(rules *MapRules) TypeSettings {
	ts.mapRules = rules

	return ts
}

// MapRules returns the map rules.
func (ts TypeSettings) MapRules() *MapRules {
	return ts.mapRules
}

// WithMapMinEntries specifies the min entries for the map.
func (ts TypeSettings) WithMapMinEntries(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	ts.mapRules.MinEntries = l

	return ts
}

// WithMapMaxEntries specifies the max entries for the map.
func (ts TypeSettings) WithMapMaxEntries(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	ts.mapRules.MaxEntries = l

	return ts
}

// WithMapMaxByteSize specifies max serialized byte size for the map. 0 means unbounded.
func (ts TypeSettings) WithMapMaxByteSize(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	ts.mapRules.MaxByteSize = l

	return ts
}

// WithMapKeyLengthPrefixType specifies MapKeyLengthPrefixType.
func (ts TypeSettings) WithMapKeyLengthPrefixType(lpt LengthPrefixType) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.KeyRules == nil {
		ts.mapRules.KeyRules = new(MapElementRules)
	}
	ts.mapRules.KeyRules.LengthPrefixType = &lpt

	return ts
}

// WithMapKeyMinLen specifies the min length for the object in the map key.
func (ts TypeSettings) WithMapKeyMinLen(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.KeyRules == nil {
		ts.mapRules.KeyRules = new(MapElementRules)
	}
	ts.mapRules.KeyRules.MinLength = l

	return ts
}

// WithMapKeyMaxLen specifies the max length for the object in the map key.
func (ts TypeSettings) WithMapKeyMaxLen(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.KeyRules == nil {
		ts.mapRules.KeyRules = new(MapElementRules)
	}
	ts.mapRules.KeyRules.MaxLength = l

	return ts
}

// MapValueLengthPrefixType specifies MapValueLengthPrefixType.
func (ts TypeSettings) WithMapValueLengthPrefixType(lpt LengthPrefixType) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.ValueRules == nil {
		ts.mapRules.ValueRules = new(MapElementRules)
	}
	ts.mapRules.ValueRules.LengthPrefixType = &lpt

	return ts
}

// WithMapValueMinLen specifies the min length for the object in the map value.
func (ts TypeSettings) WithMapValueMinLen(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.ValueRules == nil {
		ts.mapRules.ValueRules = new(MapElementRules)
	}
	ts.mapRules.ValueRules.MinLength = l

	return ts
}

// WithMapValueMaxLen specifies the max length for the object in the map value.
func (ts TypeSettings) WithMapValueMaxLen(l uint) TypeSettings {
	if ts.mapRules == nil {
		ts.mapRules = new(MapRules)
	}
	if ts.mapRules.ValueRules == nil {
		ts.mapRules.ValueRules = new(MapElementRules)
	}
	ts.mapRules.ValueRules.MaxLength = l

	return ts
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
	if ts.mapRules == nil {
		ts.mapRules = other.mapRules
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
