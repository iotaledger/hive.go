package udp

import (
	"github.com/iotaledger/hive.go/v2/events"
	"github.com/iotaledger/hive.go/v2/syncutils"
	"net"
	"strconv"
)

type UDPServer struct {
	socket            net.PacketConn
	socketMutex       syncutils.RWMutex
	ReceiveBufferSize int
	Events            udpServerEvents
}

func (srv *UDPServer) GetSocket() net.PacketConn {
	srv.socketMutex.RLock()
	defer srv.socketMutex.RUnlock()
	return srv.socket
}

func (srv *UDPServer) Shutdown() {
	srv.socketMutex.Lock()
	defer srv.socketMutex.Unlock()
	if srv.socket != nil {
		socket := srv.socket
		srv.socket = nil

		socket.Close()
	}
}

func (srv *UDPServer) Listen(address string, port int) {
	if socket, err := net.ListenPacket("udp", address+":"+strconv.Itoa(port)); err != nil {
		srv.Events.Error.Trigger(err)

		return
	} else {
		srv.socketMutex.Lock()
		srv.socket = socket
		srv.socketMutex.Unlock()
	}

	srv.Events.Start.Trigger()
	defer srv.Events.Shutdown.Trigger()

	buf := make([]byte, srv.ReceiveBufferSize)
	for srv.GetSocket() != nil {
		if bytesRead, addr, err := srv.GetSocket().ReadFrom(buf); err != nil {
			if srv.GetSocket() != nil {
				srv.Events.Error.Trigger(err)
			}
		} else {
			srv.Events.ReceiveData.Trigger(addr.(*net.UDPAddr), buf[:bytesRead])
		}
	}
}

func NewServer(receiveBufferSize int) *UDPServer {
	return &UDPServer{
		ReceiveBufferSize: receiveBufferSize,
		Events: udpServerEvents{
			Start:       events.NewEvent(events.VoidCaller),
			Shutdown:    events.NewEvent(events.VoidCaller),
			ReceiveData: events.NewEvent(udpAddrAndDataCaller),
			Error:       events.NewEvent(events.ErrorCaller),
		},
	}
}
