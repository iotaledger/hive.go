package redblacktree

import (
	"github.com/iotaledger/hive.go/stringify"
)

type Node struct {
	Key    interface{}
	Value  interface{}
	Left   *Node
	Right  *Node
	Parent *Node
	color  Color
}

func (n *Node) Color() Color {
	if n == nil {
		return ColorBlack
	}

	return n.color
}

func (n *Node) GrandParent() *Node {
	if n != nil && n.Parent != nil {
		return n.Parent.Parent
	}

	return nil
}

func (n *Node) Uncle() *Node {
	if n == nil || n.Parent == nil || n.Parent.Parent == nil {
		return nil
	}

	return n.Parent.Sibling()
}

func (n *Node) Sibling() *Node {
	if n == nil || n.Parent == nil {
		return nil
	}

	if n == n.Parent.Left {
		return n.Parent.Right
	}

	return n.Parent.Left
}

func (n *Node) String() string {
	return stringify.Struct("Node",
		stringify.StructField("key", n.Key),
		stringify.StructField("left", n.Left),
		stringify.StructField("right", n.Right),
	)
}
