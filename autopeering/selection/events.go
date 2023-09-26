package selection

import (
	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/autopeering/salt"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/runtime/event"
)

// Events contains all the events that are triggered during the neighbor selection.
type Events struct {
	// A SaltUpdated event is triggered, when the private and public salt were updated.
	SaltUpdated *event.Event1[*SaltUpdatedEvent]
	// An OutgoingPeering event is triggered, when a valid response of PeeringRequest has been received.
	OutgoingPeering *event.Event1[*PeeringEvent]
	// An IncomingPeering event is triggered, when a valid PeerRequest has been received.
	IncomingPeering *event.Event1[*PeeringEvent]
	// A Dropped event is triggered, when a neighbor is dropped or when a drop message is received.
	Dropped *event.Event1[*DroppedEvent]
}

func newEvents() *Events {
	return &Events{
		SaltUpdated:     event.New1[*SaltUpdatedEvent](),
		OutgoingPeering: event.New1[*PeeringEvent](),
		IncomingPeering: event.New1[*PeeringEvent](),
		Dropped:         event.New1[*DroppedEvent](),
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
