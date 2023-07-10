package set

// Mutations represents a set of mutations that can be applied to a Set atomically.
type Mutations[ElementType comparable] interface {
	// WithAddedElements is a setter for the added elements of the mutations.
	WithAddedElements(elements Readable[ElementType]) Mutations[ElementType]

	// WithDeletedElements is a setter for the deleted elements of the mutations.
	WithDeletedElements(elements Readable[ElementType]) Mutations[ElementType]

	// AddedElements returns the elements that are supposed to be added.
	AddedElements() Set[ElementType]

	// DeletedElements returns the elements that are supposed to be removed.
	DeletedElements() Set[ElementType]

	// IsEmpty returns true if the Mutations instance is empty.
	IsEmpty() bool
}

// NewMutations creates a new Mutations instance.
func NewMutations[ElementType comparable](elements ...ElementType) Mutations[ElementType] {
	return &mutations[ElementType]{
		addedElements:   New[ElementType](elements...),
		deletedElements: New[ElementType](),
	}
}

// mutations is the default implementation of the Mutations interface.
type mutations[ElementType comparable] struct {
	// AddedElements are the elements that are supposed to be added.
	addedElements Set[ElementType]

	// deletedElements are the elements that are supposed to be removed.
	deletedElements Set[ElementType]
}

// WithAddedElements is a setter for the added elements of the mutations.
func (m *mutations[ElementType]) WithAddedElements(elements Readable[ElementType]) Mutations[ElementType] {
	m.addedElements.AddAll(elements)

	return m
}

// WithDeletedElements sets the deleted elements of the mutations.
func (m *mutations[ElementType]) WithDeletedElements(elements Readable[ElementType]) Mutations[ElementType] {
	m.deletedElements.AddAll(elements)

	return m
}

// AddedElements returns the elements that are supposed to be added.
func (m *mutations[ElementType]) AddedElements() Set[ElementType] {
	return m.addedElements
}

// DeletedElements returns the elements that are supposed to be removed.
func (m *mutations[ElementType]) DeletedElements() Set[ElementType] {
	return m.deletedElements
}

// IsEmpty returns true if the Mutations instance is empty.
func (m *mutations[ElementType]) IsEmpty() bool {
	return m.addedElements.IsEmpty() && m.deletedElements.IsEmpty()
}

// code contract (make sure the type implements all required methods).
var _ Mutations[int] = new(mutations[int])
