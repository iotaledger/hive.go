package set

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
)

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

	// Apply tries to apply the given mutations to the set atomically and returns the applied mutations.
	Apply(mutations Mutations[ElementType]) (appliedMutations Mutations[ElementType])

	// Decode decodes the set from a byte slice.
	Decode(b []byte) (bytesRead int, err error)

	// ToReadOnly returns a read-only version of the set.
	ToReadOnly() ReadOnly[ElementType]

	// ReadOnly imports the read methods from ReadOnly.
	ReadOnly[ElementType]
}

// New creates a new Set with the given elements.
func New[T comparable](elements ...T) Set[T] {
	s := &set[T]{
		readOnly: &readOnly[T]{
			OrderedMap: orderedmap.New[T, types.Empty](),
		},
	}

	for _, element := range elements {
		s.OrderedMap.Set(element, types.Void)
	}

	return s
}

// set implements the Set interface.
type set[ElementType comparable] struct {
	*readOnly[ElementType]

	mutex sync.RWMutex
}

// Add adds a new element to the set and returns true if the element was not present in the set before.
func (s *set[ElementType]) Add(element ElementType) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return !lo.Return2(s.Set(element, types.Void))
}

// AddAll tries to add all elements to the set and returns the elements that have been added.
func (s *set[ElementType]) AddAll(elements ReadOnly[ElementType]) (addedElements Set[ElementType]) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.OrderedMap.Delete(element)
}

// DeleteAll deletes the given elements from the set.
func (s *set[ElementType]) DeleteAll(other ReadOnly[ElementType]) (removedElements Set[ElementType]) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	removedElements = New[ElementType]()
	_ = other.ForEach(func(element ElementType) (err error) {
		if s.Delete(element) {
			removedElements.Add(element)
		}

		return nil
	})

	return removedElements
}

// Apply tries to apply the given mutations to the set atomically and returns the applied mutations.
func (s *set[ElementType]) Apply(mutations Mutations[ElementType]) (appliedMutations Mutations[ElementType]) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addedElements := New[ElementType]()
	mutations.AddedElements().Range(func(element ElementType) {
		if !lo.Return2(s.Set(element, types.Void)) {
			addedElements.Add(element)
		}
	})

	removedElements := New[ElementType]()
	mutations.DeletedElements().Range(func(element ElementType) {
		if s.OrderedMap.Delete(element) {
			removedElements.Add(element)
		}
	})

	return NewMutations[ElementType]().WithAddedElements(addedElements).WithDeletedElements(removedElements)
}

// ToReadOnly returns a read-only version of the set.
func (s *set[ElementType]) ToReadOnly() ReadOnly[ElementType] {
	return s.readOnly
}

// code contract (make sure the type implements all required methods).
var _ Set[int] = new(set[int])
