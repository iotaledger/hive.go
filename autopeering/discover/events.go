package discover

import (
	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/runtime/event"
)

// Events contains all the events that are triggered during the peer discovery.
type Events struct {
	// A PeerDiscovered event is triggered, when a new peer has been discovered and verified.
	PeerDiscovered *event.Event1[*PeerDiscoveredEvent]
	// A PeerDeleted event is triggered, when a discovered and verified peer could not be re-verified.
	PeerDeleted *event.Event1[*PeerDeletedEvent]
}

func newEvents() *Events {
	return &Events{
		PeerDiscovered: event.New1[*PeerDiscoveredEvent](),
		PeerDeleted:    event.New1[*PeerDeletedEvent](),
	}
}

// PeerDiscoveredEvent bundles the information of the discovered peer.
type PeerDiscoveredEvent struct {
	Peer *peer.Peer // discovered peer
}

// PeerDeletedEvent bundles the information of the deleted peer.
type PeerDeletedEvent struct {
	Peer *peer.Peer // deleted peer
}
