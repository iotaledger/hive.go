package discover

import (
	"github.com/iotaledger/hive.go/v2/autopeering/peer"
	"github.com/iotaledger/hive.go/v2/events"
)

// Events contains all the events that are triggered during the peer discovery.
type Events struct {
	// A PeerDiscovered event is triggered, when a new peer has been discovered and verified.
	PeerDiscovered *events.Event
	// A PeerDeleted event is triggered, when a discovered and verified peer could not be re-verified.
	PeerDeleted *events.Event
}

// DiscoveredEvent bundles the information of the discovered peer.
type DiscoveredEvent struct {
	Peer *peer.Peer // discovered peer
}

// DeletedEvent bundles the information of the deleted peer.
type DeletedEvent struct {
	Peer *peer.Peer // deleted peer
}

func peerDiscovered(handler interface{}, params ...interface{}) {
	handler.(func(*DiscoveredEvent))(params[0].(*DiscoveredEvent))
}

func peerDeleted(handler interface{}, params ...interface{}) {
	handler.(func(*DeletedEvent))(params[0].(*DeletedEvent))
}
