package ds

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/ds/serializableorderedmap"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ds/walker"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// region set //////////////////////////////////////////////////////////////////////////////////////////////////////////

// set is the standard implementation of the Set interface.
type set[ElementType comparable] struct {
	// readable embeds the readable part of the set.
	*readableSet[ElementType]

	// applyMutex is used to make calls to Apply atomic.
	applyMutex sync.RWMutex
}

// newSet creates a new set with the given elements.
func newSet[ElementType comparable](elements ...ElementType) *set[ElementType] {
	return &set[ElementType]{
		readableSet: newReadableSet(elements...),
	}
}

// Add adds a new element to the set and returns true if the element was not present in the set before.
func (s *set[ElementType]) Add(element ElementType) bool {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	return !lo.Return2(s.Set(element, types.Void))
}

// AddAll tries to add all elements to the set and returns the elements that have been added.
func (s *set[ElementType]) AddAll(elements ReadableSet[ElementType]) (addedElements Set[ElementType]) {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	addedElements = NewSet[ElementType]()
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
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	return s.OrderedMap.Delete(element)
}

// DeleteAll deletes the given elements from the set.
func (s *set[ElementType]) DeleteAll(other ReadableSet[ElementType]) (removedElements Set[ElementType]) {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	removedElements = NewSet[ElementType]()
	_ = other.ForEach(func(element ElementType) (err error) {
		if s.Delete(element) {
			removedElements.Add(element)
		}

		return nil
	})

	return removedElements
}

// Apply tries to apply the given mutations to the set atomically and returns the mutations that have been applied.
func (s *set[ElementType]) Apply(mutations SetMutations[ElementType]) (appliedMutations SetMutations[ElementType]) {
	s.applyMutex.Lock()
	defer s.applyMutex.Unlock()

	return s.apply(mutations)
}

// Compute tries to compute the mutations for the set atomically and returns the applied mutations.
func (s *set[ElementType]) Compute(mutationFactory func(set ReadableSet[ElementType]) SetMutations[ElementType]) (appliedMutations SetMutations[ElementType]) {
	s.applyMutex.Lock()
	defer s.applyMutex.Unlock()

	return s.apply(mutationFactory(s.readableSet))
}

// Replace replaces the elements of the set with the given elements and returns the previous elements of the set.
func (s *set[ElementType]) Replace(elements ReadableSet[ElementType]) (previousElements Set[ElementType]) {
	s.applyMutex.Lock()
	defer s.applyMutex.Unlock()

	previousElements = NewSet(s.ToSlice()...)
	s.Clear()

	elements.Range(func(element ElementType) {
		s.Set(element, types.Void)
	})

	return previousElements
}

// ReadOnly returns a read-only version of the set.
func (s *set[ElementType]) ReadOnly() ReadableSet[ElementType] {
	return s.readableSet
}

// apply tries to apply the given mutations to the set atomically and returns the mutations that have been applied.
func (s *set[ElementType]) apply(mutations SetMutations[ElementType]) (appliedMutations SetMutations[ElementType]) {
	addedElements := NewSet[ElementType]()
	mutations.AddedElements().Range(func(element ElementType) {
		if !lo.Return2(s.Set(element, types.Void)) {
			addedElements.Add(element)
		}
	})

	removedElements := NewSet[ElementType]()
	mutations.DeletedElements().Range(func(element ElementType) {
		if s.OrderedMap.Delete(element) {
			removedElements.Add(element)
		}
	})

	return NewSetMutations[ElementType]().WithAddedElements(addedElements).WithDeletedElements(removedElements)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region readableSet //////////////////////////////////////////////////////////////////////////////////////////////////

// readableSet is the standard implementation of the ReadableSet interface.
//
//nolint:tagliatelle // heck knows why this linter fails here
type readableSet[T comparable] struct {
	*serializableorderedmap.SerializableOrderedMap[T, types.Empty] `serix:""`
}

// newReadableSet creates a new readable set with the given elements.
func newReadableSet[T comparable](elements ...T) *readableSet[T] {
	r := &readableSet[T]{
		SerializableOrderedMap: serializableorderedmap.New[T, types.Empty](),
	}

	for _, element := range elements {
		r.OrderedMap.Set(element, types.Void)
	}

	return r
}

// HasAll returns true if the set contains all elements of the given set.
func (r *readableSet[T]) HasAll(other ReadableSet[T]) bool {
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
func (r *readableSet[T]) ForEach(callback func(element T) error) (err error) {
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
func (r *readableSet[T]) Range(callback func(element T)) {
	if r != nil {
		r.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
			callback(element)

			return true
		})
	}
}

// Intersect returns the intersection of the set and the given set.
func (r *readableSet[T]) Intersect(other ReadableSet[T]) (intersection Set[T]) {
	return r.Filter(other.Has)
}

// Filter returns a new set with all elements that satisfy the given predicate.
func (r *readableSet[T]) Filter(predicate func(element T) bool) (filtered Set[T]) {
	filtered = NewSet[T]()
	_ = r.ForEach(func(element T) (err error) {
		if predicate(element) {
			filtered.Add(element)
		}

		return nil
	})

	return filtered
}

// Equals returns true if the set contains the same elements as the given set.
func (r *readableSet[T]) Equals(other ReadableSet[T]) (equal bool) {
	return r == other || (r != nil && other != nil && r.Size() == other.Size() && r.HasAll(other))
}

// Any returns a random element from the set (and if one exists).
func (r *readableSet[T]) Any() (element T, exists bool) {
	if r != nil {
		r.OrderedMap.ForEach(func(firstElement T, _ types.Empty) bool {
			element = firstElement
			exists = true

			return false
		})
	}

	return element, exists
}

// Is returns true if the given element is the only element in the set.
func (r *readableSet[T]) Is(element T) bool {
	return r.Size() == 1 && r.Has(element)
}

// Iterator returns an iterator for the set.
func (r *readableSet[T]) Iterator() *walker.Walker[T] {
	return walker.New[T](false).PushAll(r.ToSlice()...)
}

// Clone returns a shallow copy of the set.
func (r *readableSet[T]) Clone() (cloned Set[T]) {
	return NewSet[T]().AddAll(r)
}

// ToSlice returns a slice representation of the set.
func (r *readableSet[T]) ToSlice() (slice []T) {
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
func (r *readableSet[T]) String() (humanReadable string) {
	var elementType T
	elementTypeName := reflect.TypeOf(elementType).Name()

	elementStrings := make([]string, 0)
	_ = r.ForEach(func(element T) (err error) {
		elementStrings = append(elementStrings, strings.TrimRight(strings.ReplaceAll(fmt.Sprintf("%+v", element), elementTypeName+"(", ""), ")"))

		return nil
	})

	return fmt.Sprintf("%ss(%s)", elementTypeName, strings.Join(elementStrings, ", "))
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region setMutations /////////////////////////////////////////////////////////////////////////////////////////////////

// setMutations is the default implementation of the SetMutations interface.
type setMutations[ElementType comparable] struct {
	// AddedElements are the elements that are supposed to be added.
	addedElements Set[ElementType]

	// deletedElements are the elements that are supposed to be removed.
	deletedElements Set[ElementType]
}

// newSetMutations creates a new setMutations instance.
func newSetMutations[ElementType comparable](elements ...ElementType) *setMutations[ElementType] {
	return &setMutations[ElementType]{
		addedElements:   NewSet[ElementType](elements...),
		deletedElements: NewSet[ElementType](),
	}
}

// WithAddedElements is a setter for the added elements of the mutations.
func (m *setMutations[ElementType]) WithAddedElements(elements Set[ElementType]) SetMutations[ElementType] {
	m.addedElements = elements

	return m
}

// WithDeletedElements sets the deleted elements of the mutations.
func (m *setMutations[ElementType]) WithDeletedElements(elements Set[ElementType]) SetMutations[ElementType] {
	m.deletedElements = elements

	return m
}

// AddedElements returns the elements that are supposed to be added.
func (m *setMutations[ElementType]) AddedElements() Set[ElementType] {
	return m.addedElements
}

// DeletedElements returns the elements that are supposed to be removed.
func (m *setMutations[ElementType]) DeletedElements() Set[ElementType] {
	return m.deletedElements
}

// IsEmpty returns true if the SetMutations instance is empty.
func (m *setMutations[ElementType]) IsEmpty() bool {
	return m.addedElements.IsEmpty() && m.deletedElements.IsEmpty()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region setArithmetic ////////////////////////////////////////////////////////////////////////////////////////////////

// setArithmetic is the default implementation of the SetArithmetic interface.
type setArithmetic[ElementType comparable] struct {
	// ShrinkingMap is used to keep track of the number of times an element is present.
	*shrinkingmap.ShrinkingMap[ElementType, int]
}

// newSetArithmetic creates a new setArithmetic instance.
func newSetArithmetic[ElementType comparable]() *setArithmetic[ElementType] {
	return &setArithmetic[ElementType]{
		ShrinkingMap: shrinkingmap.New[ElementType, int](),
	}
}

// Add adds the given mutations to the elements and returns the resulting net mutations for the set that are formed by
// tracking the elements that rise above the given threshold.
func (s *setArithmetic[ElementType]) Add(mutations SetMutations[ElementType], threshold ...int) SetMutations[ElementType] {
	m := NewSetMutations[ElementType]()

	mutations.AddedElements().Range(s.AddedElementsCollector(m, threshold...))
	mutations.DeletedElements().Range(s.SubtractedElementsCollector(m, threshold...))

	return m
}

// AddedElementsCollector returns a function that adds an element to the given mutations if its occurrence count reaches
// the given threshold (after the addition).
func (s *setArithmetic[ElementType]) AddedElementsCollector(mutations SetMutations[ElementType], threshold ...int) func(ElementType) {
	return s.elementsCollector(mutations.AddedElements(), mutations.DeletedElements(), true, lo.First(threshold, 1))
}

// Subtract subtracts the given mutations from the elements and returns the resulting net mutations for the set that are
// formed by tracking the elements that fall below the given threshold.
func (s *setArithmetic[ElementType]) Subtract(mutations SetMutations[ElementType], threshold ...int) SetMutations[ElementType] {
	m := NewSetMutations[ElementType]()

	mutations.AddedElements().Range(s.SubtractedElementsCollector(m, threshold...))
	mutations.DeletedElements().Range(s.AddedElementsCollector(m, threshold...))

	return m
}

// SubtractedElementsCollector returns a function that deletes an element from the given mutations if its occurrence count
// falls below the given threshold (after the subtraction).
func (s *setArithmetic[ElementType]) SubtractedElementsCollector(mutations SetMutations[ElementType], threshold ...int) func(ElementType) {
	return s.elementsCollector(mutations.DeletedElements(), mutations.AddedElements(), false, lo.First(threshold, 1))
}

// elementsCollector returns a function that collects elements in the given sets that pass the given threshold in either
// direction.
func (s *setArithmetic[ElementType]) elementsCollector(targetSet, opposingSet Set[ElementType], increase bool, threshold int) func(ElementType) {
	return func(element ElementType) {
		if s.Compute(element, func(currentValue int, _ bool) int {
			return currentValue + lo.Cond(increase, 1, -1)
		}) == lo.Cond(increase, threshold, threshold-1) && !opposingSet.Delete(element) {
			targetSet.Add(element)
		}
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
