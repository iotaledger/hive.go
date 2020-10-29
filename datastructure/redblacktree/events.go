package redblacktree

import (
	"github.com/iotaledger/hive.go/events"
)

type TreeEvents struct {
	NodeInserted *events.Event
}

type NodeInsertedEvent struct {
	InsertedNode *Node
	Floor        *Node
	Ceiling      *Node
}

func nodeInsertedEventHandler(handler interface{}, params ...interface{}) {
	handler.(func(*NodeInsertedEvent))(params[0].(*NodeInsertedEvent))
}
