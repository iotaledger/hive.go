package selection

import (
	"fmt"
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/autopeering/distance"
	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/identity"
)

type Neighborhood struct {
	neighbors []peer.PeerDistance
	size      int
	mu        sync.RWMutex
}

func NewNeighborhood(size int) *Neighborhood {
	return &Neighborhood{
		neighbors: []peer.PeerDistance{},
		size:      size,
	}
}

func (nh *Neighborhood) String() string {
	return fmt.Sprintf("%d/%d", nh.GetNumPeers(), nh.size)
}

func (nh *Neighborhood) getFromChannel(channel int) (peer.PeerDistance, int) {
	nh.mu.RLock()
	defer nh.mu.RUnlock()

	channelConnected := false
	index := 0
	furthest := peer.PeerDistance{
		Remote:   nil,
		Channel:  channel,
		Distance: 0,
	}
	for i, n := range nh.neighbors {
		if n.Channel == channel && n.Distance > furthest.Distance {
			furthest = n
			index = i
			channelConnected = true
		}
	}
	if !channelConnected {
		return peer.PeerDistance{
			Remote:   nil,
			Distance: distance.Max,
			Channel:  channel,
		}, len(nh.neighbors)
	}

	return furthest, index
}

// Select returns peer with candidate to replace existing connection on a given channel.
func (nh *Neighborhood) Select(candidates []peer.PeerDistance, channel int) peer.PeerDistance {
	if len(candidates) > 0 {
		target, _ := nh.getFromChannel(channel)
		for _, candidate := range candidates {
			if candidate.Distance < target.Distance {
				return candidate
			}
		}
	}
	return peer.PeerDistance{}
}

// Add tries to add a new peer with distance to the neighborhood.
// It returns true, if the peer was added, or false if the neighborhood was full.
func (nh *Neighborhood) Add(toAdd peer.PeerDistance) bool {
	nh.mu.Lock()
	defer nh.mu.Unlock()
	if len(nh.neighbors) >= nh.size {
		return false
	}
	nh.neighbors = append(nh.neighbors, toAdd)
	return true
}

// RemovePeer removes the peer with the given ID from the neighborhood.
// It returns the peer that was removed or nil of no such peer exists.
func (nh *Neighborhood) RemovePeer(id identity.ID) *peer.Peer {
	nh.mu.Lock()
	defer nh.mu.Unlock()

	index := nh.getPeerIndex(id)
	if index < 0 {
		return nil
	}
	n := nh.neighbors[index]

	// remove index from slice
	if index < len(nh.neighbors)-1 {
		copy(nh.neighbors[index:], nh.neighbors[index+1:])
	}
	nh.neighbors[len(nh.neighbors)-1] = peer.PeerDistance{}
	nh.neighbors = nh.neighbors[:len(nh.neighbors)-1]

	return n.Remote
}

func (nh *Neighborhood) getPeerIndex(id identity.ID) int {
	for i, p := range nh.neighbors {
		if p.Remote.ID() == id {
			return i
		}
	}
	return -1
}

// UpdateInboundDistance updates distances of incoming connections.
func (nh *Neighborhood) UpdateInboundDistance(localArs *arrow.ArRow) {
	nh.mu.Lock()
	defer nh.mu.Unlock()
	now := time.Now().Unix()
	epoch := uint64(now - now%int64(arrowLifetime.Seconds()))
	for i, n := range nh.neighbors {
		peerArs, _ := arrow.NewArRow(localArs.GetExpiration().Sub(time.Now()), outboundNeighborSize, n.Remote.Identity, epoch)
		nh.neighbors[i].Distance = distance.ByArs(localArs.GetRows()[n.Channel], peerArs.GetArs()[n.Channel])
	}
}

// UpdateOutboundDistance updates distances of outgoing connections.
func (nh *Neighborhood) UpdateOutboundDistance(localArs *arrow.ArRow) {
	nh.mu.Lock()
	defer nh.mu.Unlock()
	now := time.Now().Unix()
	epoch := uint64(now - now%int64(arrowLifetime.Seconds()))
	for i, n := range nh.neighbors {
		peerArs, _ := arrow.NewArRow(localArs.GetExpiration().Sub(time.Now()), outboundNeighborSize, n.Remote.Identity, epoch)
		nh.neighbors[i].Distance = distance.ByArs(localArs.GetArs()[n.Channel], peerArs.GetRows()[n.Channel])
	}
}
func (nh *Neighborhood) IsFull() bool {
	nh.mu.RLock()
	defer nh.mu.RUnlock()
	return len(nh.neighbors) >= nh.size
}

func (nh *Neighborhood) GetPeers() []*peer.Peer {
	nh.mu.RLock()
	defer nh.mu.RUnlock()
	result := make([]*peer.Peer, len(nh.neighbors))
	for i, n := range nh.neighbors {
		result[i] = n.Remote
	}
	return result
}

func (nh *Neighborhood) GetNumPeers() int {
	nh.mu.RLock()
	defer nh.mu.RUnlock()
	return len(nh.neighbors)
}
