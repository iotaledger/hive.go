package discover

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/autopeering/peer/peertest"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/autopeering/server"
	"github.com/iotaledger/hive.go/autopeering/server/servertest"
	"github.com/iotaledger/hive.go/database/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNetwork = "test"
	testIP      = "127.0.0.1"
	testPort    = 8000
	graceTime   = 100 * time.Millisecond
)

var log = logger.NewExampleLogger("discover")

func init() {
	// decrease parameters to simplify and speed up tests
	SetParameters(Parameters{
		ReverifyInterval: 500 * time.Millisecond,
		QueryInterval:    1000 * time.Millisecond,
		MaxManaged:       10,
		MaxReplacements:  2,
	})
}

func TestProtVerifyMaster(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()

	peerA := getPeer(protA)

	// use peerA as masters peer
	protB, closeB := newTestProtocol("B", connB, log, peerA)

	time.Sleep(graceTime) // wait for the packages to ripple through the network
	closeB()              // close srvB to avoid race conditions, when asserting

	if assert.EqualValues(t, 1, len(protB.mgr.active)) {
		assert.EqualValues(t, peerA, &protB.mgr.active[0].Peer)
		assert.EqualValues(t, 1, protB.mgr.active[0].verifiedCount)
	}
}

func TestProtPingPong(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerA := getPeer(protA)
	peerB := getPeer(protB)

	// send connA Ping from node A to B
	t.Run("A->B", func(t *testing.T) { assert.NoError(t, protA.Ping(peerB)) })
	time.Sleep(graceTime)

	// send connA Ping from node B to A
	t.Run("B->A", func(t *testing.T) { assert.NoError(t, protB.Ping(peerA)) })
	time.Sleep(graceTime)
}

func TestProtPingTimeout(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	closeB() // close the connection right away to prevent any replies

	// send connA Ping from node A to B
	err := protA.Ping(getPeer(protB))
	assert.EqualError(t, err, server.ErrTimeout.Error())
}

func TestProtVerifiedPeers(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerB := getPeer(protB)

	// send connA Ping from node A to B
	assert.NoError(t, protA.Ping(peerB))
	time.Sleep(graceTime)

	// protA should have peerB as the single verified peer
	assert.ElementsMatch(t, []*peer.Peer{peerB}, protA.GetVerifiedPeers())
	for _, p := range protA.GetVerifiedPeers() {
		assert.Equal(t, p, protA.GetVerifiedPeer(p.ID()))
	}
}

func TestProtVerifiedPeer(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerA := getPeer(protA)
	peerB := getPeer(protB)

	// send connA Ping from node A to B
	assert.NoError(t, protA.Ping(peerB))
	time.Sleep(graceTime)

	// we should have peerB as connA verified peer
	assert.Equal(t, peerB, protA.GetVerifiedPeer(peerB.ID()))
	// we should not have ourselves as connA verified peer
	assert.Nil(t, protA.GetVerifiedPeer(peerA.ID()))
}

func TestProtDiscoveryRequest(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerA := getPeer(protA)
	peerB := getPeer(protB)

	// request peers from node A
	t.Run("A->B", func(t *testing.T) {
		if ps, err := protA.DiscoveryRequest(peerB); assert.NoError(t, err) {
			//assert.ElementsMatch(t, []*peer.Peer{peerA}, ps)
			if assert.Equal(t, 1, len(ps)) {
				assert.Equal(t, peerA.ID(), ps[0].ID())
			}
		}
	})
	// request peers from node B
	t.Run("B->A", func(t *testing.T) {
		if ps, err := protB.DiscoveryRequest(peerA); assert.NoError(t, err) {
			if assert.Equal(t, 1, len(ps)) {
				assert.Equal(t, peerB.ID(), ps[0].ID())
			}
		}
	})
}

func TestProtServices(t *testing.T) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()

	err := protA.local().UpdateService(service.FPCKey, "fpc", 123)
	require.NoError(t, err)

	peerA := getPeer(protA)

	// use peerA as masters peer
	protB, closeB := newTestProtocol("B", connB, log, peerA)
	defer closeB()

	time.Sleep(graceTime) // wait for the packages to ripple through the network
	ps := protB.GetVerifiedPeers()

	if assert.ElementsMatch(t, []*peer.Peer{peerA}, ps) {
		assert.Equal(t, protA.local().Services(), ps[0].Services())
	}
}

func TestProtDiscovery(t *testing.T) {
	connM := servertest.NewConn()
	defer connM.Close()
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()
	connC := servertest.NewConn()
	defer connC.Close()

	protM, closeM := newTestProtocol("M", connM, log)
	defer closeM()
	time.Sleep(graceTime) // wait for the master to initialize

	protA, closeA := newTestProtocol("A", connA, log, getPeer(protM))
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log, getPeer(protM))
	defer closeB()
	protC, closeC := newTestProtocol("C", connC, log, getPeer(protM))
	defer closeC()

	time.Sleep(queryInterval + graceTime)    // wait for the next discovery cycle
	time.Sleep(reverifyInterval + graceTime) // wait for the next verification cycle

	// now the full network should be discovered
	assert.ElementsMatch(t, []*peer.Peer{getPeer(protA), getPeer(protB), getPeer(protC)}, protM.GetVerifiedPeers())
	assert.ElementsMatch(t, []*peer.Peer{getPeer(protM), getPeer(protB), getPeer(protC)}, protA.GetVerifiedPeers())
	assert.ElementsMatch(t, []*peer.Peer{getPeer(protM), getPeer(protA), getPeer(protC)}, protB.GetVerifiedPeers())
	assert.ElementsMatch(t, []*peer.Peer{getPeer(protM), getPeer(protA), getPeer(protB)}, protC.GetVerifiedPeers())
}

func BenchmarkPingPong(b *testing.B) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	log := logger.NewNopLogger() // disable logging

	// disable query/reverify
	reverifyInterval = time.Hour
	queryInterval = time.Hour

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerB := getPeer(protB)

	// send initial Ping to ensure that every peer is verified
	err := protA.Ping(peerB)
	require.NoError(b, err)
	time.Sleep(graceTime)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// send connA Ping from node A to B
		_ = protA.Ping(peerB)
	}

	b.StopTimer()
}

func BenchmarkDiscoveryRequest(b *testing.B) {
	connA := servertest.NewConn()
	defer connA.Close()
	connB := servertest.NewConn()
	defer connB.Close()

	log := logger.NewNopLogger() // disable logging

	// disable query/reverify
	reverifyInterval = time.Hour
	queryInterval = time.Hour

	protA, closeA := newTestProtocol("A", connA, log)
	defer closeA()
	protB, closeB := newTestProtocol("B", connB, log)
	defer closeB()

	peerB := getPeer(protB)

	// send initial DiscoveryRequest to ensure that every peer is verified
	_, err := protA.DiscoveryRequest(peerB)
	require.NoError(b, err)
	time.Sleep(graceTime)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = protA.DiscoveryRequest(peerB)
	}

	b.StopTimer()
}

// newTestProtocol creates connA new discovery server and also returns the teardown.
func newTestProtocol(name string, conn *net.UDPConn, logger *logger.Logger, masters ...*peer.Peer) (*Protocol, func()) {
	db, _ := peer.NewDB(mapdb.NewMapDB())
	addr := conn.LocalAddr().(*net.UDPAddr)
	local := peertest.NewLocal(addr.Network(), addr.IP, addr.Port, db)
	log := logger.Named(name)

	prot := New(local, 0, 0, Logger(log), MasterPeers(masters))

	srv := server.Serve(local, conn, log, prot)
	prot.Start(srv)

	teardown := func() {
		srv.Close()
		prot.Close()
	}
	return prot, teardown
}

func getPeer(p *Protocol) *peer.Peer {
	return p.local().Peer
}
