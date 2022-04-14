package tcp

import (
	"github.com/iotaledger/hive.go/generics/event"
	"github.com/iotaledger/hive.go/network"
)

type tcpServerEvents struct {
	Start    *event.Event[*StartEvent]
	Shutdown *event.Event[*ShutdownEvent]
	Connect  *event.Event[*ConnectEvent]
	Error    *event.Event[error]
}

func newTcpServerEvents() (new *tcpServerEvents) {
	return &tcpServerEvents{
		Start:    event.New[*StartEvent](),
		Shutdown: event.New[*ShutdownEvent](),
		Connect:  event.New[*ConnectEvent](),
		Error:    event.New[error](),
	}
}

type StartEvent struct{}
type ShutdownEvent struct{}
type ConnectEvent struct {
	ManagedConnection *network.ManagedConnection
}
