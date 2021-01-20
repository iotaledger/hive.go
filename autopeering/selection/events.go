package selection

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/identity"
)

// Events contains all the events that are triggered during the neighbor selection.
type Events struct {
	// A ArRowUpdated event is triggered, when the ArRow values are updated.
	ArRowUpdated *events.Event

	// An OutgoingPeering event is triggered, when a valid response of PeeringRequest has been received.
	OutgoingPeering *events.Event
	// An IncomingPeering event is triggered, when a valid PeerRequest has been received.
	IncomingPeering *events.Event
	// A Dropped event is triggered, when a neighbor is dropped or when a drop message is received.
	Dropped *events.Event
}

// ArRowUpdatedEvent bundles the information sent in the ArRowUpdated event.
type ArRowUpdatedEvent struct {
	ArRow *arrow.ArRow // the updated arrow
}

// PeeringEvent bundles the information sent in the OutgoingPeering and IncomingPeering event.
type PeeringEvent struct {
	Peer    *peer.Peer // peering partner
	Status  bool       // true, when the peering partner has accepted the request
	Channel int        //  channel on which nodes tried to connect.
}

// DroppedEvent bundles the information sent in Dropped events.
type DroppedEvent struct {
	DroppedID identity.ID // ID of the peer that gets dropped.
}

func arsUpdatedCaller(handler interface{}, params ...interface{}) {
	handler.(func(*ArRowUpdatedEvent))(params[0].(*ArRowUpdatedEvent))
}
func peeringCaller(handler interface{}, params ...interface{}) {
	handler.(func(*PeeringEvent))(params[0].(*PeeringEvent))
}

func droppedCaller(handler interface{}, params ...interface{}) {
	handler.(func(*DroppedEvent))(params[0].(*DroppedEvent))
}
