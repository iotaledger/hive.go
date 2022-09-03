package event_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/generics/event"
)

// region Events /////////////////////////////////////////////////////////////////////////////////////////////

// Events is an example of a set of events that can be linked to other Events.
type Events struct {
	// BlockMissing is triggered when a block is missing.
	BlockMissing *event.LinkableCollectionEvent[int, Events, *Events]

	// BlockReceived is triggered when a block is received.
	BlockReceived *event.LinkableCollectionEvent[bool, Events, *Events]

	event.LinkableCollection[Events, *Events]
}

// NewEvents contains the constructor of the Events object (it is generated by a generic factory).
var NewEvents = event.LinkableCollectionConstructor[Events](func(e *Events) {
	e.BlockMissing = event.NewLinkableCollectionEvent[int](e, func(target *Events) {
		e.BlockMissing.LinkTo(target.BlockMissing)
	})
	e.BlockReceived = event.NewLinkableCollectionEvent[bool](e, func(target *Events) {
		e.BlockReceived.LinkTo(target.BlockReceived)
	})
})

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Tests ////////////////////////////////////////////////////////////////////////////////////////////////////////

// TestLinkableEvents tests if switching between linked sources correctly works.
func TestLinkableEvents(t *testing.T) {
	// create source events
	internalEvents1 := NewEvents()
	internalEvents2 := NewEvents()
	publicEvents := NewEvents(internalEvents1)

	// create trigger counters
	var internalEvents1TriggeredCount, sourceEvents2TriggeredCount, publicEventsTriggerCount uint64
	internalEvents1.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&internalEvents1TriggeredCount, 1) }))
	internalEvents2.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&sourceEvents2TriggeredCount, 1) }))
	publicEvents.BlockMissing.Hook(event.NewClosure[int](func(event int) { atomic.AddUint64(&publicEventsTriggerCount, 1) }))

	// trigger on both source events (after linked through initialization)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	require.EqualValues(t, 1, internalEvents1TriggeredCount)
	require.EqualValues(t, 1, sourceEvents2TriggeredCount)
	require.EqualValues(t, 1, publicEventsTriggerCount)

	publicEvents.LinkTo(internalEvents2)

	// trigger on chosen source event
	internalEvents2.BlockMissing.Trigger(1)
	internalEvents2.BlockMissing.Trigger(2)
	require.EqualValues(t, 1, internalEvents1TriggeredCount)
	require.EqualValues(t, 3, sourceEvents2TriggeredCount)
	require.EqualValues(t, 3, publicEventsTriggerCount)

	publicEvents.LinkTo(internalEvents1)

	// trigger on other source event
	internalEvents2.BlockMissing.Trigger(2)
	internalEvents2.BlockMissing.Trigger(7)
	require.EqualValues(t, 1, internalEvents1TriggeredCount)
	require.EqualValues(t, 5, sourceEvents2TriggeredCount)
	require.EqualValues(t, 3, publicEventsTriggerCount)

	// trigger on both source events (after linked through LinkTo)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	require.EqualValues(t, 2, internalEvents1TriggeredCount)
	require.EqualValues(t, 6, sourceEvents2TriggeredCount)
	require.EqualValues(t, 4, publicEventsTriggerCount)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
