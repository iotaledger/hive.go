package account

import (
	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/runtime/event"
)

// events is a collection of events that can be triggered by the SybilProtection.
type events[I index.Type] struct {
	// WeightUpdated is triggered when a weight of a node is updated.
	WeightsUpdated *event.Event1[*AccountsUpdateBatch[I]]

	// LinkableCollection is a generic trait that allows to link multiple collections of events together.
	event.Group[events[I], *events[I]]
}

func newEvents[I index.Type]() *events[I] {
	return event.CreateGroupConstructor(func() *events[I] {
		return &events[I]{
			WeightsUpdated: event.New1[*AccountsUpdateBatch[I]](),
		}
	})()
}
