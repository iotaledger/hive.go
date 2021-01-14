package selection

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/autopeering/peer/peertest"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
)

const (
	testArrowLifetime  = 3 * time.Hour // disable arrow updates
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
		ArRowLifetime:          testArrowLifetime,
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

func TestEvents(t *testing.T) {
	// we want many drops/connects
	const (
		nNeighbors = 2
		nNodes     = 10
	)
	SetParameters(Parameters{
		OutboundNeighborSize:   nNeighbors,
		InboundNeighborSize:    nNeighbors,
		ArRowLifetime:          3 * testArrowLifetime,
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

// eventNetwork reconstructs the neighbors for the triggered events
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
	mgr.events.OutgoingPeering.Attach(events.NewClosure(func(ev *PeeringEvent) { e.outgoingPeering(id, ev) }))
	mgr.events.IncomingPeering.Attach(events.NewClosure(func(ev *PeeringEvent) { e.incomingPeering(id, ev) }))
	mgr.events.Dropped.Attach(events.NewClosure(func(ev *DroppedEvent) { e.dropped(id, ev) }))
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

func (n *networkMock) PeeringRequest(p *peer.Peer, s int) (bool, error) {
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
	networkMock := &networkMock{loc: local, mgr: mgrMap}
	m := newManager(networkMock, networkMock.GetKnownPeers, log.Named(name), &options{})
	mgrMap[m.getID()] = m
	return m
}
