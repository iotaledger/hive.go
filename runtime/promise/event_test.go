package promise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventTrigger(t *testing.T) {
	event := NewEvent()

	triggered := false
	event.OnTrigger(func() {
		triggered = true
	})

	event.Trigger()

	require.True(t, triggered, "event should be triggered")
}

func TestEventAlreadyTriggered(t *testing.T) {
	event := NewEvent()
	event.Trigger()

	triggered := false
	event.OnTrigger(func() {
		triggered = true
	})

	require.True(t, triggered, "event should be triggered immediately")
}

func TestEventUnsubscribe(t *testing.T) {
	event := NewEvent()

	triggered := false
	unsubscribe := event.OnTrigger(func() {
		triggered = true
	})

	unsubscribe()
	event.Trigger()

	require.False(t, triggered, "event should be triggered immediately")
}

func TestEvent1Trigger(t *testing.T) {
	event := NewEvent1[int]()

	triggered := false
	event.OnTrigger(func(val int) {
		require.Equal(t, 5, val, "event should be triggered with correct value")

		triggered = true
	})

	event.Trigger(5)

	require.True(t, triggered, "event should be triggered")
}

func TestEvent1AlreadyTriggered(t *testing.T) {
	event := NewEvent1[int]()
	event.Trigger(5)

	triggered := false
	event.OnTrigger(func(val int) {
		require.Equal(t, 5, val, "event should be triggered with correct value")

		triggered = true
	})

	require.True(t, triggered, "event should be triggered immediately")
}

func TestEvent1Unsubscribe(t *testing.T) {
	event := NewEvent1[int]()

	triggered := false
	unsubscribe := event.OnTrigger(func(val int) {
		require.Equal(t, 5, val, "event should be triggered with correct value")

		triggered = true
	})

	unsubscribe()
	event.Trigger(5)

	require.False(t, triggered, "event should be triggered immediately")
}
