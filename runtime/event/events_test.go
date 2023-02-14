package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark(b *testing.B) {
	testEvent := New1[int]()
	testEvent.Hook(func(int) {})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testEvent.Trigger(i)
	}
}

func TestLink(t *testing.T) {
	sourceEvents := NewEvents()

	eventTriggered := 0
	linkedEvents := NewEvents(sourceEvents)
	linkedEvents.BlockIssued.Hook(func(int) {
		eventTriggered++
	})

	sourceEvents.BlockIssued.Trigger(7)
	require.Equal(t, eventTriggered, 1)

	linkedEvents.LinkTo(nil)

	sourceEvents.BlockIssued.Trigger(7)
	require.Equal(t, eventTriggered, 1)

	linkedEvents.LinkTo(sourceEvents)

	sourceEvents.BlockIssued.Trigger(7)
	require.Equal(t, eventTriggered, 2)
}

// Events represents events happening on a block factory.
type Events struct {
	// Triggered when a block is issued, i.e. sent to the protocol to be processed.
	BlockIssued *Event1[int]

	// Fired when an error occurred.
	Error *Event1[error]

	Group[Events, *Events]
}

// NewEvents contains the constructor of the Events object (it is generated by a generic factory).
var NewEvents = GroupConstructor(func() *Events {
	return &Events{
		BlockIssued: New1[int](),
		Error:       New1[error](),
	}
})
