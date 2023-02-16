package thresholdmap

import "github.com/emirpasic/gods/trees/redblacktree"

// Element is a wrapper for the Node used in the underlying red-black RedBlackTree.
type Element[K any, V any] struct {
	*redblacktree.Node
}

// Key returns the Key of the Element.
func (e *Element[K, V]) Key() K {
	return e.Node.Key.(K)
}

// Value returns the Value of the Element.
func (e *Element[K, V]) Value() V {
	return e.Node.Value.(V)
}
