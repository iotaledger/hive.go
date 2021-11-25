package selection

import (
	"github.com/iotaledger/hive.go/v2/autopeering/peer"
	"github.com/iotaledger/hive.go/v2/autopeering/salt"
	"github.com/iotaledger/hive.go/v2/events"
	"github.com/iotaledger/hive.go/v2/identity"
)

// Events contains all the events that are triggered during the neighbor selection.
type Events struct {
	// A SaltUpdated event is triggered, when the private and public salt were updated.
	SaltUpdated *events.Event
	// An OutgoingPeering event is triggered, when a valid response of PeeringRequest has been received.
	OutgoingPeering *events.Event
	// An IncomingPeering event is triggered, when a valid PeerRequest has been received.
	IncomingPeering *events.Event
	// A Dropped event is triggered, when a neighbor is dropped or when a drop message is received.
	Dropped *events.Event
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

func saltUpdatedCaller(handler interface{}, params ...interface{}) {
	handler.(func(*SaltUpdatedEvent))(params[0].(*SaltUpdatedEvent))
}

func peeringCaller(handler interface{}, params ...interface{}) {
	handler.(func(*PeeringEvent))(params[0].(*PeeringEvent))
}

func droppedCaller(handler interface{}, params ...interface{}) {
	handler.(func(*DroppedEvent))(params[0].(*DroppedEvent))
}
