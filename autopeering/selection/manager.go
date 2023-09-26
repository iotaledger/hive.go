package selection

import (
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v2"
	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/autopeering/mana"
	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/autopeering/salt"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/timeutil"
)

const (
	accept = true
	reject = false

	// buffer size of the channels handling inbound requests and drops.
	queueSize = 10
)

// A network represents the communication layer for the manager.
type network interface {
	local() *peer.Local

	PeeringRequest(*peer.Peer, *salt.Salt) (bool, error)
	PeeringDrop(*peer.Peer)
}

type peeringRequest struct {
	peer *peer.Peer
	salt *salt.Salt
	back chan bool
}

type manager struct {
	net               network
	getPeersToConnect func() []*peer.Peer
	log               *logger.Logger
	dropOnUpdate      bool      // set true to drop all neighbors when the salt is updated
	neighborValidator Validator // potential neighbor validator

	useMana          bool // set true to use mana
	manaFunc         mana.Func
	rankedPeers      []*peer.Peer // valid peers ranked by mana
	rankedPeersMutex sync.RWMutex
	r                int
	ro               float64

	Events   *Events
	inbound  *Neighborhood
	outbound *Neighborhood

	skiplist  *ttlcache.Cache
	blocklist *ttlcache.Cache

	dropChan    chan identity.ID
	requestChan chan peeringRequest

	wg      sync.WaitGroup
	closing chan struct{}
}

func newManager(net network, peersFunc func() []*peer.Peer, log *logger.Logger, opts *options) *manager {
	blocklist := ttlcache.NewCache()
	if err := blocklist.SetTTL(opts.neighborBlockDuration); err != nil {
		log.Panicw("Failed to set TTL for neighbors blocklist cache", "err", err)
	}
	skiplist := ttlcache.NewCache()
	if err := skiplist.SetTTL(opts.neighborSkipTimeout); err != nil {
		log.Panicw("Failed to set TTL for neighbors skiplist cache", "err", err)
	}

	return &manager{
		net:               net,
		getPeersToConnect: peersFunc,
		log:               log,
		dropOnUpdate:      opts.dropOnUpdate,
		neighborValidator: opts.neighborValidator,
		useMana:           opts.useMana,
		manaFunc:          opts.manaFunc,
		rankedPeers:       []*peer.Peer{},
		r:                 opts.r,
		ro:                opts.ro,
		inbound:           NewNeighborhood(inboundNeighborSize),
		outbound:          NewNeighborhood(outboundNeighborSize),
		skiplist:          skiplist,
		blocklist:         blocklist,
		dropChan:          make(chan identity.ID, queueSize),
		requestChan:       make(chan peeringRequest, queueSize),
		closing:           make(chan struct{}),
		Events:            newEvents(),
	}
}

func (m *manager) start() {
	if m.getPublicSalt() == nil || m.getPrivateSalt() == nil {
		m.updateSalt()
	}
	m.wg.Add(1)
	go m.loop()
}

func (m *manager) close() {
	close(m.closing)
	m.wg.Wait()
	if err := m.blocklist.Close(); err != nil {
		m.log.Warnw("Failed to close blocklist cache", "err", err)
	}
	if err := m.skiplist.Close(); err != nil {
		m.log.Warnw("Failed to close skiplist cache", "err", err)
	}
}

func (m *manager) getID() identity.ID {
	return m.net.local().ID()
}

func (m *manager) getPublicSalt() *salt.Salt {
	return m.net.local().GetPublicSalt()
}

func (m *manager) getPrivateSalt() *salt.Salt {
	return m.net.local().GetPrivateSalt()
}

func (m *manager) getNeighbors() []*peer.Peer {
	var neighbors []*peer.Peer
	neighbors = append(neighbors, m.inbound.GetPeers()...)
	neighbors = append(neighbors, m.outbound.GetPeers()...)

	return neighbors
}

func (m *manager) getInNeighbors() []*peer.Peer {
	return m.inbound.GetPeers()
}

func (m *manager) getOutNeighbors() []*peer.Peer {
	return m.outbound.GetPeers()
}

func (m *manager) requestPeering(p *peer.Peer, s *salt.Salt) bool {
	var status bool

	back := make(chan bool)
	select {
	case m.requestChan <- peeringRequest{peer: p, salt: s, back: back}:
		status = <-back
	default:
		// a full queue should count as a failed request
		status = false
	}

	return status
}

func (m *manager) blockNeighbor(id identity.ID, ttl ...time.Duration) {
	if len(ttl) > 0 {
		if err := m.blocklist.SetWithTTL(id.EncodeBase58(), nil, ttl[0]); err != nil {
			m.log.Warnw("Failed to set neighbor to blocklist cache", "err", err)
		}
	} else {
		if err := m.blocklist.Set(id.EncodeBase58(), nil); err != nil {
			m.log.Warnw("Failed to set neighbor to blocklist cache", "err", err)
		}
	}

	m.removeNeighbor(id)
}

func (m *manager) unblockNeighbor(id identity.ID) {
	if err := m.blocklist.Remove(id.EncodeBase58()); err != nil && !errors.Is(err, ttlcache.ErrNotFound) {
		m.log.Warnw("Failed to remove neighbor from blocklist cache", "err", err)
	}

	// we need to remove the neighbor from the skiplist as well,
	// because they are added to the skiplist at first connection attempt if they were blocked.
	if err := m.skiplist.Remove(id.EncodeBase58()); err != nil && !errors.Is(err, ttlcache.ErrNotFound) {
		m.log.Warnw("Failed to remove neighbor from skiplist cache", "err", err)
	}
}

func (m *manager) removeNeighbor(id identity.ID) {
	m.dropChan <- id
}

func (m *manager) loop() {
	defer m.wg.Done()

	var updateOutResultChan chan peer.PeerDistance
	updateTimer := time.NewTimer(0) // setting this to 0 will cause a trigger right away
	defer timeutil.CleanupTimer(updateTimer)

Loop:
	for {
		select {

		// update the outbound neighbors
		case <-updateTimer.C:
			updateOutResultChan = make(chan peer.PeerDistance)
			// check salt and update if necessary
			if m.getPublicSalt().Expired() {
				m.updateSalt()
			}
			// check for new peers to connect to in a separate go routine
			go m.updateOutbound(updateOutResultChan)

		// handle the result of updateOutbound
		case req := <-updateOutResultChan:
			if req.Remote != nil {
				// if the peer is already in inbound, do not add it and remove it from inbound
				if p := m.inbound.RemovePeer(req.Remote.ID()); p != nil {
					m.triggerPeeringEvent(true, req.Remote, false)
					m.dropPeering(p)
				} else {
					added := m.addNeighbor(m.outbound, req)
					m.triggerPeeringEvent(true, req.Remote, added)
				}
			}
			// call updateOutbound again after the given interval
			updateOutResultChan = nil
			updateTimer.Reset(m.getUpdateTimeout())

		// handle a drop request
		case id := <-m.dropChan:
			droppedPeer := m.inbound.RemovePeer(id)
			if p := m.outbound.RemovePeer(id); p != nil {
				droppedPeer = p
				m.addToSkiplist(id)
				// if not yet updating, trigger an immediate update
				if updateOutResultChan == nil && updateTimer.Stop() {
					updateTimer.Reset(0)
				}
			}
			if droppedPeer != nil {
				m.dropPeering(droppedPeer)
			}

		// handle an inbound request
		case req := <-m.requestChan:
			status := m.handleInRequest(req)
			// trigger in the main loop to guarantee order of events
			m.triggerPeeringEvent(false, req.peer, status)

		// on close, exit the loop
		case <-m.closing:
			break Loop
		}
	}

	// wait for the updateOutbound to finish
	if updateOutResultChan != nil {
		<-updateOutResultChan
	}
}

func (m *manager) addToSkiplist(id identity.ID) {
	m.log.Debugw("Adding neighbor to skiplist", "peerId", id)
	if err := m.skiplist.Set(id.EncodeBase58(), nil); err != nil {
		m.log.Warnw("Failed to set neighbor to skiplist cache", "err", err)
	}
}

func (m *manager) cleanSkiplist() {
	if err := m.skiplist.Purge(); err != nil {
		m.log.Warnw("Failed to purge neighbor skiplist cache", "err", err)
	}
}

func (m *manager) getUpdateTimeout() time.Duration {
	result := outboundUpdateInterval
	if m.outbound.IsFull() {
		result = fullOutboundUpdateInterval
	}
	saltExpiration := time.Until(m.getPublicSalt().GetExpiration())
	if saltExpiration < result {
		result = saltExpiration
	}

	return result
}

// updateOutbound updates outbound neighbors.
func (m *manager) updateOutbound(resultChan chan<- peer.PeerDistance) {
	var result peer.PeerDistance
	defer func() { resultChan <- result }() // assure that a result is always sent to the channel
	candidate := m.getOutboundPeeringCandidate()
	if candidate.Remote == nil {
		return
	}

	status, err := m.net.PeeringRequest(candidate.Remote, m.getPublicSalt())
	if err != nil {
		m.addToSkiplist(candidate.Remote.ID())
		m.log.Debugw("error requesting peering",
			"id", candidate.Remote.ID(),
			"addr", candidate.Remote.Address(), "err", err,
		)

		return
	}
	if !status {
		m.addToSkiplist(candidate.Remote.ID())
		m.triggerPeeringEvent(true, candidate.Remote, false)

		return
	}

	result = candidate
}

func (m *manager) getOutboundPeeringCandidate() (candidate peer.PeerDistance) {
	knownPeers := m.getPeersToConnect()

	if m.useMana {
		// extract identities from known peers
		identities := []*identity.Identity{}
		peerMap := make(map[*identity.Identity]*peer.Peer)

		for _, peer := range knownPeers {
			identities = append(identities, peer.Identity)
			peerMap[peer.Identity] = peer
		}

		// rank known peers by mana
		manaRank := mana.RankByVariableRange(m.manaFunc, m.net.local().Identity, identities, m.r, m.ro)

		// retrieve peers from identities
		m.rankedPeersMutex.Lock()
		m.rankedPeers = []*peer.Peer{}
		for _, identity := range manaRank {
			m.rankedPeers = append(m.rankedPeers, peerMap[identity])
		}
		m.rankedPeersMutex.Unlock()

		m.rankedPeersMutex.RLock()
		knownPeers = m.rankedPeers
		m.rankedPeersMutex.RUnlock()
	}

	// Filter out blocklisted peers.
	allowedPeers := make([]*peer.Peer, 0, len(knownPeers))
	for _, p := range knownPeers {
		if !m.isInBlocklist(p.ID()) {
			allowedPeers = append(allowedPeers, p)
		}
	}

	// sort verified peers by distance
	distList := peer.SortBySalt(lo.PanicOnErr(m.getID().Bytes()), m.getPublicSalt().GetBytes(), allowedPeers)

	// filter out current neighbors
	filter := m.getConnectedFilter()
	if m.neighborValidator != nil {
		filter.AddCondition(m.neighborValidator.IsValid)
	}
	distList = filter.Apply(distList)

	// filter out previous rejections
	filteredList := make([]peer.PeerDistance, 0, len(distList))
	for _, dist := range distList {
		if !m.isInSkiplist(dist.Remote.ID()) {
			filteredList = append(filteredList, dist)
		}
	}

	if len(filteredList) == 0 {
		return
	}
	// reset rejectionFilter so that in the next call filteredList is full again
	if len(filteredList) < 2 {
		m.cleanSkiplist()
	}

	// select new candidate
	candidate = m.outbound.Select(filteredList)

	return candidate
}

func (m *manager) handleInRequest(req peeringRequest) (resp bool) {
	resp = reject
	defer func() { req.back <- resp }() // assure that a response is always issued

	if m.isInBlocklist(req.peer.ID()) {
		return
	}

	if !m.isValidNeighbor(req.peer) {
		return
	}

	reqDistance := peer.NewPeerDistance(lo.PanicOnErr(m.getID().Bytes()), m.getPrivateSalt().GetBytes(), req.peer)
	filter := m.getConnectedFilter()
	filteredList := filter.Apply([]peer.PeerDistance{reqDistance})
	if len(filteredList) == 0 {
		return
	}

	toAccept := m.inbound.Select(filteredList)
	if toAccept.Remote == nil {
		return
	}

	if m.addNeighbor(m.inbound, toAccept) {
		resp = accept
	}

	return
}

func (m *manager) addNeighbor(nh *Neighborhood, toAdd peer.PeerDistance) bool {
	// drop furthest neighbor if necessary
	if furthest, _ := nh.getFurthest(); furthest.Remote != nil {
		if p := nh.RemovePeer(furthest.Remote.ID()); p != nil {
			m.dropPeering(p)
		}
	}

	return nh.Add(toAdd)
}

func (m *manager) updateSalt() {
	public, _ := salt.NewSalt(saltLifetime)
	m.net.local().SetPublicSalt(public)
	private, _ := salt.NewSalt(saltLifetime)
	m.net.local().SetPrivateSalt(private)

	// clean the rejection filter
	m.cleanSkiplist()

	if !m.dropOnUpdate { // update distance without dropping neighbors
		m.outbound.UpdateDistance(lo.PanicOnErr(m.getID().Bytes()), m.getPublicSalt().GetBytes())
		m.inbound.UpdateDistance(lo.PanicOnErr(m.getID().Bytes()), m.getPrivateSalt().GetBytes())
	} else { // drop all the neighbors
		m.dropNeighborhood(m.inbound)
		m.dropNeighborhood(m.outbound)
	}

	m.log.Debugw("salt updated",
		"public", saltLifetime,
		"private", saltLifetime,
	)
	m.Events.SaltUpdated.Trigger(&SaltUpdatedEvent{Public: public, Private: private})
}

func (m *manager) dropNeighborhood(nh *Neighborhood) {
	for _, p := range nh.GetPeers() {
		nh.RemovePeer(p.ID())
		m.dropPeering(p)
	}
}

// dropPeering sends the peering drop over the network and triggers the corresponding event.
func (m *manager) dropPeering(p *peer.Peer) {
	m.net.PeeringDrop(p)

	m.log.Debugw("peering dropped",
		"id", p.ID(),
		"#out", m.outbound,
		"#in", m.inbound,
	)
	m.Events.Dropped.Trigger(&DroppedEvent{Peer: p, DroppedID: p.ID()})
}

func (m *manager) getConnectedFilter() *Filter {
	filter := NewFilter()
	filter.AddPeer(m.getID())              // set filter for oneself
	filter.AddPeers(m.inbound.GetPeers())  // set filter for inbound neighbors
	filter.AddPeers(m.outbound.GetPeers()) // set filter for outbound neighbors

	return filter
}

// isValidNeighbor returns whether the given peer is a valid neighbor candidate.
func (m *manager) isValidNeighbor(p *peer.Peer) bool {
	// do not connect to oneself
	if m.getID() == p.ID() {
		return false
	}

	// do not connect if request comes from outside of manaRank
	if m.useMana {
		exist := false
		m.rankedPeersMutex.RLock()
		for _, peer := range m.rankedPeers {
			if p.ID() == peer.ID() {
				exist = true

				break
			}
		}
		m.rankedPeersMutex.RUnlock()
		if !exist {
			return false
		}
	}

	if m.neighborValidator == nil {
		return true
	}

	return m.neighborValidator.IsValid(p)
}

func (m *manager) triggerPeeringEvent(isOut bool, p *peer.Peer, status bool) {
	if isOut {
		m.log.Debugw("peering requested",
			"direction", "out",
			"status", status,
			"to", p.ID(),
			"#out", m.outbound,
			"#in", m.inbound,
		)
		m.Events.OutgoingPeering.Trigger(&PeeringEvent{
			Peer:     p,
			Status:   status,
			Distance: peer.NewPeerDistance(lo.PanicOnErr(m.getID().Bytes()), m.getPublicSalt().GetBytes(), p).Distance,
		})
	} else {
		m.log.Debugw("peering requested",
			"direction", "in",
			"status", status,
			"from", p.ID(),
			"#out", m.outbound,
			"#in", m.inbound,
		)
		m.Events.IncomingPeering.Trigger(&PeeringEvent{
			Peer:     p,
			Status:   status,
			Distance: peer.NewPeerDistance(lo.PanicOnErr(m.getID().Bytes()), m.getPrivateSalt().GetBytes(), p).Distance,
		})
	}
}

func (m *manager) isInBlocklist(id identity.ID) bool {
	if _, err := m.blocklist.Get(id.EncodeBase58()); err != nil {
		if !errors.Is(err, ttlcache.ErrNotFound) {
			m.log.Warnw("Failed to retrieve record for peer from blocklist cache",
				"peerId", id)
		}

		return false
	}

	return true
}

func (m *manager) isInSkiplist(id identity.ID) bool {
	if _, err := m.skiplist.Get(id.EncodeBase58()); err != nil {
		if !errors.Is(err, ttlcache.ErrNotFound) {
			m.log.Warnw("Failed to retrieve record for peer from skiplist cache",
				"peerId", id)
		}

		return false
	}

	return true
}
