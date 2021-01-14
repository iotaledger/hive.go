package server

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/autopeering/server/servertest"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

const graceTime = 5 * time.Millisecond

var log = logger.NewExampleLogger("server")

const (
	MPing MType = iota
	MPong
)

type testMessage interface {
	Message
	Marshal() []byte
}

type Ping struct{ emptypb.Empty }
type Pong struct{ emptypb.Empty }

func (*Ping) Type() MType     { return MPing }
func (*Ping) Marshal() []byte { return append([]byte{}, byte(MPing)) }

func (*Pong) Type() MType     { return MPong }
func (*Pong) Marshal() []byte { return append([]byte{}, byte(MPong)) }

func sendPong(args mock.Arguments) {
	srv := args.Get(0).(*Server)
	addr := args.Get(1).(*net.UDPAddr)
	srv.Send(addr, new(Pong).Marshal())
}

var (
	pingMock *mock.Mock
	pongMock *mock.Mock
)

func setupTest() func(t *testing.T) {
	pingMock = new(mock.Mock)
	pongMock = new(mock.Mock)

	return func(t *testing.T) {
		pingMock.AssertExpectations(t)
		pingMock = nil
		pongMock.AssertExpectations(t)
		pongMock = nil
	}
}

func handle(s *Server, fromAddr *net.UDPAddr, from *identity.Identity, data []byte) (bool, error) {
	msg, err := unmarshal(data)
	if err != nil {
		return false, err
	}

	switch msg.Type() {
	case MPing:
		pingMock.Called(s, fromAddr, from, data)

	case MPong:
		if s.IsExpectedReply(fromAddr.IP, from.ID(), msg) {
			pongMock.Called(s, fromAddr, from, data)
		}

	default:
		panic("unknown message type")
	}

	return true, nil
}

func unmarshal(data []byte) (testMessage, error) {
	if len(data) != 1 {
		return nil, ErrInvalidMessage
	}

	switch MType(data[0]) {
	case MPing:
		return new(Ping), nil
	case MPong:
		return new(Pong), nil
	}
	return nil, ErrInvalidMessage
}

func newTestDB(t require.TestingT) *peer.DB {
	db, err := peer.NewDB(mapdb.NewMapDB())
	require.NoError(t, err)
	return db
}

func TestSrvEncodeDecodePing(t *testing.T) {
	services := service.New()
	services.Update(service.PeeringKey, "dummy", 8000)
	local, err := peer.NewLocal(net.IPv4zero, services, newTestDB(t))
	require.NoError(t, err)
	s := &Server{local: local}

	ping := new(Ping)
	packet := s.encode(ping.Marshal())

	data, id, err := decode(packet)
	require.NoError(t, err)

	msg, _ := unmarshal(data)
	assert.Equal(t, local.LocalIdentity().Identity, id)
	assert.Equal(t, msg, ping)
}

func newTestServer(t require.TestingT, name string, conn *net.UDPConn) (*Server, func()) {
	l := log.Named(name)

	services := service.New()
	services.Update(service.PeeringKey, conn.LocalAddr().Network(), conn.LocalAddr().(*net.UDPAddr).Port)
	local, err := peer.NewLocal(conn.LocalAddr().(*net.UDPAddr).IP, services, newTestDB(t))
	require.NoError(t, err)

	srv := Serve(local, conn, l, HandlerFunc(handle))

	return srv, srv.Close
}

func sendPing(s *Server, p *peer.Peer) error {
	ping := new(Ping)
	isPong := func(msg Message) bool {
		_, ok := msg.(*Pong)
		return ok
	}

	errc := s.SendExpectingReply(p.Address(), p.ID(), ping.Marshal(), MPong, isPong)
	return <-errc
}

func TestPingPong(t *testing.T) {
	a := servertest.NewConn()
	defer a.Close()
	b := servertest.NewConn()
	defer b.Close()

	srvA, closeA := newTestServer(t, "A", a)
	defer closeA()
	srvB, closeB := newTestServer(t, "B", b)
	defer closeB()

	peerA := srvA.Local().Peer
	peerB := srvB.Local().Peer

	t.Run("A->B", func(t *testing.T) {
		defer setupTest()(t)

		// B expects valid ping from A and sends pong back
		pingMock.On("handle", srvB, peerA.Address(), peerA.Identity, mock.Anything).Run(sendPong).Once()
		// A expects valid pong from B
		pongMock.On("handle", srvA, peerB.Address(), peerB.Identity, mock.Anything).Once()

		assert.NoError(t, sendPing(srvA, peerB))
		time.Sleep(graceTime)

	})

	t.Run("B->A", func(t *testing.T) {
		defer setupTest()(t)

		pingMock.On("handle", srvA, peerB.Address(), peerB.Identity, mock.Anything).Run(sendPong).Once() // A expects valid ping from B and sends pong back
		pongMock.On("handle", srvB, peerA.Address(), peerA.Identity, mock.Anything).Once()               // B expects valid pong from A

		assert.NoError(t, sendPing(srvB, peerA))
		time.Sleep(graceTime)
	})
}

func TestSrvPingTimeout(t *testing.T) {
	defer setupTest()(t)

	a := servertest.NewConn()
	defer a.Close()
	b := servertest.NewConn()
	defer b.Close()

	srvA, closeA := newTestServer(t, "A", a)
	defer closeA()
	srvB, closeB := newTestServer(t, "B", b)
	closeB()

	peerB := srvB.Local().Peer
	assert.EqualError(t, sendPing(srvA, peerB), ErrTimeout.Error())
}

func TestUnexpectedPong(t *testing.T) {
	defer setupTest()(t)

	a := servertest.NewConn()
	defer a.Close()
	b := servertest.NewConn()
	defer b.Close()

	srvA, closeA := newTestServer(t, "A", a)
	defer closeA()
	srvB, closeB := newTestServer(t, "B", b)
	defer closeB()

	// there should never be a Ping.Handle
	// there should never be a Pong.Handle

	srvA.Send(srvB.LocalAddr(), new(Pong).Marshal())
}
