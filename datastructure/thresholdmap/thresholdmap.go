package thresholdmap

import (
	"github.com/iotaledger/hive.go/datastructure/genericcomparator"
	"github.com/iotaledger/hive.go/datastructure/redblacktree"
)

// region ThresholdMap /////////////////////////////////////////////////////////////////////////////////////////////////

// ThresholdMap is a data structure that allows to map keys bigger or lower than a certain threshold to a given value.
type ThresholdMap struct {
	mode Mode
	tree *redblacktree.RedBlackTree
}

// New returns a ThresholdMap that operates in the given Mode and that can also receive an optional comparator function
// to support custom key types.
func New(mode Mode, optionalComparator ...genericcomparator.Type) *ThresholdMap {
	if len(optionalComparator) >= 1 {
		return &ThresholdMap{
			mode: mode,
			tree: redblacktree.New(optionalComparator[0]),
		}
	}

	return &ThresholdMap{
		mode: mode,
		tree: redblacktree.New(),
	}
}

// Set adds a new threshold that maps all keys >= or <= (depending on the Mode) the value given by key to a certain
// value.
func (t *ThresholdMap) Set(key interface{}, value interface{}) {
	t.tree.Set(key, value)
}

// Get returns the value of the next higher or lower existing threshold (depending on the mode) and a flag that
// indicates if there is a threshold that covers the given value.
func (t *ThresholdMap) Get(key interface{}) (value interface{}, exists bool) {
	var foundNode *redblacktree.Node
	switch t.mode {
	case UpperThresholdMode:
		foundNode = t.tree.Ceiling(key)
	case LowerThresholdMode:
		foundNode = t.tree.Floor(key)
	default:
		panic("unsupported mode")
	}
	exists = foundNode != nil

	if exists {
		value = foundNode.Value()
	}

	return
}

// Floor returns the largest key that is <= the given key (and a boolean value indicating if it exists).
func (t *ThresholdMap) Floor(key interface{}) (floor interface{}, exists bool) {
	if node := t.tree.Floor(key); node != nil {
		return node.Key(), true
	}

	return nil, false
}

// Ceiling returns the smallest key that is >= the given key (and a boolean value indicating if it exists).
func (t *ThresholdMap) Ceiling(key interface{}) (floor interface{}, exists bool) {
	if node := t.tree.Ceiling(key); node != nil {
		return node.Key(), true
	}

	return nil, false
}

// Delete removes a threshold from the map.
func (t *ThresholdMap) Delete(key interface{}) (element *Element, success bool) {
	node, success := t.tree.Delete(key)
	element = t.wrapNode(node)

	return
}

// Keys returns a list of thresholds that have been set in the map.
func (t *ThresholdMap) Keys() []interface{} {
	return t.tree.Keys()
}

// Values returns a list of values that are associated to the thresholds in the map.
func (t *ThresholdMap) Values() []interface{} {
	return t.tree.Values()
}

// GetElement returns the Element that is used to store the next higher or lower threshold (depending on the mode)
// belonging to the given key (or nil if none exists).
func (t *ThresholdMap) GetElement(key interface{}) *Element {
	switch t.mode {
	case UpperThresholdMode:
		return t.wrapNode(t.tree.Ceiling(key))
	default:
		return t.wrapNode(t.tree.Floor(key))
	}
}

// MinElement returns the smallest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap) MinElement() *Element {
	return t.wrapNode(t.tree.Min())
}

// MaxElement returns the largest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap) MaxElement() *Element {
	return t.wrapNode(t.tree.Max())
}

// DeleteElement removes the given Element from the map.
func (t *ThresholdMap) DeleteElement(element *Element) {
	if element == nil {
		return
	}

	t.tree.DeleteNode(element.Node)
}

// ForEach provides a callback based iterator that iterates through all Elements in the map.
func (t *ThresholdMap) ForEach(iterator func(node *Element) bool) {
	t.tree.ForEach(func(node *redblacktree.Node) bool {
		return iterator(t.wrapNode(node))
	})
}

// Iterator returns an Iterator object that can be used to manually iterate through the Elements in the map. It accepts
// an optional starting Element where the iteration begins.
func (t *ThresholdMap) Iterator(optionalStartingNode ...*Element) *Iterator {
	if len(optionalStartingNode) >= 1 {
		return NewIterator(optionalStartingNode[0])
	}

	return NewIterator(t.wrapNode(t.tree.Min()))
}

// Size returns the amount of thresholds that are stored in the map.
func (t *ThresholdMap) Size() int {
	return t.tree.Size()
}

// Empty returns true of the map has no thresholds.
func (t *ThresholdMap) Empty() bool {
	return t.tree.Empty()
}

// Clear removes all Elements from the map.
func (t *ThresholdMap) Clear() {
	t.tree.Clear()
}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (t *ThresholdMap) wrapNode(node *redblacktree.Node) (element *Element) {
	if node == nil {
		return
	}

	return &Element{node}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Element //////////////////////////////////////////////////////////////////////////////////////////////////////

// Element is a wrapper for the Node used in the underlying red-black RedBlackTree.
type Element struct {
	*redblacktree.Node
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Mode /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Mode encodes different modes of function for the ThresholdMap that specifies if the defines keys act as upper or
// lower thresholds.
type Mode bool

const (
	// LowerThresholdMode interprets the keys of the ThresholdMap as lower thresholds which means that querying the map
	// will return the value of the largest node whose key is <= than the queried value.
	LowerThresholdMode = true

	// UpperThresholdMode interprets the keys of the ThresholdMap as upper thresholds which means that querying the map
	// will return the value of the smallest node whose key is >= than the queried value.
	UpperThresholdMode = false
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Iterator /////////////////////////////////////////////////////////////////////////////////////////////////////

// Iterator is an object that allows to iterate over the ThresholdMap by providing methods to walk through the map in a
// deterministic order.
type Iterator struct {
	start   *Element
	current *Element
	state   IteratorState
}

// NewIterator is the constructor of the Iterator that takes the starting Element as its parameter.
func NewIterator(startingElement *Element) *Iterator {
	return &Iterator{
		start:   startingElement,
		current: startingElement,
	}
}

// State returns the current IteratorState that the Iterator is in.
func (i *Iterator) State() IteratorState {
	return i.state
}

// HasNext returns true if there is another Element after the previously retrieved Element that can be requested via the
// Next method.
func (i *Iterator) HasNext() bool {
	switch i.state {
	case InitialState:
		fallthrough
	case LeftEndReachedState:
		return i.current != nil
	case IterationStartedState:
		return i.current.Successor() != nil
	}

	return false
}

// HasPrev returns true if there is another Element before the previously retrieved Element that can be requested via
// the Prev method.
func (i *Iterator) HasPrev() bool {
	switch i.state {
	case InitialState:
		fallthrough
	case RightEndReachedState:
		return i.current != nil
	case IterationStartedState:
		return i.current.Predecessor() != nil
	}

	return false
}

// Next returns the next Element in the Iterator and advances the internal pointer. The method panics if there is no
// next Element that can be retrieved (always use HasNext to check if another Element can be requested).
func (i *Iterator) Next() *Element {
	if i.state == RightEndReachedState || i.current == nil {
		panic("no next element found in iterator")
	}

	if i.state == IterationStartedState {
		i.current = i.wrapNode(i.current.Successor())
	}

	if i.current.Successor() == nil {
		i.state = RightEndReachedState
	} else {
		i.state = IterationStartedState
	}

	return i.current
}

// Prev returns the previous Element in the Iterator and moves back the internal pointer. The method panics if there is
// no previous Element that can be retrieved (always use HasPrev to check if another Element can be requested).
func (i *Iterator) Prev() *Element {
	if i.state == LeftEndReachedState || i.current == nil {
		panic("no next element found in iterator")
	}

	if i.state == IterationStartedState {
		i.current = i.wrapNode(i.current.Predecessor())
	}

	if i.current.Predecessor() == nil {
		i.state = LeftEndReachedState
	} else {
		i.state = IterationStartedState
	}

	return i.current
}

// Reset resets the Iterator to its initial Element.
func (i *Iterator) Reset() {
	i.current = i.start
	i.state = InitialState
}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (i *Iterator) wrapNode(node *redblacktree.Node) (element *Element) {
	if node == nil {
		return
	}

	return &Element{node}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

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
