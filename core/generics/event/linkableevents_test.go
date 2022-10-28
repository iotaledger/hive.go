package event_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/generics/event"
)

// region Tests ////////////////////////////////////////////////////////////////////////////////////////////////////////

// TestLinkableEvents tests if switching between linked sources correctly works.
func TestLinkableEvents(t *testing.T) {
	// create source events
	internalEvents1 := NewEvents()
	internalEvents2 := NewEvents()
	publicEvents := NewEvents(internalEvents1)

	// create trigger counters
	var internal1TriggeredCount, internal1NestedTriggeredCount, internal2TriggeredCount, internal2NestedTriggeredCount,
		publicTriggerCount, publicNestedTriggerCount uint64

	// register triggers
	internalEvents1.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&internal1TriggeredCount, 1) }))
	internalEvents1.NestedEvents.BlockReceived.Hook(event.NewClosure(func(bool) { atomic.AddUint64(&internal1NestedTriggeredCount, 1) }))
	internalEvents2.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&internal2TriggeredCount, 1) }))
	internalEvents2.NestedEvents.BlockReceived.Hook(event.NewClosure(func(bool) { atomic.AddUint64(&internal2NestedTriggeredCount, 1) }))
	publicEvents.BlockMissing.Hook(event.NewClosure[int](func(event int) { atomic.AddUint64(&publicTriggerCount, 1) }))
	publicEvents.NestedEvents.BlockReceived.Hook(event.NewClosure(func(isBlockReceived bool) { atomic.AddUint64(&publicNestedTriggerCount, 1) }))

	// trigger on both source events (after linked through initialization)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	internalEvents1.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 1, internal1TriggeredCount)
	require.EqualValues(t, 1, internal1NestedTriggeredCount)
	require.EqualValues(t, 1, internal2TriggeredCount)
	require.EqualValues(t, 1, internal2NestedTriggeredCount)
	require.EqualValues(t, 1, publicTriggerCount)
	require.EqualValues(t, 1, publicNestedTriggerCount)

	publicEvents.LinkTo(internalEvents2)

	// trigger on chosen source event
	internalEvents2.BlockMissing.Trigger(1)
	internalEvents2.BlockMissing.Trigger(2)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 1, internal1TriggeredCount)
	require.EqualValues(t, 1, internal1NestedTriggeredCount)
	require.EqualValues(t, 3, internal2TriggeredCount)
	require.EqualValues(t, 3, internal2NestedTriggeredCount)
	require.EqualValues(t, 3, publicTriggerCount)
	require.EqualValues(t, 3, publicNestedTriggerCount)

	publicEvents.LinkTo(internalEvents1)

	// trigger on other source event
	internalEvents2.BlockMissing.Trigger(2)
	internalEvents2.BlockMissing.Trigger(7)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 1, internal1TriggeredCount)
	require.EqualValues(t, 1, internal1NestedTriggeredCount)
	require.EqualValues(t, 5, internal2TriggeredCount)
	require.EqualValues(t, 5, internal2NestedTriggeredCount)
	require.EqualValues(t, 3, publicTriggerCount)
	require.EqualValues(t, 3, publicNestedTriggerCount)

	// trigger on both source events (after linked through LinkTo)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	internalEvents1.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 2, internal1TriggeredCount)
	require.EqualValues(t, 2, internal1NestedTriggeredCount)
	require.EqualValues(t, 6, internal2TriggeredCount)
	require.EqualValues(t, 6, internal2NestedTriggeredCount)
	require.EqualValues(t, 4, publicTriggerCount)
	require.EqualValues(t, 4, publicNestedTriggerCount)

	publicEvents.LinkTo(nil)

	// trigger on both source events (after linked through LinkTo)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	internalEvents1.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 3, internal1TriggeredCount)
	require.EqualValues(t, 3, internal1NestedTriggeredCount)
	require.EqualValues(t, 7, internal2TriggeredCount)
	require.EqualValues(t, 7, internal2NestedTriggeredCount)
	require.EqualValues(t, 4, publicTriggerCount)
	require.EqualValues(t, 4, publicNestedTriggerCount)

	publicEvents.LinkTo(internalEvents1)

	// trigger on both source events (after linked through LinkTo)
	internalEvents1.BlockMissing.Trigger(4)
	internalEvents2.BlockMissing.Trigger(5)
	internalEvents1.NestedEvents.BlockReceived.Trigger(true)
	internalEvents2.NestedEvents.BlockReceived.Trigger(true)
	require.EqualValues(t, 4, internal1TriggeredCount)
	require.EqualValues(t, 4, internal1NestedTriggeredCount)
	require.EqualValues(t, 8, internal2TriggeredCount)
	require.EqualValues(t, 8, internal2NestedTriggeredCount)
	require.EqualValues(t, 5, publicTriggerCount)
	require.EqualValues(t, 5, publicNestedTriggerCount)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Events ///////////////////////////////////////////////////////////////////////////////////////////////////////

// Events is an example of a set of events that can be linked to other Events.
type Events struct {
	// BlockMissing is triggered when a block is missing.
	BlockMissing *event.Linkable[int]

	// NestedEvents is a nested set of events that can be linked to other Events.
	NestedEvents *NestedEvents

	// LinkableCollection imports the logic to make the struct linkable.
	event.LinkableCollection[Events, *Events]
}

// NewEvents is the constructor of the Events object (it is generated by a generic factory).
var NewEvents = event.LinkableConstructor(func() (newInstance *Events) {
	return &Events{
		BlockMissing: event.NewLinkable[int](),
		NestedEvents: NewNestedEvents(),
	}
})

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region NestedEvents /////////////////////////////////////////////////////////////////////////////////////////////////

// NewEvents is a nested set of Events that is embedded in the Events object.
type NestedEvents struct {
	// BlockReceived is triggered when a block is received.
	BlockReceived *event.Linkable[bool]

	// LinkableCollection imports the logic to make the struct linkable.
	event.LinkableCollection[NestedEvents, *NestedEvents]
}

// NewNestedEvents is the constructor of the NestedEvents object (it is generated by a generic factory).
var NewNestedEvents = event.LinkableConstructor(func() (newInstance *NestedEvents) {
	return &NestedEvents{
		BlockReceived: event.NewLinkable[bool](),
	}
})

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
