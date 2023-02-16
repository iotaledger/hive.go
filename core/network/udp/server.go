package udp

import (
	"net"
	"strconv"

	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

//nolint:revive // better be explicit here
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
	socket, err := net.ListenPacket("udp", address+":"+strconv.Itoa(port))
	if err != nil {
		srv.Events.Error.Trigger(err)

		return
	}

	srv.socketMutex.Lock()
	srv.socket = socket
	srv.socketMutex.Unlock()

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
			Start:       event.New(),
			Shutdown:    event.New(),
			ReceiveData: event.New2[*net.UDPAddr, []byte](),
			Error:       event.New1[error](),
		},
	}
}
