package event_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/generics/event"
)

// region LinkableEvents /////////////////////////////////////////////////////////////////////////////////////////////

// LinkableEvents is an example of a set of events that can be linked to other LinkableEvents, that is used for tests.
type LinkableEvents struct {
	// BlockMissing is triggered when a block is missing.
	BlockMissing *event.LinkableCollectionEvent[int, LinkableEvents, *LinkableEvents]

	// BlockReceived is triggered when a block is received.
	BlockReceived *event.LinkableCollectionEvent[bool, LinkableEvents, *LinkableEvents]

	// LinkableCollection imports the functionality to support link-ability between collections of the same type.
	event.LinkableCollection[LinkableEvents, *LinkableEvents]
}

// NewLinkableEvents is the constructor of the LinkableEvents object (it is generically generated).
var NewLinkableEvents = event.LinkableCollectionConstructor[LinkableEvents](func(e *LinkableEvents) {
	e.BlockMissing = event.NewLinkableCollectionEvent[int](e, func(target *LinkableEvents) { e.BlockMissing.LinkTo(target.BlockMissing) })
	e.BlockReceived = event.NewLinkableCollectionEvent[bool](e, func(target *LinkableEvents) { e.BlockReceived.LinkTo(target.BlockReceived) })
})

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Tests ////////////////////////////////////////////////////////////////////////////////////////////////////////

// TestLinkableEvents tests if switching between linked sources correctly works.
func TestLinkableEvents(t *testing.T) {
	// create source events
	sourceEvents1 := NewLinkableEvents()
	sourceEvents2 := NewLinkableEvents()
	linkedEvents := NewLinkableEvents(sourceEvents1)

	// create trigger counters
	var sourceEvents1BlockMissingTriggeredCount, sourceEvents2BlockMissingTriggeredCount, linkedEventsBlockMissingTriggeredCount uint64
	sourceEvents1.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&sourceEvents1BlockMissingTriggeredCount, 1) }))
	sourceEvents2.BlockMissing.Hook(event.NewClosure(func(int) { atomic.AddUint64(&sourceEvents2BlockMissingTriggeredCount, 1) }))
	linkedEvents.BlockMissing.Hook(event.NewClosure[int](func(event int) { atomic.AddUint64(&linkedEventsBlockMissingTriggeredCount, 1) }))

	// trigger on both source events (after linked through initialization)
	sourceEvents1.BlockMissing.Trigger(4)
	sourceEvents2.BlockMissing.Trigger(5)
	require.EqualValues(t, 1, sourceEvents1BlockMissingTriggeredCount)
	require.EqualValues(t, 1, sourceEvents2BlockMissingTriggeredCount)
	require.EqualValues(t, 1, linkedEventsBlockMissingTriggeredCount)

	linkedEvents.LinkTo(sourceEvents1)

	// trigger on chosen source event
	sourceEvents1.BlockMissing.Trigger(1)
	sourceEvents1.BlockMissing.Trigger(2)
	require.EqualValues(t, 3, sourceEvents1BlockMissingTriggeredCount)
	require.EqualValues(t, 1, sourceEvents2BlockMissingTriggeredCount)
	require.EqualValues(t, 3, linkedEventsBlockMissingTriggeredCount)

	linkedEvents.LinkTo(sourceEvents2)

	// trigger on other source event
	sourceEvents1.BlockMissing.Trigger(2)
	sourceEvents1.BlockMissing.Trigger(7)
	require.EqualValues(t, 5, sourceEvents1BlockMissingTriggeredCount)
	require.EqualValues(t, 1, sourceEvents2BlockMissingTriggeredCount)
	require.EqualValues(t, 3, linkedEventsBlockMissingTriggeredCount)

	// trigger on both source events (after linked through LinkTo)
	sourceEvents1.BlockMissing.Trigger(4)
	sourceEvents2.BlockMissing.Trigger(5)
	require.EqualValues(t, 6, sourceEvents1BlockMissingTriggeredCount)
	require.EqualValues(t, 2, sourceEvents2BlockMissingTriggeredCount)
	require.EqualValues(t, 4, linkedEventsBlockMissingTriggeredCount)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
