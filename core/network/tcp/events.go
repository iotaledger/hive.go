package tcp

import (
	"github.com/iotaledger/hive.go/core/network"
	"github.com/iotaledger/hive.go/runtime/event"
)

type serverEvents struct {
	Start    *event.Event1[*StartEvent]
	Shutdown *event.Event1[*ShutdownEvent]
	Connect  *event.Event1[*ConnectEvent]
	Error    *event.Event1[error]
}

func newServerEvents() *serverEvents {
	return &serverEvents{
		Start:    event.New1[*StartEvent](),
		Shutdown: event.New1[*ShutdownEvent](),
		Connect:  event.New1[*ConnectEvent](),
		Error:    event.New1[error](),
	}
}

type StartEvent struct{}
type ShutdownEvent struct{}
type ConnectEvent struct {
	ManagedConnection *network.ManagedConnection
}
