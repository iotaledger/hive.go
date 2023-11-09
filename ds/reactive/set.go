package reactive

import (
	"github.com/iotaledger/hive.go/ds"
)

// region Set ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// Set is a reactive Set implementation that allows consumers to subscribe to its changes.
type Set[ElementType comparable] interface {
	// WriteableSet imports the write methods of the Set interface.
	ds.WriteableSet[ElementType]

	// ReadableSet imports the read methods of the Set interface.
	ReadableSet[ElementType]
}

// NewSet creates a new Set with the given elements.
func NewSet[T comparable](elements ...T) Set[T] {
	return newSet(elements...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ReadableSet //////////////////////////////////////////////////////////////////////////////////////////////////

// ReadableSet is a reactive Set implementation that allows consumers to subscribe to its value.
type ReadableSet[ElementType comparable] interface {
	// OnUpdate registers the given callback that is triggered when the value changes.
	OnUpdate(callback func(appliedMutations ds.SetMutations[ElementType]), triggerWithInitialZeroValue ...bool) (unsubscribe func())

	// SubtractReactive returns a new set that will automatically be updated to always hold all elements of the current
	// set minus the elements of the other sets.
	SubtractReactive(others ...ReadableSet[ElementType]) Set[ElementType]

	// WithElements is a utility function that allows to set up dynamic behavior based on the elements of the Set which
	// is torn down once the element is removed gi(or the returned teardown function is called). It accepts an optional
	// condition that has to be satisfied for the setup function to be called.
	WithElements(setup func(element ElementType) (teardown func()), condition ...func(ElementType) bool) (teardown func())

	// ReadableSet imports the read methods of the Set interface.
	ds.ReadableSet[ElementType]
}

// NewReadableSet creates a new ReadableSet with the given elements.
func NewReadableSet[ElementType comparable](elements ...ElementType) ReadableSet[ElementType] {
	return newReadableSet(elements...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region DerivedSet ///////////////////////////////////////////////////////////////////////////////////////////////////

// DerivedSet is a reactive Set implementation that allows consumers to subscribe to its changes and that inherits its
// elements from other sets.
type DerivedSet[ElementType comparable] interface {
	Set[ElementType]

	InheritFrom(sources ...ReadableSet[ElementType]) (unsubscribe func())
}

// NewDerivedSet creates a new DerivedSet with the given elements.
func NewDerivedSet[ElementType comparable]() DerivedSet[ElementType] {
	return newDerivedSet[ElementType]()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
