package set

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ds/walker"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// region Set

// Set is a generic thread-safe collection of unique elements.
type Set[ElementType comparable] interface {
	// Add adds a new element to the set and returns true if the element was not present in the set before.
	Add(element ElementType) bool

	// AddAll adds all elements to the set and returns true if any element has been added.
	AddAll(elements ReadOnly[ElementType]) (addedElements Set[ElementType])

	// Delete deletes the given element from the set.
	Delete(element ElementType) bool

	// DeleteAll deletes the given elements from the set.
	DeleteAll(other ReadOnly[ElementType]) (removedElements Set[ElementType])

	// Decode decodes the set from a byte slice.
	Decode(b []byte) (bytesRead int, err error)

	// ToReadOnly returns a read-only version of the set.
	ToReadOnly() ReadOnly[ElementType]

	// ReadOnly imports the read methods from ReadOnly.
	ReadOnly[ElementType]
}

// New creates a new Set with the given elements.
func New[T comparable](elements ...T) Set[T] {
	s := &set[T]{&readOnly[T]{
		OrderedMap: orderedmap.New[T, types.Empty](),
	}}

	for _, element := range elements {
		s.OrderedMap.Set(element, types.Void)
	}

	return s
}

// set implements the Set interface.
type set[ElementType comparable] struct {
	*readOnly[ElementType]
}

// Add adds a new element to the set and returns true if the element was not present in the set before.
func (s *set[ElementType]) Add(element ElementType) bool {
	return !lo.Return2(s.Set(element, types.Void))
}

// AddAll adds all elements to the set and returns true if any element has been added.
func (s *set[ElementType]) AddAll(elements ReadOnly[ElementType]) (addedElements Set[ElementType]) {
	addedElements = New[ElementType]()
	_ = elements.ForEach(func(element ElementType) (err error) {
		if !lo.Return2(s.Set(element, types.Void)) {
			addedElements.Add(element)
		}

		return nil
	})

	return addedElements
}

// Delete deletes the given element from the set.
func (s *set[ElementType]) Delete(element ElementType) bool {
	return s.OrderedMap.Delete(element)
}

// DeleteAll deletes the given elements from the set.
func (s *set[ElementType]) DeleteAll(other ReadOnly[ElementType]) (removedElements Set[ElementType]) {
	removedElements = New[ElementType]()
	_ = other.ForEach(func(element ElementType) (err error) {
		if s.Delete(element) {
			removedElements.Add(element)
		}

		return nil
	})

	return removedElements
}

// ToReadOnly returns a read-only version of the set.
func (s *set[ElementType]) ToReadOnly() ReadOnly[ElementType] {
	return s.readOnly
}

// code contract (make sure the type implements all required methods).
var _ Set[int] = new(set[int])

// endregion

// region ReadOnly

// ReadOnly is a generic thread-safe collection of unique elements that can only be read.
type ReadOnly[ElementType comparable] interface {
	// Has returns true if the set contains the given element.
	Has(key ElementType) (has bool)

	// HasAll returns true if the set contains all elements of the given set.
	HasAll(other ReadOnly[ElementType]) bool

	// ForEach iterates through all elements of the set (returning an error will stop the iteration).
	ForEach(callback func(element ElementType) error) error

	// Range iterates through all elements of the set.
	Range(callback func(element ElementType))

	// Intersect returns the intersection of the set and the given set.
	Intersect(other ReadOnly[ElementType]) Set[ElementType]

	// Filter returns a new set with all elements that satisfy the given predicate.
	Filter(predicate func(element ElementType) bool) Set[ElementType]

	// Equals returns true if the set contains the same elements as the given set.
	Equals(other ReadOnly[ElementType]) bool

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
	Encode() ([]byte, error)

	// String returns a string representation of the set.
	String() string
}

// readOnly implements the ReadOnly interface.
type readOnly[T comparable] struct {
	*orderedmap.OrderedMap[T, types.Empty] `serix:"0"`
}

// HasAll returns true if the set contains all elements of the given set.
func (r *readOnly[T]) HasAll(other ReadOnly[T]) bool {
	if r == nil {
		return false
	}

	return other.ForEach(func(element T) error {
		if !r.Has(element) {
			return ierrors.New("element not found")
		}

		return nil
	}) == nil
}

// ForEach iterates through all elements of the set (returning an error will stop the iteration).
func (r *readOnly[T]) ForEach(callback func(element T) error) (err error) {
	if r == nil {
		return nil
	}

	r.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
		if err = callback(element); err != nil {
			return false
		}

		return true
	})

	return err
}

// Range iterates through all elements of the set.
func (r *readOnly[T]) Range(callback func(element T)) {
	if r != nil {
		r.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
			callback(element)

			return true
		})
	}
}

// Intersect returns the intersection of the set and the given set.
func (r *readOnly[T]) Intersect(other ReadOnly[T]) (intersection Set[T]) {
	return r.Filter(other.Has)
}

// Filter returns a new set with all elements that satisfy the given predicate.
func (r *readOnly[T]) Filter(predicate func(element T) bool) (filtered Set[T]) {
	filtered = New[T]()
	_ = r.ForEach(func(element T) (err error) {
		if predicate(element) {
			filtered.Add(element)
		}

		return nil
	})

	return filtered
}

// Equals returns true if the set contains the same elements as the given set.
func (r *readOnly[T]) Equals(other ReadOnly[T]) (equal bool) {
	return r == other || (r != nil && other != nil && r.Size() == other.Size() && r.HasAll(other))
}

// Is returns true if the given element is the only element in the set.
func (r *readOnly[T]) Is(element T) bool {
	return r.Size() == 1 && r.Has(element)
}

// Iterator returns an iterator for the set.
func (r *readOnly[T]) Iterator() *walker.Walker[T] {
	return walker.New[T](false).PushAll(r.ToSlice()...)
}

// Clone returns a shallow copy of the set.
func (r *readOnly[T]) Clone() (cloned Set[T]) {
	return New[T]().AddAll(r)
}

// ToSlice returns a slice representation of the set.
func (r *readOnly[T]) ToSlice() (slice []T) {
	slice = make([]T, 0)

	if r != nil {
		_ = r.ForEach(func(element T) error {
			slice = append(slice, element)

			return nil
		})
	}

	return slice
}

// String returns a string representation of the set.
func (r *readOnly[T]) String() (humanReadable string) {
	var elementType T
	elementTypeName := reflect.TypeOf(elementType).Name()

	elementStrings := make([]string, 0)
	_ = r.ForEach(func(element T) (err error) {
		elementStrings = append(elementStrings, strings.TrimRight(strings.ReplaceAll(fmt.Sprintf("%+v", element), elementTypeName+"(", ""), ")"))

		return nil
	})

	return fmt.Sprintf("%ss(%s)", elementTypeName, strings.Join(elementStrings, ", "))
}

// code contract (make sure the type implements all required methods).
var _ ReadOnly[int] = new(readOnly[int])

// endregion
