package network

import "github.com/iotaledger/hive.go/generics/event"

type ManagedConnectionEvents struct {
	ReceiveData *event.Event[*ReceivedDataEvent]
	Close       *event.Event[*CloseEvent]
	Error       *event.Event[error]
}

func newManagedConnectionEvents() (new *ManagedConnectionEvents) {
	return &ManagedConnectionEvents{
		ReceiveData: event.New[*ReceivedDataEvent](),
		Close:       event.New[*CloseEvent](),
		Error:       event.New[error](),
	}
}

type ReceivedDataEvent struct {
	Data []byte
}

type CloseEvent struct{}
