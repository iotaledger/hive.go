package thresholdmap

import "github.com/emirpasic/gods/trees/redblacktree"

// Iterator is an object that allows to iterate over the ThresholdMap by providing methods to walk through the map in a
// deterministic order.
type Iterator[K any, V any] struct {
	start   *Element[K, V]
	current *Element[K, V]
	state   IteratorState
}

// NewIterator is the constructor of the Iterator that takes the starting Element as its parameter.
func NewIterator[K any, V any](startingElement *Element[K, V]) *Iterator[K, V] {
	return &Iterator[K, V]{
		start:   startingElement,
		current: startingElement,
	}
}

// State returns the current IteratorState that the Iterator is in.
func (i *Iterator[K, V]) State() IteratorState {
	return i.state
}

// HasNext returns true if there is another Element after the previously retrieved Element that can be requested via the
// Next method.
func (i *Iterator[K, V]) HasNext() bool {
	switch i.state {
	case InitialState:
		fallthrough
	case LeftEndReachedState:
		return i.current != nil
	case IterationStartedState:
		return i.current.Right != nil
	}

	return false
}

// HasPrev returns true if there is another Element before the previously retrieved Element that can be requested via
// the Prev method.
func (i *Iterator[K, V]) HasPrev() bool {
	switch i.state {
	case InitialState:
		fallthrough
	case RightEndReachedState:
		return i.current != nil
	case IterationStartedState:
		return i.current.Left != nil
	}

	return false
}

// Next returns the next Element in the Iterator and advances the internal pointer. The method panics if there is no
// next Element that can be retrieved (always use HasNext to check if another Element can be requested).
func (i *Iterator[K, V]) Next() *Element[K, V] {
	if i.state == RightEndReachedState || i.current == nil {
		panic("no next element found in iterator")
	}

	if i.state == IterationStartedState {
		i.current = i.wrapNode(i.current.Right)
	}

	if i.current.Right == nil {
		i.state = RightEndReachedState
	} else {
		i.state = IterationStartedState
	}

	return i.current
}

// Prev returns the previous Element in the Iterator and moves back the internal pointer. The method panics if there is
// no previous Element that can be retrieved (always use HasPrev to check if another Element can be requested).
func (i *Iterator[K, V]) Prev() *Element[K, V] {
	if i.state == LeftEndReachedState || i.current == nil {
		panic("no next element found in iterator")
	}

	if i.state == IterationStartedState {
		i.current = i.wrapNode(i.current.Left)
	}

	if i.current.Left == nil {
		i.state = LeftEndReachedState
	} else {
		i.state = IterationStartedState
	}

	return i.current
}

// Reset resets the Iterator to its initial Element.
func (i *Iterator[K, V]) Reset() {
	i.current = i.start
	i.state = InitialState
}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (i *Iterator[K, V]) wrapNode(node *redblacktree.Node) (element *Element[K, V]) {
	if node == nil {
		return
	}

	return &Element[K, V]{node}
}

// region IteratorState ////////////////////////////////////////////////////////////////////////////////////////////////

// IteratorState represents the state of the Iterator that is used to track where in the set of contained Elements the
// pointer is currently located.
type IteratorState int

const (
	// InitialState is the state of the Iterator before the first Element has been retrieved.
	InitialState IteratorState = iota

	// IterationStartedState is the state of the Iterator after the first Element has been retrieved and before we have
	// reached either the first or the last Element.
	IterationStartedState

	// LeftEndReachedState is the state of the Iterator after we have reached the smallest Element.
	LeftEndReachedState

	// RightEndReachedState is the state of the Iterator after we have reached the largest Element.
	RightEndReachedState
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
