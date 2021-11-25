package network

import (
	"github.com/iotaledger/hive.go/v2/events"
)

type ManagedConnectionEvents struct {
	ReceiveData *events.Event
	Close       *events.Event
	Error       *events.Event
}

func ManagedConnectionCaller(handler interface{}, params ...interface{}) {
	handler.(func(*ManagedConnection))(params[0].(*ManagedConnection))
}
