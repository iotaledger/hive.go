package redblacktree

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/stringify"
)

type Comparator func(a interface{}, b interface{}) int

type Tree struct {
	Events     *TreeEvents
	Root       *Node
	Comparator Comparator
	size       int
}

func New(comparator Comparator) *Tree {
	return &Tree{
		Comparator: comparator,
		Events: &TreeEvents{
			NodeInserted: events.NewEvent(nodeInsertedEventHandler),
		},
	}
}

func (t *Tree) Get(key interface{}) (value interface{}, found bool) {
	node := t.GetNode(key)
	if found = node != nil; found {
		value = node.Value

		return
	}

	return
}

func (t *Tree) Put(key interface{}, value interface{}) {
	if t.Root == nil {
		t.Root = &Node{Key: key, Value: value, color: ColorBlack}
		t.size++

		return
	}

	insertedNode := (*Node)(nil)
	floor := (*Node)(nil)
	ceiling := (*Node)(nil)
	currentNode := t.Root

Iteration:
	for {
		switch t.Comparator(key, currentNode.Key) {
		case 0:
			currentNode.Key = key
			currentNode.Value = value

			return
		case -1:
			if currentNode.Left == nil {
				ceiling = currentNode
				insertedNode = &Node{Key: key, Value: value, color: ColorRed}
				currentNode.Left = insertedNode

				break Iteration
			}

			ceiling = currentNode
			currentNode = currentNode.Left
		case 1:
			if currentNode.Right == nil {
				floor = currentNode
				insertedNode = &Node{Key: key, Value: value, color: ColorRed}
				currentNode.Right = insertedNode

				break Iteration
			}

			floor = currentNode
			currentNode = currentNode.Right
		}
	}
	insertedNode.Parent = currentNode
	t.size++

	t.insertCase1(insertedNode)

	t.Events.NodeInserted.Trigger(&NodeInsertedEvent{
		InsertedNode: insertedNode,
		Floor:        floor,
		Ceiling:      ceiling,
	})
}

func (t *Tree) GetNode(key interface{}) (node *Node) {
	for node = t.Root; node != nil; {
		switch t.Comparator(key, node.Key) {
		case 0:
			return
		case -1:
			node = node.Left
		case 1:
			node = node.Right
		}
	}

	return
}

func (t *Tree) String() string {
	return stringify.Struct("Tree",
		stringify.StructField("size", t.size),
		stringify.StructField("root", t.Root),
	)
}

func (t *Tree) insertCase1(node *Node) {
	if node.Parent == nil {
		node.color = ColorBlack

		return
	}

	t.insertCase2(node)
}

func (t *Tree) insertCase2(node *Node) {
	if node.Parent.Color() == ColorBlack {
		return
	}

	t.insertCase3(node)
}

func (t *Tree) insertCase3(node *Node) {
	uncle := node.Uncle()
	if uncle.Color() == ColorRed {
		node.Parent.color = ColorBlack
		uncle.color = ColorBlack

		grandParent := node.GrandParent()
		grandParent.color = ColorRed
		t.insertCase1(grandParent)

		return
	}

	t.insertCase4(node)
}

func (t *Tree) insertCase4(node *Node) {
	grandParent := node.GrandParent()
	if node == node.Parent.Right && node.Parent == grandParent.Left {
		t.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandParent.Right {
		t.rotateRight(node.Parent)
		node = node.Right
	}

	t.insertCase5(node)
}

func (t *Tree) insertCase5(node *Node) {
	node.Parent.color = ColorBlack
	grandparent := node.GrandParent()
	grandparent.color = ColorRed

	if node == node.Parent.Left && node.Parent == grandparent.Left {
		t.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		t.rotateLeft(grandparent)
	}
}

func (t *Tree) rotateLeft(node *Node) {
	right := node.Right
	t.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (t *Tree) rotateRight(node *Node) {
	left := node.Left
	t.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (t *Tree) replaceNode(old *Node, new *Node) {
	if old.Parent == nil {
		t.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}
