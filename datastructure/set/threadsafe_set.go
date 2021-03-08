package set

import (
	"sync"

	"github.com/iotaledger/hive.go/types"
)

// threadSafeSet implements a thread safe Set.
type threadSafeSet struct {
	elements      map[interface{}]types.Empty
	elementsMutex sync.RWMutex
}

// newThreadSafeSet returns a new thread safe Set.
func newThreadSafeSet() *threadSafeSet {
	return &threadSafeSet{
		elements: make(map[interface{}]types.Empty),
	}
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (set *threadSafeSet) Add(element interface{}) bool {
	set.elementsMutex.RLock()
	if _, elementExists := set.elements[element]; elementExists {
		set.elementsMutex.RUnlock()

		return false
	}

	set.elementsMutex.RUnlock()
	set.elementsMutex.Lock()
	defer set.elementsMutex.Unlock()

	if _, elementExists := set.elements[element]; elementExists {
		return false
	}

	set.elements[element] = types.Void

	return true
}

// Delete removes the element from the Set and returns true if it did exist.
func (set *threadSafeSet) Delete(element interface{}) bool {
	set.elementsMutex.RLock()
	if _, exists := set.elements[element]; !exists {
		set.elementsMutex.RUnlock()

		return false
	}
	set.elementsMutex.RUnlock()

	set.elementsMutex.Lock()
	defer set.elementsMutex.Unlock()
	if _, exists := set.elements[element]; !exists {
		return false
	}

	delete(set.elements, element)

	return true
}

// Has returns true if the element exists in the Set.
func (set *threadSafeSet) Has(element interface{}) bool {
	set.elementsMutex.RLock()
	defer set.elementsMutex.RUnlock()

	_, exists := set.elements[element]

	return exists
}

// ForEach iterates through the set and calls the callback for every element.
func (set *threadSafeSet) ForEach(callback func(element interface{})) {
	set.elementsMutex.RLock()
	copiedElements := make(map[interface{}]types.Empty)
	for element := range set.elements {
		copiedElements[element] = types.Void
	}
	set.elementsMutex.RUnlock()

	for element := range copiedElements {
		callback(element)
	}
}

// Clear removes all elements from the Set.
func (set *threadSafeSet) Clear() {
	set.elementsMutex.Lock()
	defer set.elementsMutex.Unlock()

	set.elements = make(map[interface{}]types.Empty)
}

// Size returns the size of the Set.
func (set *threadSafeSet) Size() int {
	set.elementsMutex.RLock()
	defer set.elementsMutex.RUnlock()

	return len(set.elements)
}

// code contract - make sure the type implements the interface
var _ Set = &threadSafeSet{}
