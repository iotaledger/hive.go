package set

import (
	"sync"
)

// Mutations represents a set of mutations that can be applied to a Set atomically.
type Mutations[ElementType comparable] interface {
	// WithAddedElements is a setter for the added elements of the mutations.
	WithAddedElements(elements Set[ElementType]) Mutations[ElementType]

	// WithDeletedElements is a setter for the deleted elements of the mutations.
	WithDeletedElements(elements Set[ElementType]) Mutations[ElementType]

	// AddedElements returns the elements that are supposed to be added.
	AddedElements() Set[ElementType]

	// DeletedElements returns the elements that are supposed to be removed.
	DeletedElements() Set[ElementType]

	// IsEmpty returns true if the Mutations instance is empty.
	IsEmpty() bool
}

// NewMutations creates a new Mutations instance.
func NewMutations[ElementType comparable]() Mutations[ElementType] {
	return &mutations[ElementType]{
		addedElements:   New[ElementType](),
		deletedElements: New[ElementType](),
	}
}

// mutations is the default implementation of the Mutations interface.
type mutations[ElementType comparable] struct {
	// AddedElements are the elements that are supposed to be added.
	addedElements Set[ElementType]

	// deletedElements are the elements that are supposed to be removed.
	deletedElements Set[ElementType]

	// mutex is used to synchronize access to the mutations.
	mutex sync.RWMutex
}

// WithAddedElements is a setter for the added elements of the mutations.
func (m *mutations[ElementType]) WithAddedElements(elements Set[ElementType]) Mutations[ElementType] {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.addedElements.AddAll(elements)

	return m
}

// WithDeletedElements sets the deleted elements of the mutations.
func (m *mutations[ElementType]) WithDeletedElements(elements Set[ElementType]) Mutations[ElementType] {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.deletedElements.AddAll(elements)

	return m
}

// AddedElements returns the elements that are supposed to be added.
func (m *mutations[ElementType]) AddedElements() Set[ElementType] {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.addedElements
}

// DeletedElements returns the elements that are supposed to be removed.
func (m *mutations[ElementType]) DeletedElements() Set[ElementType] {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.deletedElements
}

// IsEmpty returns true if the Mutations instance is empty.
func (m *mutations[ElementType]) IsEmpty() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.addedElements.IsEmpty() && m.deletedElements.IsEmpty()
}

// code contract (make sure the type implements all required methods).
var _ Mutations[int] = new(mutations[int])
