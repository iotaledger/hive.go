package set

import (
	"sync"
)

// Mutations represents a set of mutations that can be applied to a Set atomically.
type Mutations[T comparable] interface {
	// WithAddedElements is a setter for the added elements of the mutations.
	WithAddedElements(elements ReadOnly[T]) Mutations[T]

	// WithDeletedElements is a setter for the deleted elements of the mutations.
	WithDeletedElements(elements ReadOnly[T]) Mutations[T]

	// AddedElements returns the elements that are supposed to be added.
	AddedElements() ReadOnly[T]

	// DeletedElements returns the elements that are supposed to be removed.
	DeletedElements() ReadOnly[T]

	// IsEmpty returns true if the Mutations instance is empty.
	IsEmpty() bool
}

// NewMutations creates a new Mutations instance.
func NewMutations[T comparable]() Mutations[T] {
	return &mutations[T]{
		addedElements:   New[T](),
		deletedElements: New[T](),
	}
}

// mutations is the default implementation of the Mutations interface.
type mutations[T comparable] struct {
	// AddedElements are the elements that are supposed to be added.
	addedElements ReadOnly[T]

	// deletedElements are the elements that are supposed to be removed.
	deletedElements ReadOnly[T]

	// mutex is used to synchronize access to the mutations.
	mutex sync.RWMutex
}

// WithAddedElements is a setter for the added elements of the mutations.
func (m *mutations[T]) WithAddedElements(elements ReadOnly[T]) Mutations[T] {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.addedElements = elements

	return m
}

// WithDeletedElements sets the deleted elements of the mutations.
func (m *mutations[T]) WithDeletedElements(elements ReadOnly[T]) Mutations[T] {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.deletedElements = elements

	return m
}

// AddedElements returns the elements that are supposed to be added.
func (m *mutations[T]) AddedElements() ReadOnly[T] {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.addedElements
}

// DeletedElements returns the elements that are supposed to be removed.
func (m *mutations[T]) DeletedElements() ReadOnly[T] {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.deletedElements
}

// IsEmpty returns true if the Mutations instance is empty.
func (m *mutations[T]) IsEmpty() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.addedElements.IsEmpty() && m.deletedElements.IsEmpty()
}

// code contract (make sure the type implements all required methods).
var _ Mutations[int] = new(mutations[int])
