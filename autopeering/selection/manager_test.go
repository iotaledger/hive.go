package selection

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/autopeering/peer/peertest"
	"github.com/izuc/zipp.foundation/autopeering/salt"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/kvstore/mapdb"
)

const (
	testSaltLifetime   = time.Hour     // disable salt updates
	testUpdateInterval = 2 * graceTime // very short update interval to speed up tests
)

func TestMgrNoDuplicates(t *testing.T) {
	const (
		nNeighbors = 4
		nNodes     = 2*nNeighbors + 1
	)
	SetParameters(Parameters{
		OutboundNeighborSize:   nNeighbors,
		InboundNeighborSize:    nNeighbors,
		SaltLifetime:           testSaltLifetime,
		OutboundUpdateInterval: testUpdateInterval,
	})

	mgrMap := make(map[identity.ID]*manager)
	runTestNetwork(nNodes, mgrMap, nil)

	for _, mgr := range mgrMap {
		assert.NotEmpty(t, mgr.getOutNeighbors())
		assert.NotEmpty(t, mgr.getInNeighbors())
		assert.Empty(t, getDuplicates(mgr.getNeighbors()))
	}
}
func TestBlocklistNeighbor(t *testing.T) {
	const (
		nNeighbors = 2
	)
	SetParameters(Parameters{
		OutboundNeighborSize:   nNeighbors,
		InboundNeighborSize:    nNeighbors,
		SaltLifetime:           testSaltLifetime,
		OutboundUpdateInterval: testUpdateInterval,
	})
	mgrsMap := make(map[identity.ID]*manager)
	blocklistingMgr := newTestManager("mgr 1", mgrsMap)
	blocklistedMgr := newTestManager("mgr 2", mgrsMap)
	blocklistingMgr.start()
	blocklistedMgr.start()
	time.Sleep(4 * graceTime)
	blocklistingMgr.blockNeighbor(blocklistedMgr.getID())
	t.Run("Blocklisting manager drops Blocklisted neighbor", func(t *testing.T) {
		assert.Eventually(t, func() bool {
			for _, p := range blocklistingMgr.getNeighbors() {
				if p == blocklistedMgr.net.local().Peer {
					return false
				}
			}

			return true
		}, time.Second, 10*time.Millisecond)
	})
	t.Run("Blocklisting manager doesn't select Blocklisted neighbor for outbound connection", func(t *testing.T) {
		got := blocklistingMgr.getOutboundPeeringCandidate()
		assert.Nil(t, got.Remote)
	})
	t.Run("Blocklisting manager rejects Blocklisted neighbor on incoming connection", func(t *testing.T) {
		assert.False(t, blocklistingMgr.requestPeering(blocklistedMgr.net.local().Peer, blocklistedMgr.net.local().GetPublicSalt()))
	})
	t.Run("Blocklisted manager stops trying to connect to Blocklisting neighbor after first failure", func(t *testing.T) {
		blocklistedMgr.cleanSkiplist()
		resultCh := make(chan peer.PeerDistance, 1)
		blocklistedMgr.updateOutbound(resultCh)
		result := <-resultCh
		assert.Nil(t, result.Remote)
		assert.True(t, blocklistingMgr.isInBlocklist(blocklistedMgr.getID()))
		assert.True(t, blocklistedMgr.isInSkiplist(blocklistingMgr.getID()))
		got := blocklistedMgr.getOutboundPeeringCandidate()
		assert.Nil(t, got.Remote)
	})
	t.Run("Blocklisted manager connects to Blocklisting neighbor after unblocking", func(t *testing.T) {
		blocklistingMgr.unblockNeighbor(blocklistedMgr.getID())
		blocklistedMgr.cleanSkiplist()
		resultCh := make(chan peer.PeerDistance, 1)
		blocklistedMgr.updateOutbound(resultCh)
		result := <-resultCh
		assert.NotNil(t, result.Remote)
		assert.False(t, blocklistingMgr.isInBlocklist(blocklistingMgr.getID()))
		assert.False(t, blocklistingMgr.isInSkiplist(blocklistingMgr.getID()))
		assert.False(t, blocklistedMgr.isInSkiplist(blocklistingMgr.getID()))
		got := blocklistedMgr.getOutboundPeeringCandidate()
		assert.NotNil(t, got.Remote)
	})
}

func TestEvents(t *testing.T) {
	// we want many drops/connects
	const (
		nNeighbors = 2
		nNodes     = 10
	)
	SetParameters(Parameters{
		OutboundNeighborSize:   nNeighbors,
		InboundNeighborSize:    nNeighbors,
		SaltLifetime:           3 * testUpdateInterval,
		OutboundUpdateInterval: testUpdateInterval,
	})

	e := newEventNetwork(t)
	mgrMap := make(map[identity.ID]*manager)
	runTestNetwork(nNodes, mgrMap, e.attach)

	// the events should lead to exactly the same neighbors
	for _, mgr := range mgrMap {
		nc := e.m[mgr.getID()]
		assert.ElementsMatchf(t, mgr.getOutNeighbors(), getValues(nc.out),
			"out neighbors of %s do not match", mgr.getID())
		assert.ElementsMatch(t, mgr.getInNeighbors(), getValues(nc.in),
			"in neighbors of %s do not match", mgr.getID())
	}
}

func getValues(m map[identity.ID]*peer.Peer) []*peer.Peer {
	result := make([]*peer.Peer, 0, len(m))
	for _, p := range m {
		result = append(result, p)
	}

	return result
}

func runTestNetwork(n int, mgrMap map[identity.ID]*manager, hook func(*manager)) {
	for i := 0; i < n; i++ {
		mgr := newTestManager(fmt.Sprintf("%d", i), mgrMap)
		if hook != nil {
			hook(mgr)
		}
	}
	for _, mgr := range mgrMap {
		mgr.start()
	}

	// give the managers time to potentially connect all other peers
	time.Sleep((time.Duration(n) - 1) * (outboundUpdateInterval + graceTime))

	// close all the managers
	for _, mgr := range mgrMap {
		mgr.close()
	}
}

func getDuplicates(peers []*peer.Peer) []*peer.Peer {
	seen := make(map[identity.ID]bool, len(peers))
	result := make([]*peer.Peer, 0, len(peers))

	for _, p := range peers {
		if !seen[p.ID()] {
			seen[p.ID()] = true
		} else {
			result = append(result, p)
		}
	}

	return result
}

type neighbors struct {
	out, in map[identity.ID]*peer.Peer
}

// eventNetwork reconstructs the neighbors for the triggered events.
type eventNetwork struct {
	sync.Mutex
	t *testing.T
	m map[identity.ID]neighbors
}

func newEventNetwork(t *testing.T) *eventNetwork {
	return &eventNetwork{
		t: t,
		m: make(map[identity.ID]neighbors),
	}
}

func (e *eventNetwork) attach(mgr *manager) {
	id := mgr.getID()
	mgr.Events.OutgoingPeering.Hook(func(ev *PeeringEvent) { e.outgoingPeering(id, ev) })
	mgr.Events.IncomingPeering.Hook(func(ev *PeeringEvent) { e.incomingPeering(id, ev) })
	mgr.Events.Dropped.Hook(func(ev *DroppedEvent) { e.dropped(id, ev) })
}

func (e *eventNetwork) outgoingPeering(id identity.ID, ev *PeeringEvent) {
	if !ev.Status {
		return
	}
	e.Lock()
	defer e.Unlock()
	s, ok := e.m[id]
	if !ok {
		s = neighbors{out: make(map[identity.ID]*peer.Peer), in: make(map[identity.ID]*peer.Peer)}
		e.m[id] = s
	}
	assert.NotContains(e.t, s.out, ev.Peer)
	assert.Less(e.t, len(s.out), 2)
	s.out[ev.Peer.ID()] = ev.Peer
}

func (e *eventNetwork) incomingPeering(id identity.ID, ev *PeeringEvent) {
	if !ev.Status {
		return
	}
	e.Lock()
	defer e.Unlock()
	s, ok := e.m[id]
	if !ok {
		s = neighbors{out: make(map[identity.ID]*peer.Peer), in: make(map[identity.ID]*peer.Peer)}
		e.m[id] = s
	}
	assert.NotContains(e.t, s.in, ev.Peer)
	assert.Less(e.t, len(s.in), 2)
	s.in[ev.Peer.ID()] = ev.Peer
}

func (e *eventNetwork) dropped(id identity.ID, ev *DroppedEvent) {
	e.Lock()
	defer e.Unlock()
	if assert.Contains(e.t, e.m, id) {
		s := e.m[id]
		delete(s.out, ev.DroppedID)
		delete(s.in, ev.DroppedID)
	}
}

type networkMock struct {
	loc *peer.Local
	mgr map[identity.ID]*manager
}

func (n *networkMock) local() *peer.Local {
	return n.loc
}

func (n *networkMock) PeeringDrop(p *peer.Peer) {
	n.mgr[p.ID()].removeNeighbor(n.local().ID())
}

func (n *networkMock) PeeringRequest(p *peer.Peer, s *salt.Salt) (bool, error) {
	return n.mgr[p.ID()].requestPeering(n.local().Peer, s), nil
}

func (n *networkMock) GetKnownPeers() []*peer.Peer {
	peers := make([]*peer.Peer, 0, len(n.mgr))
	for _, m := range n.mgr {
		peers = append(peers, m.net.local().Peer)
	}

	return peers
}

func newTestManager(name string, mgrMap map[identity.ID]*manager) *manager {
	db, _ := peer.NewDB(mapdb.NewMapDB())
	local := peertest.NewLocal("mock", net.IPv4zero, 0, db)
	nm := &networkMock{loc: local, mgr: mgrMap}
	m := newManager(nm, nm.GetKnownPeers, log.Named(name), &options{})
	mgrMap[m.getID()] = m

	return m
}
