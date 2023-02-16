package network

import (
	"github.com/iotaledger/hive.go/runtime/event"
)

type ManagedConnectionEvents struct {
	ReceiveData *event.Event1[*ReceivedDataEvent]
	Close       *event.Event1[*CloseEvent]
	Error       *event.Event1[error]
}

func newManagedConnectionEvents() *ManagedConnectionEvents {
	return &ManagedConnectionEvents{
		ReceiveData: event.New1[*ReceivedDataEvent](),
		Close:       event.New1[*CloseEvent](),
		Error:       event.New1[error](),
	}
}

type ReceivedDataEvent struct {
	Data []byte
}

type CloseEvent struct{}
