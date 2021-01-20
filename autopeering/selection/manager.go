package selection

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/logger"
	"sync"
	"time"
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

	PeeringRequest(*peer.Peer, int) (bool, error)
	PeeringDrop(*peer.Peer)
}

type peeringRequest struct {
	peer    *peer.Peer
	channel int
}

type manager struct {
	net               network
	getPeersToConnect func() []*peer.Peer
	log               *logger.Logger
	dropOnUpdate      bool      // set true to drop all neighbors when the arrow is updated
	neighborValidator Validator // potential neighbor validator

	events   Events
	inbound  *Neighborhood
	outbound *Neighborhood

	rejectionFilter map[int]*Filter

	dropChan    chan identity.ID
	requestChan chan peeringRequest
	replyChan   chan bool

	wg      sync.WaitGroup
	closing chan struct{}
}

func newManager(net network, peersFunc func() []*peer.Peer, log *logger.Logger, opts *options) *manager {
	rejectionFilters := make(map[int]*Filter)
	for channel := 0; channel < outboundNeighborSize; channel++ {
		rejectionFilters[channel] = NewFilter()
	}
	return &manager{
		net:               net,
		getPeersToConnect: peersFunc,
		log:               log,
		dropOnUpdate:      opts.dropOnUpdate,
		neighborValidator: opts.neighborValidator,
		events: Events{
			ArRowUpdated:    events.NewEvent(arsUpdatedCaller),
			OutgoingPeering: events.NewEvent(peeringCaller),
			IncomingPeering: events.NewEvent(peeringCaller),
			Dropped:         events.NewEvent(droppedCaller),
		},
		inbound:         NewNeighborhood(inboundNeighborSize),
		outbound:        NewNeighborhood(outboundNeighborSize),
		rejectionFilter: rejectionFilters,
		dropChan:        make(chan identity.ID, queueSize),
		requestChan:     make(chan peeringRequest, queueSize),
		replyChan:       make(chan bool, 1),
		closing:         make(chan struct{}),
	}
}

func (m *manager) start() {
	if m.getArRow() == nil {
		m.updateArRow()
	}

	m.wg.Add(1)
	go m.loop()
}

func (m *manager) close() {
	close(m.closing)
	m.wg.Wait()
}

func (m *manager) getID() identity.ID {
	return m.net.local().ID()
}

func (m *manager) getArRow() *arrow.ArRow {
	return m.net.local().GetArRow()
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

func (m *manager) requestPeering(p *peer.Peer, channel int) bool {
	var status bool
	select {
	case m.requestChan <- peeringRequest{p, channel}:
		status = <-m.replyChan
	default:
		// a full queue should count as a failed request
		status = false
	}
	return status
}

func (m *manager) removeNeighbor(id identity.ID) {
	m.dropChan <- id
}

func (m *manager) loop() {
	defer m.wg.Done()

	var updateOutResultChan chan peer.PeerDistance
	updateTimer := time.NewTimer(0) // setting this to 0 will cause a trigger right away
	defer updateTimer.Stop()

Loop:
	for {
		select {

		// update the outbound neighbors
		case <-updateTimer.C:
			updateOutResultChan = make(chan peer.PeerDistance)
			// check arrow and update if necessary
			if m.getArRow().Expired() {
				m.updateArRow()
			}

			// check for new peers to connect to in a separate go routine
			go m.updateOutbound(updateOutResultChan)

		// handle the result of updateOutbound
		case req := <-updateOutResultChan:

			if req.Remote != nil {
				// if the peer is already in inbound, do not add it and remove it from inbound
				if p := m.inbound.RemovePeer(req.Remote.ID()); p != nil {
					m.triggerPeeringEvent(true, req.Channel, req.Remote, false)
					m.dropPeering(p)
				} else {
					m.addNeighbor(m.outbound, req)
					m.triggerPeeringEvent(true, req.Channel, req.Remote, true)
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

				//m.rejectionFilter[peerDistance.Channel].AddPeer(p.ID())
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
			m.triggerPeeringEvent(false, req.channel, req.peer, status)

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

func (m *manager) getUpdateTimeout() time.Duration {
	result := outboundUpdateInterval
	if m.outbound.IsFull() {
		result = fullOutboundUpdateInterval
	}
	arsExpiration := time.Until(m.getArRow().GetExpiration())
	if arsExpiration < result {
		result = arsExpiration
	}
	return result
}

// updateOutbound updates outbound neighbors.
func (m *manager) updateOutbound(resultChan chan<- peer.PeerDistance) {
	var result peer.PeerDistance
	defer func() { resultChan <- result }() // assure that a result is always sent to the channel
	now := time.Now().Unix()
	epoch := uint64(now - now%int64(arrowLifetime.Seconds()))
	for channel := 0; channel < outboundNeighborSize; channel++ {
		// sort verified peers by distance
		distList := peer.SortByOutbound(channel, m.getArRow(), m.getPeersToConnect(), epoch)

		// filter out current neighbors
		filter := m.getConnectedFilter()
		if m.neighborValidator != nil {
			filter.AddCondition(m.neighborValidator.IsValid)
		}

		// filter out previous rejections
		filteredList := m.rejectionFilter[channel].Apply(filter.Apply(distList))
		if len(filteredList) == 0 {
			continue
		}

		// select new candidate
		candidate := m.outbound.Select(filteredList, channel)
		if candidate.Remote == nil {
			continue
		}
		status, err := m.net.PeeringRequest(candidate.Remote, channel)
		if err != nil {
			m.rejectionFilter[channel].AddPeer(candidate.Remote.ID())
			m.log.Debugw("error requesting peering",
				"id", candidate.Remote.ID(),
				"addr", candidate.Remote.Address(), "err", err,
			)
			return
		}
		if !status {
			m.rejectionFilter[channel].AddPeer(candidate.Remote.ID())
			m.triggerPeeringEvent(true, channel, candidate.Remote, false)
			return
		}

		result = candidate
		break
	}
}

func (m *manager) handleInRequest(req peeringRequest) (resp bool) {
	resp = reject
	defer func() { m.replyChan <- resp }() // assure that a response is always issued

	if !m.isValidNeighbor(req.peer) {
		return
	}
	now := time.Now().Unix()
	epoch := uint64(now - now%int64(arrowLifetime.Seconds()))

	peerArs, _ := arrow.NewArRow(time.Until(m.getArRow().GetExpiration()), outboundNeighborSize, req.peer.Identity, epoch)
	reqDistance := peer.NewPeerDistance(m.getArRow().GetRows()[req.channel], peerArs.GetArs()[req.channel], req.channel, req.peer)
	filter := m.getConnectedFilter()
	filteredList := filter.Apply([]peer.PeerDistance{reqDistance})
	if len(filteredList) == 0 {
		return
	}

	toAccept := m.inbound.Select(filteredList, reqDistance.Channel)
	if toAccept.Remote == nil {
		return
	}

	m.addNeighbor(m.inbound, toAccept)
	resp = accept
	return
}

func (m *manager) addNeighbor(nh *Neighborhood, toAdd peer.PeerDistance) {
	// drop furthest neighbor if necessary
	if furthest, _ := nh.getFromChannel(toAdd.Channel); furthest.Remote != nil {
		if p := nh.RemovePeer(furthest.Remote.ID()); p != nil {
			m.dropPeering(p)
		}
	}
	nh.Add(toAdd)
}

func (m *manager) updateArRow() {
	now := time.Now().Unix()
	epoch := uint64(now - now%int64(arrowLifetime.Seconds()))
	newArRow, _ := arrow.NewArRow(arrowLifetime, outboundNeighborSize, m.net.local().Identity, epoch)
	m.net.local().SetArRow(newArRow)

	// clean the rejection filter
	for channel := range m.rejectionFilter {
		m.rejectionFilter[channel].Clean()
	}

	if !m.dropOnUpdate { // update distance without dropping neighbors
		m.outbound.UpdateOutboundDistance(m.getArRow())
		m.inbound.UpdateInboundDistance(m.getArRow())
	} else { // drop all the neighbors
		m.dropNeighborhood(m.inbound)
		m.dropNeighborhood(m.outbound)
	}

	m.log.Debugw("arrow updated",
		"arrow", arrowLifetime,
	)
	m.events.ArRowUpdated.Trigger(&ArRowUpdatedEvent{ArRow: newArRow})
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
	m.events.Dropped.Trigger(&DroppedEvent{DroppedID: p.ID()})
}

func (m *manager) getConnectedFilter() *Filter {
	filter := NewFilter()
	filter.AddPeer(m.getID())              //set filter for oneself
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
	if m.neighborValidator == nil {
		return true
	}
	return m.neighborValidator.IsValid(p)
}

func (m *manager) triggerPeeringEvent(isOut bool, channel int, p *peer.Peer, status bool) {
	if isOut {
		m.log.Debugw("peering requested",
			"direction", "out",
			"status", status,
			"to", p.ID(),
			"channel", channel,
			"#out", m.outbound,
			"#in", m.inbound,
		)
		m.events.OutgoingPeering.Trigger(&PeeringEvent{
			Peer:    p,
			Status:  status,
			Channel: channel,
		})
	} else {
		m.log.Debugw("peering requested",
			"direction", "in",
			"status", status,
			"from", p.ID(),
			"channel", channel,
			"#out", m.outbound,
			"#in", m.inbound,
		)
		m.events.IncomingPeering.Trigger(&PeeringEvent{
			Peer:    p,
			Status:  status,
			Channel: channel,
		})
	}
}
