package set

import (
	"github.com/iotaledger/hive.go/datastructure/set"
)

// type genericSet[T comparable] struct { implements a generic wrapper for a non-generic Set.
type genericSet[T comparable] struct {
	set.Set
}

// newGenericSetWrapper returns a new generic Set.
func newGenericSet[T comparable](s set.Set) *genericSet[T] {
	return &genericSet[T]{
		Set: s,
	}
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (set *genericSet[T]) Add(element T) bool {
	return set.Set.Add(element)
}

// Delete removes the element from the Set and returns true if it did exist.
func (set *genericSet[T]) Delete(element T) bool {
	return set.Set.Delete(element)
}

// Has returns true if the element exists in the Set.
func (set *genericSet[T]) Has(element T) bool {
	return set.Set.Has(element)
}

// ForEach iterates through the set and calls the callback for every element.
func (set *genericSet[T]) ForEach(callback func(element T)) {
	set.Set.ForEach(func(element interface{}) {
		callback(element.(T))
	})
}

// code contract - make sure the type implements the interface
var _ Set[int] = &genericSet[int]{}
