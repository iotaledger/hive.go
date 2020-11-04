package thresholdmap

import (
	"github.com/iotaledger/hive.go/datastructure/genericcomparator"
	"github.com/iotaledger/hive.go/datastructure/redblacktree"
)

type ThresholdMap struct {
	tree *redblacktree.Tree
}

func New(optionalComparator ...genericcomparator.Type) *ThresholdMap {
	if len(optionalComparator) >= 1 {
		return &ThresholdMap{
			tree: redblacktree.New(optionalComparator[0]),
		}
	}

	return &ThresholdMap{
		tree: redblacktree.New(),
	}
}

func (t *ThresholdMap) Set(key interface{}, value interface{}) {
	t.tree.Set(key, value)
}

func (t *ThresholdMap) Get(key interface{}) (value interface{}, exists bool) {
	floor := t.tree.Floor(key)

	if exists = floor != nil; !exists {
		return
	}

	value = floor.Value()
	return
}

func (t *ThresholdMap) Delete(key interface{}) (element *Element, success bool) {
	node, success := t.tree.Delete(key)
	element = t.wrapNode(node)

	return
}

func (t *ThresholdMap) Keys() []interface{} {
	return t.tree.Keys()
}

func (t *ThresholdMap) Values() []interface{} {
	return t.tree.Values()
}

func (t *ThresholdMap) GetElement(key interface{}) *Element {
	return t.wrapNode(t.tree.Floor(key))
}

func (t *ThresholdMap) MinElement() *Element {
	return t.wrapNode(t.tree.Min())
}

func (t *ThresholdMap) MaxElement() *Element {
	return t.wrapNode(t.tree.Max())
}

func (t *ThresholdMap) DeleteElement(element *Element) {
	if element == nil {
		return
	}

	t.tree.DeleteNode(element.Node)
}

func (t *ThresholdMap) ForEach(iterator func(node *Element) bool) {
	t.tree.ForEach(func(node *redblacktree.Node) bool {
		return iterator(t.wrapNode(node))
	})
}

func (t *ThresholdMap) Iterator(optionalStartingNode ...*Element) *Iterator {
	if len(optionalStartingNode) >= 1 {
		return &Iterator{
			start: optionalStartingNode[0],
		}
	}

	return &Iterator{
		start: t.wrapNode(t.tree.Min()),
	}
}

func (t *ThresholdMap) Size() int {
	return t.tree.Size()
}

func (t *ThresholdMap) Empty() bool {
	return t.tree.Empty()
}

func (t *ThresholdMap) Clear() {
	t.tree.Clear()
}

func (t *ThresholdMap) wrapNode(node *redblacktree.Node) (element *Element) {
	if node == nil {
		return
	}

	return &Element{node}
}
