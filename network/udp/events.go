package udp

import (
	"github.com/iotaledger/hive.go/v2/events"
	"net"
)

type udpServerEvents struct {
	Start       *events.Event
	Shutdown    *events.Event
	ReceiveData *events.Event
	Error       *events.Event
}

func udpAddrAndDataCaller(handler interface{}, params ...interface{}) {
	handler.(func(*net.UDPAddr, []byte))(params[0].(*net.UDPAddr), params[1].([]byte))
}
