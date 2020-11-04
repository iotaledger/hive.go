package thresholdmap

import (
	"github.com/iotaledger/hive.go/datastructure/redblacktree"
)

type iteratorState int

const (
	beginning iteratorState = iota
	between
	end
)

type Iterator struct {
	start   *Element
	current *Element
	state   iteratorState
}

func (i *Iterator) HasNext() bool {
	switch i.state {
	case beginning:
		return i.start != nil
	case between:
		return i.current.Successor() != nil
	}

	return false
}

func (i *Iterator) Next() *Element {
	switch i.state {
	case beginning:
		i.current = i.start
		i.state = between
	case between:
		i.current = i.wrapNode(i.current.Successor())
		if i.current.Successor() == nil {
			i.state = end
		}
	default:
		panic("no next element found in iterator")
	}

	return i.current
}

func (i *Iterator) Reset() {
	i.current = nil
	i.state = beginning
}

func (i *Iterator) wrapNode(node *redblacktree.Node) (element *Element) {
	if node == nil {
		return
	}

	return &Element{node}
}
