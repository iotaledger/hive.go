package buffconn

import (
	"github.com/iotaledger/hive.go/runtime/event"
)

// BufferedConnectionEvents contains all the events that are triggered during the peer discovery.
type BufferedConnectionEvents struct {
	ReceiveMessage *event.Event1[*ReceiveMessageEvent]
	Close          *event.Event1[*CloseEvent]
}

func newBufferedConnectionEvents() *BufferedConnectionEvents {
	return &BufferedConnectionEvents{
		ReceiveMessage: event.New1[*ReceiveMessageEvent](),
		Close:          event.New1[*CloseEvent](),
	}
}

type ReceiveMessageEvent struct {
	Data []byte
}

type CloseEvent struct{}
