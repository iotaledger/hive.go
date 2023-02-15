package udp

import (
	"net"

	"github.com/iotaledger/hive.go/runtime/event"
)

type udpServerEvents struct {
	Start       *event.Event
	Shutdown    *event.Event
	ReceiveData *event.Event2[*net.UDPAddr, []byte]
	Error       *event.Event1[error]
}

func udpAddrAndDataCaller(handler interface{}, params ...interface{}) {
	handler.(func(*net.UDPAddr, []byte))(params[0].(*net.UDPAddr), params[1].([]byte))
}
