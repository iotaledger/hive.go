package reactive

import (
	"cmp"
)

// region SortedSet ////////////////////////////////////////////////////////////////////////////////////////////////////

// SortedSet is a reactive Set implementation that allows consumers to subscribe to its changes and that keeps a sorted
// perception of its elements.
type SortedSet[ElementType comparable] interface {
	// Set imports the methods of the Set interface.
	Set[ElementType]

	// Ascending returns a slice of all elements of the set in ascending order.
	Ascending() []ElementType

	// Descending returns a slice of all elements of the set in descending order.
	Descending() []ElementType

	// HeaviestElement returns the element with the heaviest weight.
	HeaviestElement() ReadableVariable[ElementType]

	// LightestElement returns the element with the lightest weight.
	LightestElement() ReadableVariable[ElementType]
}

// NewSortedSet creates a new SortedSet instance that sorts its elements by the given weightVariable. It is possible to
// optionally provide a tieBreaker function that is used to break ties between elements with the same weight.
func NewSortedSet[ElementType comparable, WeightType cmp.Ordered](weightVariable func(element ElementType) Variable[WeightType], optTieBreaker ...func(left, right ElementType) int) SortedSet[ElementType] {
	return newSortedSet(weightVariable, optTieBreaker...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
