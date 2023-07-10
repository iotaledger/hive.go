package set

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
)

// Set is a generic thread-safe collection of unique elements.
type Set[ElementType comparable] interface {
	// Writeable imports the write methods of the Set interface.
	Writeable[ElementType]

	// Readable imports the read methods of the Set interface.
	Readable[ElementType]
}

// New creates a new Set with the given elements.
func New[T comparable](elements ...T) Set[T] {
	return &set[T]{
		readable: newReadable(elements...),
	}
}

// set is the standard implementation of the Set interface.
type set[ElementType comparable] struct {
	// readable embeds the readable part of the set.
	*readable[ElementType]

	// applyMutex is used to make calls to Apply atomic.
	applyMutex sync.RWMutex
}

// Add adds a new element to the set and returns true if the element was not present in the set before.
func (s *set[ElementType]) Add(element ElementType) bool {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	return !lo.Return2(s.Set(element, types.Void))
}

// AddAll tries to add all elements to the set and returns the elements that have been added.
func (s *set[ElementType]) AddAll(elements Readable[ElementType]) (addedElements Set[ElementType]) {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

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
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

	return s.OrderedMap.Delete(element)
}

// DeleteAll deletes the given elements from the set.
func (s *set[ElementType]) DeleteAll(other Readable[ElementType]) (removedElements Set[ElementType]) {
	s.applyMutex.RLock()
	defer s.applyMutex.RUnlock()

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
	s.applyMutex.Lock()
	defer s.applyMutex.Unlock()

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
func (s *set[ElementType]) ToReadOnly() Readable[ElementType] {
	return s.readable
}
