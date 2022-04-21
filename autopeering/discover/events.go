package discover

import (
	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/generics/event"
)

// Events contains all the events that are triggered during the peer discovery.
type Events struct {
	// A PeerDiscovered event is triggered, when a new peer has been discovered and verified.
	PeerDiscovered *event.Event[*PeerDiscoveredEvent]
	// A PeerDeleted event is triggered, when a discovered and verified peer could not be re-verified.
	PeerDeleted *event.Event[*PeerDeletedEvent]
}

func newEvents() (new *Events) {
	return &Events{
		PeerDiscovered: event.New[*PeerDiscoveredEvent](),
		PeerDeleted:    event.New[*PeerDeletedEvent](),
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
