package ds

import (
	"github.com/iotaledger/hive.go/ds/walker"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// region Set //////////////////////////////////////////////////////////////////////////////////////////////////////////

// Set is a generic thread-safe collection of unique elements.
type Set[ElementType comparable] interface {
	// WriteableSet imports the write methods of the Set interface.
	WriteableSet[ElementType]

	// ReadableSet imports the read methods of the Set interface.
	ReadableSet[ElementType]
}

// NewSet creates a new Set with the given elements.
func NewSet[T comparable](elements ...T) Set[T] {
	return newSet(elements...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ReadableSet //////////////////////////////////////////////////////////////////////////////////////////////////

// ReadableSet bundles the read methods of the Set interface.
type ReadableSet[ElementType comparable] interface {
	// Has returns true if the set contains the given element.
	Has(key ElementType) (has bool)

	// HasAll returns true if the set contains all elements of the given set.
	HasAll(other ReadableSet[ElementType]) bool

	// ForEach iterates through all elements of the set (returning an error will stop the iteration).
	ForEach(callback func(element ElementType) error) error

	// Range iterates through all elements of the set.
	Range(callback func(element ElementType))

	// Intersect returns the intersection of the set and the given set.
	Intersect(other ReadableSet[ElementType]) Set[ElementType]

	// Filter returns a new set with all elements that satisfy the given predicate.
	Filter(predicate func(element ElementType) bool) Set[ElementType]

	// Equals returns true if the set contains the same elements as the given set.
	Equals(other ReadableSet[ElementType]) bool

	// Any returns a random element from the set (and if one exists).
	Any() (element ElementType, exists bool)

	// Is returns true if the given element is the only element in the set.
	Is(element ElementType) bool

	// Iterator returns an iterator for the set.
	Iterator() *walker.Walker[ElementType]

	// Clone returns a shallow copy of the set.
	Clone() Set[ElementType]

	// Size returns the number of elements in the set.
	Size() int

	// IsEmpty returns true if the set is empty.
	IsEmpty() bool

	// Clear removes all elements from the set.
	Clear()

	// ToSlice returns a slice representation of the set.
	ToSlice() []ElementType

	// Encode encodes the set into a byte slice.
	Encode(serixAPI *serix.API) ([]byte, error)

	// String returns a string representation of the set.
	String() string
}

// NewReadableSet creates a new readable Set with the given elements.
func NewReadableSet[T comparable](elements ...T) ReadableSet[T] {
	return newReadableSet[T](elements...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region WriteableSet /////////////////////////////////////////////////////////////////////////////////////////////////

// WriteableSet bundles all write methods of the Set interface.
type WriteableSet[ElementType comparable] interface {
	// Add adds a new element to the set and returns true if the element was not present in the set before.
	Add(element ElementType) bool

	// AddAll adds all elements to the set and returns true if any element has been added.
	AddAll(elements ReadableSet[ElementType]) (addedElements Set[ElementType])

	// Delete deletes the given element from the set.
	Delete(element ElementType) bool

	// DeleteAll deletes the given elements from the set.
	DeleteAll(other ReadableSet[ElementType]) (removedElements Set[ElementType])

	// Apply tries to apply the given mutations to the set atomically and returns the applied mutations.
	Apply(mutations SetMutations[ElementType]) (appliedMutations SetMutations[ElementType])

	// Compute tries to compute the mutations for the set atomically and returns the applied mutations.
	Compute(mutationFactory func(set ReadableSet[ElementType]) SetMutations[ElementType]) (appliedMutations SetMutations[ElementType])

	// Replace replaces the elements of the set with the given elements and returns the removed elements.
	Replace(elements ReadableSet[ElementType]) (removedElements Set[ElementType])

	// Decode decodes the set from a byte slice.
	Decode(serixAPI *serix.API, data []byte) (bytesRead int, err error)

	// ReadOnly returns a read-only version of the set.
	ReadOnly() ReadableSet[ElementType]
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region SetMutations /////////////////////////////////////////////////////////////////////////////////////////////////

// SetMutations represents a set of mutations that can be applied to a Set atomically.
type SetMutations[ElementType comparable] interface {
	// WithAddedElements is a setter for the added elements of the setMutations.
	WithAddedElements(elements Set[ElementType]) SetMutations[ElementType]

	// WithDeletedElements is a setter for the deleted elements of the setMutations.
	WithDeletedElements(elements Set[ElementType]) SetMutations[ElementType]

	// AddedElements returns the elements that are supposed to be added.
	AddedElements() Set[ElementType]

	// DeletedElements returns the elements that are supposed to be removed.
	DeletedElements() Set[ElementType]

	// IsEmpty returns true if the SetMutations instance is empty.
	IsEmpty() bool
}

// NewSetMutations creates a new SetMutations instance.
func NewSetMutations[ElementType comparable](elements ...ElementType) SetMutations[ElementType] {
	return newSetMutations(elements...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region SetArithmetic ////////////////////////////////////////////////////////////////////////////////////////////////

// SetArithmetic is an interface that allows to perform arithmetic operations on a set of elements to return the
// resulting mutations of the operation.
type SetArithmetic[ElementType comparable] interface {
	// Add adds the given mutations to the elements and returns the resulting net mutations for the set that are formed
	// by tracking the elements that rise above the given threshold (defaults to 1).
	Add(mutations SetMutations[ElementType], threshold ...int) SetMutations[ElementType]

	// AddedElementsCollector returns a function that adds an element to the given mutations if its occurrence count
	// reaches the given threshold (defaults to 1) after the addition.
	AddedElementsCollector(mutations SetMutations[ElementType], threshold ...int) func(addedElement ElementType)

	// Subtract subtracts the given mutations from the elements and returns the resulting net mutations for the set that
	// are formed by tracking the elements that fall below the given threshold (defaults to 1).
	Subtract(mutations SetMutations[ElementType], threshold ...int) SetMutations[ElementType]

	// SubtractedElementsCollector returns a function that deletes an element from the given mutations if its occurrence
	// count falls below the given threshold (defaults to 1) after the subtraction.
	SubtractedElementsCollector(mutations SetMutations[ElementType], threshold ...int) func(ElementType)
}

// NewSetArithmetic creates a new SetArithmetic instance.
func NewSetArithmetic[ElementType comparable]() SetArithmetic[ElementType] {
	return newSetArithmetic[ElementType]()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
