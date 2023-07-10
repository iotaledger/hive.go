package set

// Writeable bundles all write methods of the Set interface.
type Writeable[ElementType comparable] interface {
	// Add adds a new element to the set and returns true if the element was not present in the set before.
	Add(element ElementType) bool

	// AddAll adds all elements to the set and returns true if any element has been added.
	AddAll(elements Readable[ElementType]) (addedElements Set[ElementType])

	// Delete deletes the given element from the set.
	Delete(element ElementType) bool

	// DeleteAll deletes the given elements from the set.
	DeleteAll(other Readable[ElementType]) (removedElements Set[ElementType])

	// Apply tries to apply the given mutations to the set atomically and returns the applied mutations.
	Apply(mutations Mutations[ElementType]) (appliedMutations Mutations[ElementType])

	// Decode decodes the set from a byte slice.
	Decode(b []byte) (bytesRead int, err error)

	// ToReadOnly returns a read-only version of the set.
	ToReadOnly() Readable[ElementType]
}
