package selection

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/discover"
	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/autopeering/peer/peertest"
	"github.com/iotaledger/hive.go/autopeering/server"
	"github.com/iotaledger/hive.go/autopeering/server/servertest"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNetwork = "udp"
	testIP      = "127.0.0.1"
	graceTime   = 100 * time.Millisecond
)

var (
	log     = logger.NewExampleLogger("discover")
	peerMap = make(map[identity.ID]*peer.Peer)
)

func TestProtocol(t *testing.T) {
	// assure that the default test parameters are used for all protocol tests
	SetParameters(Parameters{
		ArRowLifetime:          testArrowLifetime,
		OutboundUpdateInterval: testUpdateInterval,
	})

	t.Run("PeeringRequest", func(t *testing.T) {
		connA := servertest.NewConn()
		defer connA.Close()
		connB := servertest.NewConn()
		defer connB.Close()

		protA, closeA := newTestProtocol("A", connA)
		defer closeA()
		protB, closeB := newTestProtocol("B", connB)
		defer closeB()

		peerA := getPeer(protA)
		peerB := getPeer(protB)

		// request peering to peer B
		t.Run("A->B", func(t *testing.T) {
			if accepted, err := protA.PeeringRequest(peerB, 0); assert.NoError(t, err) {
				assert.True(t, accepted)
			}
		})
		// request peering to peer A
		t.Run("B->A", func(t *testing.T) {
			if accepted, err := protB.PeeringRequest(peerA, 0); assert.NoError(t, err) {
				assert.True(t, accepted)
			}
		})
	})

	t.Run("PeeringDrop", func(t *testing.T) {
		connA := servertest.NewConn()
		defer connA.Close()
		connB := servertest.NewConn()
		defer connB.Close()

		protA, closeA := newTestProtocol("A", connA)
		defer closeA()
		protB, closeB := newTestProtocol("B", connB)
		defer closeB()

		peerA := getPeer(protA)
		peerB := getPeer(protB)

		// request peering to peer B
		status, err := protA.PeeringRequest(peerB, 0)
		require.NoError(t, err)
		assert.True(t, status)

		require.Contains(t, protB.GetNeighbors(), peerA)

		// drop peer A
		protA.PeeringDrop(peerB)
		time.Sleep(graceTime)
		require.NotContains(t, protB.GetNeighbors(), peerA)
	})

	t.Run("FullTest", func(t *testing.T) {
		connA := servertest.NewConn()
		defer connA.Close()
		connB := servertest.NewConn()
		defer connB.Close()

		protA, closeA := newFullTestProtocol("A", connA)
		defer closeA()

		time.Sleep(graceTime) // wait for the master to initialize

		protB, closeB := newFullTestProtocol("B", connB, getPeer(protA))
		defer closeB()

		time.Sleep(outboundUpdateInterval + graceTime) // wait for the next outbound cycle

		// the two peers should be peered
		assert.ElementsMatch(t, []*peer.Peer{getPeer(protB)}, protA.GetNeighbors())
		assert.ElementsMatch(t, []*peer.Peer{getPeer(protA)}, protB.GetNeighbors())
	})
}

// dummyDiscovery is a dummy implementation of DiscoveryProtocol never returning any verified peers.
type dummyDiscovery struct{}

func (dummyDiscovery) IsVerified(_ identity.ID, _ net.IP) bool   { return true }
func (dummyDiscovery) EnsureVerified(_ *peer.Peer) error         { return nil }
func (dummyDiscovery) GetVerifiedPeer(id identity.ID) *peer.Peer { return peerMap[id] }
func (dummyDiscovery) GetVerifiedPeers() []*peer.Peer            { return []*peer.Peer{} }

// newTestProtocol creates a new neighborhood server and also returns the teardown.
func newTestProtocol(name string, conn *net.UDPConn) (*Protocol, func()) {
	addr := conn.LocalAddr().(*net.UDPAddr)
	db, _ := peer.NewDB(mapdb.NewMapDB())
	local := peertest.NewLocal(addr.Network(), addr.IP, addr.Port, db)
	// add the new peer to the global map for dummyDiscovery
	peerMap[local.ID()] = local.Peer
	l := log.Named(name)

	prot := New(local, dummyDiscovery{}, Logger(l.Named("sel")))
	srv := server.Serve(local, conn, l.Named("srv"), prot)
	prot.Start(srv)

	teardown := func() {
		srv.Close()
		prot.Close()
	}
	return prot, teardown
}

// newTestProtocol creates a new server handling discover as well as neighborhood and also returns the teardown.
func newFullTestProtocol(name string, conn *net.UDPConn, masterPeers ...*peer.Peer) (*Protocol, func()) {
	addr := conn.LocalAddr().(*net.UDPAddr)
	db, _ := peer.NewDB(mapdb.NewMapDB())
	local := peertest.NewLocal(addr.Network(), addr.IP, addr.Port, db)
	// add the new peer to the global map for dummyDiscovery
	peerMap[local.ID()] = local.Peer
	l := log.Named(name)

	discovery := discover.New(local, 0, 0,
		discover.Logger(l.Named("disc")),
		discover.MasterPeers(masterPeers),
	)
	selection := New(local, discovery, Logger(l.Named("sel")))

	srv := server.Serve(local, conn, l.Named("srv"), discovery, selection)

	discovery.Start(srv)
	selection.Start(srv)

	teardown := func() {
		srv.Close()
		selection.Close()
		discovery.Close()
	}
	return selection, teardown
}

func getPeer(p *Protocol) *peer.Peer {
	return p.local().Peer
}
