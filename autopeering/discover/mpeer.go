package discover

import (
	"fmt"
	"sync/atomic"

	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/crypto/identity"
)

// mpeer represents a discovered peer with additional data.
// The fields of Peer may not be modified.
type mpeer struct {
	peer          atomic.Value
	verifiedCount atomic.Uint32 // how often that peer has been re-verified
	lastNewPeers  atomic.Uint32 // number of returned new peers when queried the last time
}

// Peer returns the wrapped peer.Peer.
func (m *mpeer) Peer() *peer.Peer {
	return m.peer.Load().(*peer.Peer)
}

// ID returns the ID of the wrapped peer.Peer.
func (m *mpeer) ID() identity.ID {
	return m.Peer().ID()
}

// String returns a string representation of the peer.
func (m *mpeer) String() string {
	return fmt.Sprintf("{%s, verifiedCount:%d, lastNewPeers:%d}", m.Peer(), m.verifiedCount.Load(), m.lastNewPeers.Load())
}

func (m *mpeer) setPeer(p *peer.Peer) {
	m.peer.Store(p)
}

func newMPeer(p *peer.Peer) *mpeer {
	m := new(mpeer)
	m.setPeer(p)

	return m
}

func wrapPeer(p *peer.Peer) *mpeer {
	return newMPeer(p)
}

func wrapPeers(ps []*peer.Peer) []*mpeer {
	result := make([]*mpeer, len(ps))
	for i, n := range ps {
		result[i] = wrapPeer(n)
	}

	return result
}

func unwrapPeer(p *mpeer) *peer.Peer {
	return p.Peer()
}

func unwrapPeers(ps []*mpeer) []*peer.Peer {
	result := make([]*peer.Peer, len(ps))
	for i, n := range ps {
		result[i] = unwrapPeer(n)
	}

	return result
}

// containsPeer returns true if a peer with the given ID is in the list.
func containsPeer(list []*mpeer, id identity.ID) bool {
	for _, p := range list {
		if p.ID() == id {
			return true
		}
	}

	return false
}

// unshiftPeer adds a new peer to the front of the list.
// If the list already contains max peers, the last is discarded.
func unshiftPeer(list []*mpeer, p *mpeer, max int) []*mpeer {
	if len(list) > max {
		panic(fmt.Sprintf("mpeer: invalid max value %d", max))
	}
	if len(list) < max {
		list = append(list, nil)
	}
	copy(list[1:], list)
	list[0] = p

	return list
}

// deletePeer is a helper that deletes the peer with the given index from the list.
func deletePeer(list []*mpeer, i int) ([]*mpeer, *mpeer) {
	if i >= len(list) {
		panic("mpeer: invalid index or empty mpeer list")
	}
	p := list[i]

	copy(list[i:], list[i+1:])
	list[len(list)-1] = nil

	return list[:len(list)-1], p
}

// deletePeerByID deletes the peer with the given ID from the list.
func deletePeerByID(list []*mpeer, id identity.ID) ([]*mpeer, *mpeer) {
	for i, p := range list {
		if p.ID() == id {
			return deletePeer(list, i)
		}
	}

	return list, nil
}

// pushPeer adds the given peer to the pack of the list.
// If the list already contains max peers, the first is discarded.
func pushPeer(list []*mpeer, p *mpeer, max int) []*mpeer {
	if len(list) > max {
		panic(fmt.Sprintf("mpeer: invalid max value %d", max))
	}
	if len(list) == max {
		copy(list, list[1:])
		list[len(list)-1] = p

		return list
	}

	return append(list, p)
}
