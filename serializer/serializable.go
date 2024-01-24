package serializer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
)

// Serializable is something which knows how to serialize/deserialize itself from/into bytes
// while also performing syntactical checks on the written/read data.
type Serializable interface {
	json.Marshaler
	json.Unmarshaler
	// Deserialize deserializes the given data (by copying) into the object and returns the amount of bytes consumed from data.
	// If the passed data is not big enough for deserialization, an error must be returned.
	// During deserialization additional validation may be performed if the given modes are set.
	Deserialize(data []byte, deSeriMode DeSerializationMode, deSeriCtx interface{}) (int, error)
	// Serialize returns a serialized byte representation.
	// During serialization additional validation may be performed if the given modes are set.
	Serialize(deSeriMode DeSerializationMode, deSeriCtx interface{}) ([]byte, error)
}

// SerializableWithSize implements Serializable interface and has the extra functionality of returning the size of the
// resulting serialized object (ideally without actually serializing it).
type SerializableWithSize interface {
	Serializable
	// Size returns the size of the serialized object
	Size() int
}

// Serializables is a slice of Serializable.
type Serializables []Serializable

// SerializableSlice is a slice of a type which can convert itself to Serializables.
type SerializableSlice interface {
	// ToSerializables returns the representation of the slice as a Serializables.
	ToSerializables() Serializables
	// FromSerializables updates the slice itself with the given Serializables.
	FromSerializables(seris Serializables)
}

// SerializableReadGuardFunc is a function that given a type prefix, returns an empty instance of the given underlying type.
// If the type doesn't resolve or is not supported in the deserialization context, an error is returned.
type SerializableReadGuardFunc func(ty uint32) (Serializable, error)

// SerializablePostReadGuardFunc is a function which inspects the read Serializable add runs additional validation against it.
type SerializablePostReadGuardFunc func(seri Serializable) error

// SerializableWriteGuardFunc is a function that given a Serializable, tells whether the given type is allowed to be serialized.
type SerializableWriteGuardFunc func(seri Serializable) error

// SerializableGuard defines the guards to de/serialize Serializable.
type SerializableGuard struct {
	// The read guard applied before reading an entire object.
	ReadGuard SerializableReadGuardFunc
	// The read guard applied after an object has been read.
	PostReadGuard SerializablePostReadGuardFunc
	// The write guard applied when writing objects.
	WriteGuard SerializableWriteGuardFunc
}

// DeSerializationMode defines the mode of de/serialization.
type DeSerializationMode byte

const (
	// DeSeriModeNoValidation instructs de/serialization to perform no validation.
	DeSeriModeNoValidation DeSerializationMode = 0
	// DeSeriModePerformValidation instructs de/serialization to perform validation.
	DeSeriModePerformValidation DeSerializationMode = 1 << 0
	// DeSeriModePerformLexicalOrdering instructs de/deserialization to automatically perform ordering of
	// certain arrays by their lexical serialized form.
	DeSeriModePerformLexicalOrdering DeSerializationMode = 1 << 1
)

// HasMode checks whether the de/serialization mode includes the given mode.
func (sm DeSerializationMode) HasMode(mode DeSerializationMode) bool {
	return sm&mode > 0
}

// ArrayValidationMode defines the mode of array validation.
type ArrayValidationMode byte

const (
	// ArrayValidationModeNone instructs the array validation to perform no validation.
	ArrayValidationModeNone ArrayValidationMode = 0
	// ArrayValidationModeNoDuplicates instructs the array validation to check for duplicates.
	ArrayValidationModeNoDuplicates ArrayValidationMode = 1 << 0
	// ArrayValidationModeLexicalOrdering instructs the array validation to check for lexical order.
	ArrayValidationModeLexicalOrdering ArrayValidationMode = 1 << 1
	// ArrayValidationModeAtMostOneOfEachTypeByte instructs the array validation to allow a given type prefix byte to occur only once in the array.
	ArrayValidationModeAtMostOneOfEachTypeByte ArrayValidationMode = 1 << 2
	// ArrayValidationModeAtMostOneOfEachTypeUint32 instructs the array validation to allow a given type prefix uint32 to occur only once in the array.
	ArrayValidationModeAtMostOneOfEachTypeUint32 ArrayValidationMode = 1 << 3
)

// HasMode checks whether the array element validation mode includes the given mode.
func (av ArrayValidationMode) HasMode(mode ArrayValidationMode) bool {
	return av&mode > 0
}

// TypePrefixes defines a set of type prefixes.
type TypePrefixes map[uint32]struct{}

// Subset checks whether every type prefix is a member of other.
func (typePrefixes TypePrefixes) Subset(other TypePrefixes) bool {
	for typePrefix := range typePrefixes {
		if _, has := other[typePrefix]; !has {
			return false
		}
	}

	return true
}

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules struct {
	// The min array bound.
	Min uint
	// The max array bound.
	Max uint
	// A map of object types which must occur within the array.
	// This is only checked on slices of types with an object type set.
	// In particular, this means this is not checked for byte slices.
	MustOccur TypePrefixes
	// The guards applied while de/serializing Serializables.
	Guards SerializableGuard
	// The mode of validation.
	ValidationMode ArrayValidationMode
}

// CheckBounds checks whether the given count violates the array bounds.
func (ar *ArrayRules) CheckBounds(count uint) error {
	if ar.Min != 0 && count < ar.Min {
		return ierrors.Wrapf(ErrArrayValidationMinElementsNotReached, "min is %d but count is %d", ar.Min, count)
	}
	if ar.Max != 0 && count > ar.Max {
		return ierrors.Wrapf(ErrArrayValidationMaxElementsExceeded, "max is %d but count is %d", ar.Max, count)
	}

	return nil
}

// ElementUniquenessSliceFunc is a function which takes a byte slice and reduces it to
// the part which is deemed relevant for uniqueness checks.
// If this function is used in conjunction with ArrayValidationModeLexicalOrdering, then the reduction
// must only reduce the slice from index 0 onwards, as otherwise lexical ordering on the set elements
// can not be enforced.
type ElementUniquenessSliceFunc func(next []byte) []byte

// ElementValidationFunc is a function which runs during array validation (e.g. lexical ordering).
type ElementValidationFunc func(index int, next []byte) error

// ElementUniqueValidator returns an ElementValidationFunc which returns an error if the given element is not unique.
func (ar *ArrayRules) ElementUniqueValidator() ElementValidationFunc {
	set := map[string]int{}

	return func(index int, next []byte) error {
		k := string(next)
		if j, has := set[k]; has {
			return ierrors.Wrapf(ErrArrayValidationViolatesUniqueness, "element %d and %d are duplicates", j, index)
		}
		set[k] = index

		return nil
	}
}

// LexicalOrderValidator returns an ElementValidationFunc which returns an error if the given byte slices
// are not ordered lexicographically.
func (ar *ArrayRules) LexicalOrderValidator() ElementValidationFunc {
	var prev []byte
	var prevIndex int

	return func(index int, next []byte) error {
		switch {
		case prev == nil:
			prev = next
			prevIndex = index
		case bytes.Compare(prev, next) > 0:
			return ierrors.Wrapf(ErrArrayValidationOrderViolatesLexicalOrder, "element %d should have been before element %d", index, prevIndex)
		default:
			prev = next
			prevIndex = index
		}

		return nil
	}
}

// LexicalOrderWithoutDupsValidator returns an ElementValidationFunc which returns an error if the given byte slices
// are not ordered lexicographically or any elements are duplicated.
func (ar *ArrayRules) LexicalOrderWithoutDupsValidator() ElementValidationFunc {
	var prev []byte
	var prevIndex int

	return func(index int, next []byte) error {
		if prev == nil {
			prevIndex = index
			prev = next

			return nil
		}
		switch bytes.Compare(prev, next) {
		case 1:
			return ierrors.Wrapf(ErrArrayValidationOrderViolatesLexicalOrder, "element %d should have been before element %d", index, prevIndex)
		case 0:
			// dup
			return ierrors.Wrapf(ErrArrayValidationViolatesUniqueness, "element %d and %d are duplicates", index, prevIndex)
		}
		prevIndex = index
		prev = next

		return nil
	}
}

// AtMostOneOfEachTypeValidator returns an ElementValidationFunc which returns an error if a given type occurs multiple
// times within the array.
func (ar *ArrayRules) AtMostOneOfEachTypeValidator(typeDenotation TypeDenotationType) ElementValidationFunc {
	seen := map[uint32]int{}

	return func(index int, next []byte) error {
		var key uint32
		switch typeDenotation {
		case TypeDenotationUint32:
			if len(next) < UInt32ByteSize {
				return ierrors.Wrap(ErrInvalidBytes, "not enough bytes to check type uniquness in array")
			}
			key = binary.LittleEndian.Uint32(next)
		case TypeDenotationByte:
			if len(next) < OneByte {
				return ierrors.Wrap(ErrInvalidBytes, "not enough bytes to check type uniquness in array")
			}
			key = uint32(next[0])
		default:
			panic(fmt.Sprintf("unknown type denotation in AtMostOneOfEachTypeValidator passed: %d", typeDenotation))
		}
		prevIndex, has := seen[key]
		if has {
			return ierrors.Wrapf(ErrArrayValidationViolatesTypeUniqueness, "element %d and %d have the same type", index, prevIndex)
		}
		seen[key] = index

		return nil
	}
}

// ElementValidationFunc returns a new ElementValidationFunc according to the given mode.
func (ar *ArrayRules) ElementValidationFunc() ElementValidationFunc {
	var arrayElementValidator ElementValidationFunc

	wrap := func(f ElementValidationFunc, f2 ElementValidationFunc) ElementValidationFunc {
		return func(index int, next []byte) error {
			if f != nil {
				if err := f(index, next); err != nil {
					return err
				}
			}

			return f2(index, next)
		}
	}

	for i := byte(1); i != 0; i <<= 1 {
		switch ArrayValidationMode(byte(ar.ValidationMode) & i) {
		case ArrayValidationModeNone:
		case ArrayValidationModeNoDuplicates:
			if ar.ValidationMode.HasMode(ArrayValidationModeLexicalOrdering) {
				continue
			}
			arrayElementValidator = wrap(arrayElementValidator, ar.ElementUniqueValidator())
		case ArrayValidationModeLexicalOrdering:
			// optimization: if lexical order and no dups are enforced, then byte comparison
			// to the previous element can be done instead of using a map
			if ar.ValidationMode.HasMode(ArrayValidationModeNoDuplicates) {
				arrayElementValidator = wrap(arrayElementValidator, ar.LexicalOrderWithoutDupsValidator())

				continue
			}
			arrayElementValidator = wrap(arrayElementValidator, ar.LexicalOrderValidator())
		case ArrayValidationModeAtMostOneOfEachTypeByte:
			arrayElementValidator = wrap(arrayElementValidator, ar.AtMostOneOfEachTypeValidator(TypeDenotationByte))
		case ArrayValidationModeAtMostOneOfEachTypeUint32:
			arrayElementValidator = wrap(arrayElementValidator, ar.AtMostOneOfEachTypeValidator(TypeDenotationUint32))
		}
	}

	return arrayElementValidator
}

// LexicalOrderedByteSlices are byte slices ordered in lexical order.
type LexicalOrderedByteSlices [][]byte

func (l LexicalOrderedByteSlices) Len() int {
	return len(l)
}

func (l LexicalOrderedByteSlices) Less(i, j int) bool {
	return bytes.Compare(l[i], l[j]) < 0
}

func (l LexicalOrderedByteSlices) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LexicalOrdered32ByteArrays are 32 byte arrays ordered in lexical order.
type LexicalOrdered32ByteArrays [][32]byte

func (l LexicalOrdered32ByteArrays) Len() int {
	return len(l)
}

func (l LexicalOrdered32ByteArrays) Less(i, j int) bool {
	return bytes.Compare(l[i][:], l[j][:]) < 0
}

func (l LexicalOrdered32ByteArrays) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LexicalOrdered36ByteArrays are 36 byte arrays ordered in lexical order.
type LexicalOrdered36ByteArrays [][36]byte

func (l LexicalOrdered36ByteArrays) Len() int {
	return len(l)
}

func (l LexicalOrdered36ByteArrays) Less(i, j int) bool {
	return bytes.Compare(l[i][:], l[j][:]) < 0
}

func (l LexicalOrdered36ByteArrays) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LexicalOrdered40ByteArrays are 40 byte arrays ordered in lexical order.
type LexicalOrdered40ByteArrays [][40]byte

func (l LexicalOrdered40ByteArrays) Len() int {
	return len(l)
}

func (l LexicalOrdered40ByteArrays) Less(i, j int) bool {
	return bytes.Compare(l[i][:], l[j][:]) < 0
}

func (l LexicalOrdered40ByteArrays) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// SortedSerializables are Serializables sorted by their serialized form.
type SortedSerializables Serializables

func (ss SortedSerializables) Len() int {
	return len(ss)
}

func (ss SortedSerializables) Less(i, j int) bool {
	iData, _ := ss[i].Serialize(DeSeriModeNoValidation, nil)
	jData, _ := ss[j].Serialize(DeSeriModeNoValidation, nil)

	return bytes.Compare(iData, jData) < 0
}

func (ss SortedSerializables) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
