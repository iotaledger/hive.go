package set

import "github.com/iotaledger/hive.go/types"

// simpleSet implements a non-thread safe Set.
type simpleSet struct {
	elements map[interface{}]types.Empty
}

// newSimpleSet returns a new non-thread safe Set.
func newSimpleSet() Set {
	return &simpleSet{
		elements: make(map[interface{}]types.Empty),
	}
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (set *simpleSet) Add(element interface{}) bool {
	if _, elementExists := set.elements[element]; elementExists {
		return false
	}

	set.elements[element] = types.Void

	return true
}

// Delete removes the element from the Set and returns true if it did exist.
func (set *simpleSet) Delete(element interface{}) bool {
	_, elementExists := set.elements[element]
	if elementExists {
		delete(set.elements, element)
	}

	return elementExists
}

// Has returns true if the element exists in the Set.
func (set *simpleSet) Has(element interface{}) bool {
	_, elementExists := set.elements[element]

	return elementExists
}

// ForEach iterates through the set and calls the callback for every element.
func (set *simpleSet) ForEach(callback func(element interface{})) {
	for element := range set.elements {
		callback(element)
	}
}

// Clear removes all elements from the Set.
func (set *simpleSet) Clear() {
	set.elements = make(map[interface{}]types.Empty)
}

// Size returns the size of the Set.
func (set *simpleSet) Size() int {
	return len(set.elements)
}

// code contract - make sure the type implements the interface
var _ Set = &simpleSet{}
