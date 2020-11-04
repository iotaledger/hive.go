package redblacktree

import (
	"github.com/iotaledger/hive.go/stringify"
)

// Node represents a Node in the Tree.
type Node struct {
	key         interface{}
	value       interface{}
	parent      *Node
	left        *Node
	right       *Node
	predecessor *Node
	successor   *Node
	isBlack     bool
}

// Key returns the key that is used to identify the Node.
func (n *Node) Key() interface{} {
	return n.key
}

// Value returns the value that is associated to the Node.
func (n *Node) Value() interface{} {
	return n.value
}

// Parent returns the parent of the Node (or nil if the Node is the root of the Tree).
func (n *Node) Parent() *Node {
	return n.parent
}

// Successor returns the Node with the next highest key (or nil if none exists).
func (n *Node) Successor() *Node {
	return n.successor
}

// Predecessor returns the Node with the next lower key (or nil if none exists).
func (n *Node) Predecessor() *Node {
	return n.predecessor
}

// IsBlack returns true if the Node is marked as black (colors are used for the self-balancing properties of the Tree)..
func (n *Node) IsBlack() bool {
	if n == nil {
		return true
	}

	return n.isBlack
}

// GrandParent returns the parent of the parent Node (or nil if it does not exist).
func (n *Node) GrandParent() *Node {
	if n != nil && n.parent != nil {
		return n.parent.parent
	}

	return nil
}

// Uncle returns the sibling of the parent Node.
func (n *Node) Uncle() *Node {
	if n == nil || n.parent == nil || n.parent.parent == nil {
		return nil
	}

	return n.parent.Sibling()
}

// Sibling returns the alternative Node sharing the same parent Node.
func (n *Node) Sibling() *Node {
	if n == nil || n.parent == nil {
		return nil
	}

	if n == n.parent.left {
		return n.parent.right
	}

	return n.parent.left
}

// Min returns the smallest of all descendants of the Node.
func (n *Node) Min() (node *Node) {
	if node = n; node == nil {
		return
	}

	for node.left != nil {
		node = node.left
	}

	return
}

// Max returns the largest of all descendants of the Node.
func (n *Node) Max() (node *Node) {
	if node = n; node == nil {
		return
	}

	for node.right != nil {
		node = node.right
	}

	return
}

// String returns a human readable version of the Node.
func (n *Node) String() string {
	return stringify.Struct("GetElement",
		stringify.StructField("key", n.key),
		stringify.StructField("value", n.value),
		stringify.StructField("left", n.left),
		stringify.StructField("right", n.right),
	)
}
