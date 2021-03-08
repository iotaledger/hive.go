package redblacktree

import (
	"github.com/iotaledger/hive.go/datastructure/genericcomparator"
	"github.com/iotaledger/hive.go/stringify"
)

// region RedBlackTree /////////////////////////////////////////////////////////////////////////////////////////////////

// RedBlackTree represents a self balancing binary search tree, that can be used to efficiently look up value associated
// to a set of keys.
type RedBlackTree struct {
	root       *Node
	min        *Node
	max        *Node
	comparator genericcomparator.Type
	size       int
}

// New creates a new red-black RedBlackTree that uses the given comparator (or the default Comparator if the parameter is
// omitted) to compare the keys used to identify the nodes.
func New(optionalComparator ...genericcomparator.Type) *RedBlackTree {
	if len(optionalComparator) >= 1 {
		return &RedBlackTree{
			comparator: optionalComparator[0],
		}
	}

	return &RedBlackTree{
		comparator: genericcomparator.Comparator,
	}
}

// Set inserts or updates a Node in the RedBlackTree and returns it together with a flag that indicates if it was
// inserted.
func (t *RedBlackTree) Set(key interface{}, value interface{}) (node *Node, inserted bool) {
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
func (t *RedBlackTree) Get(key interface{}) (value interface{}, found bool) {
	if node := t.Node(key); node != nil {
		value = node.value
		found = true
		return
	}

	return
}

// Delete removes a Node belonging to the given key from the RedBlackTree and returns it (if it existed) together with a flag
// that indicates if it existed.
func (t *RedBlackTree) Delete(key interface{}) (node *Node, success bool) {
	node = t.Node(key)
	if success = node != nil; !success {
		return
	}

	t.DeleteNode(node)

	return
}

// DeleteNode removes the Node from the RedBlackTree (which can be i.e. useful for modifying the RedBlackTree while
// iterating.
func (t *RedBlackTree) DeleteNode(node *Node) {
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

// ForEach iterates through the Nodes of the RedBlackTree in ascending order and calls the iterator function for each Node. The
// iteration aborts as soon as the iterator function returns false.
func (t *RedBlackTree) ForEach(iterator func(node *Node) bool) {
	abortIteration := false
	for currentNode := t.Min(); currentNode != nil && !abortIteration; currentNode = currentNode.successor {
		abortIteration = !iterator(currentNode)
	}
}

// Keys returns an ordered list of keys that are stored in the RedBlackTree.
func (t *RedBlackTree) Keys() (keys []interface{}) {
	keys = make([]interface{}, 0, t.size)
	for currentNode := t.Min(); currentNode != nil; currentNode = currentNode.successor {
		keys = append(keys, currentNode.key)
	}

	return
}

// Values returns an ordered list of values that are stored in the RedBlackTree.
func (t *RedBlackTree) Values() (values []interface{}) {
	values = make([]interface{}, 0, t.size)
	for currentNode := t.Min(); currentNode != nil; currentNode = currentNode.successor {
		values = append(values, currentNode.value)
	}

	return
}

// Node returns the Node that belongs to the given key (or nil if it doesn't exist).
func (t *RedBlackTree) Node(key interface{}) (node *Node) {
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

// Min returns the Node with the smallest key (or nil if the RedBlackTree is empty).
func (t *RedBlackTree) Min() *Node {
	return t.root.Min()
}

// Max returns the Node with the largest key (or nil if the RedBlackTree is empty).
func (t *RedBlackTree) Max() *Node {
	return t.root.Max()
}

// Floor returns the Node with the largest key that is <= the given key (or nil if no floor was found).
func (t *RedBlackTree) Floor(key interface{}) (floor *Node) {
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
func (t *RedBlackTree) Ceiling(key interface{}) (ceiling *Node) {
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

// Size returns the amount of Nodes in the RedBlackTree.
func (t *RedBlackTree) Size() int {
	return t.size
}

// Empty returns true if the RedBlackTree has no Nodes.
func (t *RedBlackTree) Empty() bool {
	return t.size == 0
}

// Clear removes all Nodes from the RedBlackTree.
func (t *RedBlackTree) Clear() {
	t.root = nil
	t.min = nil
	t.max = nil
	t.size = 0
}

// String returns a human readable version of the RedBlackTree.
func (t *RedBlackTree) String() string {
	return stringify.Struct("RedBlackTree",
		stringify.StructField("size", t.size),
		stringify.StructField("root", t.root),
	)
}

// insertCase1 is an internal utility function that implements the 1st insert case.
func (t *RedBlackTree) insertCase1(node *Node) {
	if node.parent == nil {
		node.isBlack = true
		return
	}

	t.insertCase2(node)
}

// insertCase2 is an internal utility function that implements the 2nd insert case.
func (t *RedBlackTree) insertCase2(node *Node) {
	if node.parent.IsBlack() {
		return
	}

	t.insertCase3(node)
}

// insertCase3 is an internal utility function that implements the 3rd insert case.
func (t *RedBlackTree) insertCase3(node *Node) {
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
func (t *RedBlackTree) insertCase4(node *Node) {
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
func (t *RedBlackTree) insertCase5(node *Node) {
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
func (t *RedBlackTree) deleteCase1(node *Node) {
	if node.parent == nil {
		return
	}

	t.deleteCase2(node)
}

// deleteCase2 is an internal utility function that implements the 2nd delete case.
func (t *RedBlackTree) deleteCase2(node *Node) {
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
func (t *RedBlackTree) deleteCase3(node *Node) {
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
func (t *RedBlackTree) deleteCase4(node *Node) {
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
func (t *RedBlackTree) deleteCase5(node *Node) {
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
func (t *RedBlackTree) deleteCase6(node *Node) {
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
func (t *RedBlackTree) rotateLeft(node *Node) {
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
func (t *RedBlackTree) rotateRight(node *Node) {
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
func (t *RedBlackTree) swapNodes(parent *Node, child *Node) {
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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Node /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Node represents a Node in the RedBlackTree.
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

// Parent returns the parent of the Node (or nil if the Node is the root of the RedBlackTree).
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

// IsBlack returns true if the Node is marked as black (colors are used for the self-balancing properties of the
// RedBlackTree).
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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
