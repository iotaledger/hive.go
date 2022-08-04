package selection

import (
	"github.com/iotaledger/hive.go/core/autopeering/peer"
	"github.com/iotaledger/hive.go/core/autopeering/salt"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/identity"
)

// Events contains all the events that are triggered during the neighbor selection.
type Events struct {
	// A SaltUpdated event is triggered, when the private and public salt were updated.
	SaltUpdated *event.Event[*SaltUpdatedEvent]
	// An OutgoingPeering event is triggered, when a valid response of PeeringRequest has been received.
	OutgoingPeering *event.Event[*PeeringEvent]
	// An IncomingPeering event is triggered, when a valid PeerRequest has been received.
	IncomingPeering *event.Event[*PeeringEvent]
	// A Dropped event is triggered, when a neighbor is dropped or when a drop message is received.
	Dropped *event.Event[*DroppedEvent]
}

func newEvents() (new *Events) {
	return &Events{
		SaltUpdated:     event.New[*SaltUpdatedEvent](),
		OutgoingPeering: event.New[*PeeringEvent](),
		IncomingPeering: event.New[*PeeringEvent](),
		Dropped:         event.New[*DroppedEvent](),
	}
}

// SaltUpdatedEvent bundles the information sent in the SaltUpdated event.
type SaltUpdatedEvent struct {
	Public, Private *salt.Salt // the updated salt
}

// PeeringEvent bundles the information sent in the OutgoingPeering and IncomingPeering event.
type PeeringEvent struct {
	Peer     *peer.Peer // peering partner
	Status   bool       // true, when the peering partner has accepted the request
	Distance uint32     // the distance between the peers
}

// DroppedEvent bundles the information sent in Dropped events.
type DroppedEvent struct {
	Peer      *peer.Peer
	DroppedID identity.ID // ID of the peer that gets dropped.
}
