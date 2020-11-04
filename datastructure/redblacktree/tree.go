package redblacktree

import (
	"github.com/iotaledger/hive.go/datastructure/genericcomparator"
	"github.com/iotaledger/hive.go/stringify"
)

// Tree represents a self balancing binary search tree, that can be used to red-black tree which
type Tree struct {
	root       *Node
	min        *Node
	max        *Node
	comparator genericcomparator.Type
	size       int
}

// New creates a new red-black Tree that uses the given comparator (or the default Comparator if the parameter is
// omitted) to compare the keys used to identify the nodes.
func New(optionalComparator ...genericcomparator.Type) *Tree {
	if len(optionalComparator) >= 1 {
		return &Tree{
			comparator: optionalComparator[0],
		}
	}

	return &Tree{
		comparator: genericcomparator.Comparator,
	}
}

// Set inserts or updates a Node in the Tree and returns it together with a flag that indicates if it was inserted.
func (t *Tree) Set(key interface{}, value interface{}) (node *Node, inserted bool) {
	if t.root == nil {
		node = &Node{key: key, value: value, isBlack: true}
		t.root = node
		t.min = node
		t.max = node

		inserted = true
		t.size++
		return
	}

	var predecessor, successor *Node
InsertNode:
	for currentNode := t.root; ; {
		switch t.comparator(key, currentNode.key) {
		case 0:
			currentNode.key = key
			currentNode.value = value

			node = currentNode
			return
		case -1:
			if currentNode.left == nil {
				successor = currentNode
				node = &Node{parent: currentNode, key: key, value: value, isBlack: false}
				currentNode.left = node
				break InsertNode
			}

			successor = currentNode
			currentNode = currentNode.left
		case 1:
			if currentNode.right == nil {
				predecessor = currentNode
				node = &Node{parent: currentNode, key: key, value: value, isBlack: false}
				currentNode.right = node
				break InsertNode
			}

			predecessor = currentNode
			currentNode = currentNode.right
		}
	}

	node.predecessor = predecessor
	node.successor = successor
	if predecessor != nil {
		predecessor.successor = node
	} else {
		t.min = node
	}
	if successor != nil {
		successor.predecessor = node
	} else {
		t.max = node
	}
	t.insertCase1(node)

	inserted = true
	t.size++
	return
}

// Get returns the value stored with the given key (or nil if the value does not exist with found being false).
func (t *Tree) Get(key interface{}) (value interface{}, found bool) {
	if node := t.Node(key); node != nil {
		value = node.value
		found = true
		return
	}

	return
}

// Delete removes a Node belonging to the given key from the Tree and returns it (if it existed) together with a flag
// that indicates if it existed.
func (t *Tree) Delete(key interface{}) (node *Node, success bool) {
	node = t.Node(key)
	if success = node != nil; !success {
		return
	}

	t.DeleteNode(node)

	return
}

// DeleteNode removes the Node from the Tree (which can be i.e. useful for modifying the Tree while iterating.
func (t *Tree) DeleteNode(node *Node) {
	if node.predecessor != nil {
		node.predecessor.successor = node.successor
	} else {
		t.min = node.successor
	}
	if node.successor != nil {
		node.successor.predecessor = node.predecessor
	} else {
		t.max = node.predecessor
	}

	if node.left != nil && node.right != nil {
		pred := node.left.Max()
		node.key = pred.key
		node.value = pred.value
		node = pred
	}

	var child *Node
	if node.left == nil || node.right == nil {
		if node.right == nil {
			child = node.left
		} else {
			child = node.right
		}
		if node.isBlack {
			node.isBlack = child.IsBlack()
			t.deleteCase1(node)
		}
		t.swapNodes(node, child)
		if node.parent == nil && child != nil {
			child.isBlack = true
		}
	}
	t.size--
}

// ForEach iterates through the Nodes of the Tree in ascending order and calls the iterator function for each Node. The
// iteration aborts as soon as the iterator function returns false.
func (t *Tree) ForEach(iterator func(node *Node) bool) {
	abortIteration := false
	for currentNode := t.Min(); currentNode != nil && !abortIteration; currentNode = currentNode.successor {
		abortIteration = !iterator(currentNode)
	}
}

// Keys returns an ordered list of keys that are stored in the Tree.
func (t *Tree) Keys() (keys []interface{}) {
	keys = make([]interface{}, 0, t.size)
	for currentNode := t.Min(); currentNode != nil; currentNode = currentNode.successor {
		keys = append(keys, currentNode.key)
	}

	return
}

// Values returns an ordered list of values that are stored in the Tree.
func (t *Tree) Values() (values []interface{}) {
	values = make([]interface{}, 0, t.size)
	for currentNode := t.Min(); currentNode != nil; currentNode = currentNode.successor {
		values = append(values, currentNode.value)
	}

	return
}

// Node returns the Node that belongs to the given key (or nil if it doesn't exist).
func (t *Tree) Node(key interface{}) (node *Node) {
	for node = t.root; node != nil; {
		switch t.comparator(key, node.key) {
		case 0:
			return
		case -1:
			node = node.left
		case 1:
			node = node.right
		}
	}

	return
}

// Min returns the Node with the smallest key (or nil if the Tree is empty).
func (t *Tree) Min() *Node {
	return t.root.Min()
}

// Max returns the Node with the largest key (or nil if the Tree is empty).
func (t *Tree) Max() *Node {
	return t.root.Max()
}

// Floor returns the Node with the largest key that is <= the given key (or nil if no floor was found).
func (t *Tree) Floor(key interface{}) (floor *Node) {
	for node := t.root; node != nil; {
		switch t.comparator(key, node.key) {
		case 0:
			floor = node
			return
		case -1:
			node = node.left
		case 1:
			floor = node
			node = node.right
		}
	}

	return
}

// Ceiling returns the Node with the smallest key that is >= the given key (or nil if no ceiling was found).
func (t *Tree) Ceiling(key interface{}) (ceiling *Node) {
	for node := t.root; node != nil; {
		switch t.comparator(key, node.key) {
		case 0:
			ceiling = node
			return
		case -1:
			ceiling = node
			node = node.left
		case 1:
			node = node.right
		}
	}

	return
}

// Size returns the amount of Nodes in the Tree.
func (t *Tree) Size() int {
	return t.size
}

// Empty returns true if the Tree has no Nodes.
func (t *Tree) Empty() bool {
	return t.size == 0
}

// Clear removes all Nodes from the Tree.
func (t *Tree) Clear() {
	t.root = nil
	t.min = nil
	t.max = nil
	t.size = 0
}

// String returns a human readable version of the Tree.
func (t *Tree) String() string {
	return stringify.Struct("Tree",
		stringify.StructField("size", t.size),
		stringify.StructField("root", t.root),
	)
}

// insertCase1 is an internal utility function that implements the 1st insert case.
func (t *Tree) insertCase1(node *Node) {
	if node.parent == nil {
		node.isBlack = true
		return
	}

	t.insertCase2(node)
}

// insertCase2 is an internal utility function that implements the 2nd insert case.
func (t *Tree) insertCase2(node *Node) {
	if node.parent.IsBlack() {
		return
	}

	t.insertCase3(node)
}

// insertCase3 is an internal utility function that implements the 3rd insert case.
func (t *Tree) insertCase3(node *Node) {
	uncle := node.Uncle()
	if !uncle.IsBlack() {
		node.parent.isBlack = true
		uncle.isBlack = true

		grandParent := node.GrandParent()
		grandParent.isBlack = false
		t.insertCase1(grandParent)
		return
	}

	t.insertCase4(node)
}

// insertCase4 is an internal utility function that implements the 4th insert case.
func (t *Tree) insertCase4(node *Node) {
	parent := node.parent
	grandParent := node.GrandParent()

	switch {
	case node == parent.right && parent == grandParent.left:
		t.rotateLeft(parent)
		node = node.left
	case node == parent.left && parent == grandParent.right:
		t.rotateRight(parent)
		node = node.right
	}

	t.insertCase5(node)
}

// insertCase5 is an internal utility function that implements the 5th insert case.
func (t *Tree) insertCase5(node *Node) {
	parent := node.parent
	grandParent := node.GrandParent()

	parent.isBlack = true
	grandParent.isBlack = false

	if node == parent.left && parent == grandParent.left {
		t.rotateRight(grandParent)
	} else if node == parent.right && parent == grandParent.right {
		t.rotateLeft(grandParent)
	}
}

// deleteCase1 is an internal utility function that implements the 1st delete case.
func (t *Tree) deleteCase1(node *Node) {
	if node.parent == nil {
		return
	}

	t.deleteCase2(node)
}

// deleteCase2 is an internal utility function that implements the 2nd delete case.
func (t *Tree) deleteCase2(node *Node) {
	if sibling := node.Sibling(); !sibling.IsBlack() {
		parent := node.parent

		parent.isBlack = false
		sibling.isBlack = true

		switch node {
		case parent.left:
			t.rotateLeft(parent)
		case parent.right:
			t.rotateRight(parent)
		}
	}

	t.deleteCase3(node)
}

// deleteCase3 is an internal utility function that implements the 3rd delete case.
func (t *Tree) deleteCase3(node *Node) {
	parent := node.parent
	sibling := node.Sibling()

	if parent.IsBlack() && sibling.IsBlack() && sibling.left.IsBlack() && sibling.right.IsBlack() {
		sibling.isBlack = false
		t.deleteCase1(parent)
		return
	}

	t.deleteCase4(node)
}

// deleteCase4 is an internal utility function that implements the 4th delete case.
func (t *Tree) deleteCase4(node *Node) {
	parent := node.parent
	sibling := node.Sibling()

	if !parent.IsBlack() && sibling.IsBlack() && sibling.left.IsBlack() && sibling.right.IsBlack() {
		sibling.isBlack = false
		parent.isBlack = true
		return
	}

	t.deleteCase5(node)
}

// deleteCase5 is an internal utility function that implements the 5th delete case.
func (t *Tree) deleteCase5(node *Node) {
	parent := node.parent
	sibling := node.Sibling()

	switch {
	case node == parent.left && sibling.IsBlack() && !sibling.left.IsBlack() && sibling.right.IsBlack():
		sibling.isBlack = false
		sibling.left.isBlack = true
		t.rotateRight(sibling)
	case node == parent.right && sibling.IsBlack() && !sibling.right.IsBlack() && sibling.left.IsBlack():
		sibling.isBlack = false
		sibling.right.isBlack = true
		t.rotateLeft(sibling)
	}

	t.deleteCase6(node)
}

// deleteCase6 is an internal utility function that implements the 6th delete case.
func (t *Tree) deleteCase6(node *Node) {
	parent := node.parent
	sibling := node.Sibling()

	sibling.isBlack = parent.IsBlack()
	parent.isBlack = true

	switch {
	case node == parent.left && !sibling.right.IsBlack():
		sibling.right.isBlack = true
		t.rotateLeft(parent)
	case !sibling.left.IsBlack():
		sibling.left.isBlack = true
		t.rotateRight(parent)
	}
}

// rotateLeft is an internal utility function that performs a left rotation to re-balance the tree.
func (t *Tree) rotateLeft(node *Node) {
	right := node.right
	t.swapNodes(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
	}
	right.left = node
	node.parent = right
}

// rotateRight is an internal utility function that performs a right rotation to re-balance the tree.
func (t *Tree) rotateRight(node *Node) {
	left := node.left
	t.swapNodes(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
	}
	left.right = node
	node.parent = left
}

// swapNodes is an internal utility function that swaps the position of a parent node with its child.
func (t *Tree) swapNodes(parent *Node, child *Node) {
	if child != nil {
		child.parent = parent.parent
	}
	if parent.parent == nil {
		t.root = child
		return
	}
	switch parent {
	case parent.parent.left:
		parent.parent.left = child
	case parent.parent.right:
		parent.parent.right = child
	}
}
