package discover

import (
	"math/rand"
	"sync"
	"time"

	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/autopeering/server"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/timeutil"
)

const (
	// PingExpiration is the time until a peer verification expires.
	PingExpiration = 12 * time.Hour
	// MaxPeersInResponse is the maximum number of peers returned in DiscoveryResponse.
	MaxPeersInResponse = 6
	// MaxServices is the maximum number of services a peer can support.
	MaxServices = 5
)

type network interface {
	local() *peer.Local

	Ping(*peer.Peer) error
	DiscoveryRequest(*peer.Peer) ([]*peer.Peer, error)
}

type manager struct {
	masters []*mpeer

	mutex        sync.Mutex // protects active and replacement
	active       []*mpeer
	replacements []*mpeer

	events *Events
	net    network
	log    *logger.Logger

	wg      sync.WaitGroup
	closing chan struct{}
}

func newManager(net network, masters []*peer.Peer, log *logger.Logger) *manager {
	return &manager{
		masters:      wrapPeers(masters),
		active:       make([]*mpeer, 0, maxManaged),
		replacements: make([]*mpeer, 0, maxReplacements),
		events:       newEvents(),
		net:          net,
		log:          log,
		closing:      make(chan struct{}),
	}
}

func (m *manager) start() {
	m.loadInitialPeers()

	m.wg.Add(1)
	go m.loop()
}

func (m *manager) self() identity.ID {
	return m.net.local().ID()
}

func (m *manager) close() {
	close(m.closing)
	m.wg.Wait()
}

func (m *manager) loop() {
	defer m.wg.Done()

	var (
		reverify     = time.NewTimer(0) // setting this to 0 will cause a trigger right away
		reverifyDone chan struct{}

		query     = time.NewTimer(server.ResponseTimeout) // trigger the first query after the reverify
		queryNext chan time.Duration
	)
	defer timeutil.CleanupTimer(reverify)
	defer timeutil.CleanupTimer(query)

Loop:
	for {
		select {
		// start verification, if not yet running
		case <-reverify.C:
			// if there is no reverifyDone, this means doReverify is not running
			if reverifyDone == nil {
				reverifyDone = make(chan struct{})
				go m.doReverify(reverifyDone)
			}

		// reset verification
		case <-reverifyDone:
			reverifyDone = nil
			reverify.Reset(reverifyInterval) // reverify again after the given interval

		// start requesting new peers, if no yet running
		case <-query.C:
			if queryNext == nil {
				queryNext = make(chan time.Duration)
				go m.doQuery(queryNext)
			}

		// on query done, reset time to given duration
		case d := <-queryNext:
			queryNext = nil
			query.Reset(d)

		// on close, exit the loop
		case <-m.closing:
			break Loop
		}
	}

	// wait for spawned goroutines to finish
	if reverifyDone != nil {
		<-reverifyDone
	}
	if queryNext != nil {
		<-queryNext
	}
}

// doReverify pings the oldest active peer.
func (m *manager) doReverify(done chan<- struct{}) {
	defer close(done)

	p := m.peerToReverify()
	if p == nil {
		return // nothing can be reverified
	}
	m.log.Debugw("reverifying",
		"peer", p,
	)

	// could not verify the peer
	if m.net.Ping(p) != nil {
		m.deletePeer(p.ID())

		return
	}

	// no need to do anything here, as the peer is bumped when handling the pong
}

// deletePeer deletes the peer with the given ID from the list of managed peers.
func (m *manager) deletePeer(id identity.ID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var mp *mpeer
	m.active, mp = deletePeerByID(m.active, id)
	if mp == nil {
		return // peer no longer exists
	}

	// master peers are never removed
	if containsPeer(m.masters, id) {
		// reset verifiedCount and re-add them to the front of the active peers
		mp.verifiedCount.Store(0)
		m.active = unshiftPeer(m.active, mp, maxManaged)

		return
	}

	m.log.Debugw("deleted",
		"peer", mp,
	)
	if mp.verifiedCount.Load() > 0 {
		m.events.PeerDeleted.Trigger(&PeerDeletedEvent{Peer: unwrapPeer(mp)})
	}

	// add a random replacement, if available
	if len(m.replacements) > 0 {
		var r *mpeer
		//nolint:gosec // we do not care about weak random numbers here
		m.replacements, r = deletePeer(m.replacements, rand.Intn(len(m.replacements)))
		m.active = pushPeer(m.active, r, maxManaged)
	}
}

// peerToReverify returns the oldest peer, or nil if empty.
func (m *manager) peerToReverify() *peer.Peer {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.active) == 0 {
		return nil
	}
	// the last peer is the oldest
	return unwrapPeer(m.active[len(m.active)-1])
}

// updatePeer moves the peer with the given ID to the front of the list of managed peers.
// It returns 0 if there was no peer with that id, otherwise the verifiedCount of the updated peer is returned.
func (m *manager) updatePeer(update *peer.Peer) uint {
	id := update.ID()
	for i, p := range m.active {
		if p.ID() == id {
			if i > 0 {
				//  move i-th peer to the front
				copy(m.active[1:], m.active[:i])
				m.active[0] = p
			}
			// update the wrapped peer and verifiedCount
			p.setPeer(update)

			return uint(p.verifiedCount.Add(1))
		}
	}

	return 0
}

func (m *manager) addReplacement(p *mpeer) bool {
	if containsPeer(m.replacements, p.ID()) {
		return false // already in the list
	}
	m.replacements = unshiftPeer(m.replacements, p, maxReplacements)

	return true
}

func (m *manager) loadInitialPeers() {
	var peers []*peer.Peer

	db := m.net.local().Database()
	if db != nil {
		peers = db.SeedPeers()
	}

	peers = append(peers, unwrapPeers(m.masters)...)
	for _, p := range peers {
		m.addDiscoveredPeer(p)
	}
}

// addDiscoveredPeer adds a newly discovered peer that has never been verified or pinged yet.
// It returns true, if the given peer was new and added, false otherwise.
func (m *manager) addDiscoveredPeer(p *peer.Peer) bool {
	// never add the local peer
	if p.ID() == m.self() {
		return false
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if containsPeer(m.active, p.ID()) {
		return false
	}
	m.log.Debugw("discovered",
		"peer", p,
	)

	mp := newMPeer(p)
	if len(m.active) >= maxManaged {
		return m.addReplacement(mp)
	}

	m.active = pushPeer(m.active, mp, maxManaged)

	return true
}

// addVerifiedPeer adds a new peer that has just been successfully pinged.
// It returns true, if the given peer was new and added, false otherwise.
//
//nolint:unparam // lets keep this for now
func (m *manager) addVerifiedPeer(p *peer.Peer) bool {
	// never add the local peer
	if p.ID() == m.self() {
		return false
	}

	m.log.Debugw("verified",
		"peer", p,
		"services", p.Services(),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// if already in the list, move it to the front
	if v := m.updatePeer(p); v > 0 {
		// trigger the event only for the first time the peer is updated
		if v == 1 {
			m.events.PeerDiscovered.Trigger(&PeerDiscoveredEvent{Peer: p})
		}

		return false
	}

	mp := newMPeer(p)
	mp.verifiedCount.Add(1)

	if len(m.active) >= maxManaged {
		return m.addReplacement(mp)
	}
	// trigger the event only when the peer is added to active
	m.events.PeerDiscovered.Trigger(&PeerDiscoveredEvent{Peer: p})

	// new nodes are added to the front
	m.active = unshiftPeer(m.active, mp, maxManaged)

	return true
}

// masterPeers returns the master peers.
func (m *manager) masterPeers() []*mpeer {
	return m.masters
}

// randomPeers returns a list of randomly selected peers.
func (m *manager) randomPeers(n int, minVerified uint) []*mpeer {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if n > len(m.active) {
		n = len(m.active)
	}

	peers := make([]*mpeer, 0, n)
	for _, i := range rand.Perm(len(m.active)) {
		if len(peers) == n {
			break
		}

		p := m.active[i]
		if uint(p.verifiedCount.Load()) < minVerified {
			continue
		}
		peers = append(peers, p)
	}

	return peers
}

// getVerifiedPeers returns all the currently managed peers that have been verified at least once.
func (m *manager) verifiedPeers() []*mpeer {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	peers := make([]*mpeer, 0, len(m.active))
	for _, mp := range m.active {
		if mp.verifiedCount.Load() == 0 {
			continue
		}
		peers = append(peers, mp)
	}

	return peers
}

// isKnown returns true if the manager is keeping track of that peer.
func (m *manager) isKnown(id identity.ID) bool {
	if id == m.self() {
		return true
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return containsPeer(m.active, id) || containsPeer(m.replacements, id)
}
