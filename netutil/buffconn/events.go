package buffconn

import "github.com/iotaledger/hive.go/generics/event"

// BufferedConnectionEvents contains all the events that are triggered during the peer discovery.
type BufferedConnectionEvents struct {
	ReceiveMessage *event.Event[*ReceiveMessageEvent]
	Close          *event.Event[*CloseEvent]
}

func newBufferedConnectionEvents() (new *BufferedConnectionEvents) {
	return &BufferedConnectionEvents{
		ReceiveMessage: event.New[*ReceiveMessageEvent](),
		Close:          event.New[*CloseEvent](),
	}
}

type ReceiveMessageEvent struct {
	Data []byte
}

type CloseEvent struct{}
