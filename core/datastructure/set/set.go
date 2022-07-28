package set

// Set is a collection of elements.
type Set interface {
	// Add adds a new element to the Set and returns true if the element was not present in the set before.
	Add(element interface{}) bool

	// Delete removes the element from the Set and returns true if it did exist.
	Delete(element interface{}) bool

	// Has returns true if the element exists in the Set.
	Has(element interface{}) bool

	// ForEach iterates through the set and calls the callback for every element.
	ForEach(callback func(element interface{}))

	// Clear removes all elements from the Set.
	Clear()

	// Size returns the size of the Set.
	Size() int
}

// New returns a new Set that is thread safe if the optional threadSafe parameter is set to true.
func New(threadSafe ...bool) Set {
	if len(threadSafe) >= 1 && threadSafe[0] {
		return newThreadSafeSet()
	}

	return newSimpleSet()
}
