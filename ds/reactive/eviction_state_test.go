package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEvictionEvent tests the EvictionEvent method
func TestEvictionEvent(t *testing.T) {
	state := newEvictionState[int]()

	// Test with a slot that should create a new event
	event := state.EvictionEvent(0)
	require.NotNil(t, event, "EvictionEvent should not return nil for a new slot")

	// Test with a slot that should create a new event
	event = state.EvictionEvent(1)
	require.NotNil(t, event, "EvictionEvent should not return nil for a new slot")

	// Test with a slot that should not create a new event
	state.Evict(1)
	event = state.EvictionEvent(0)
	require.Equal(t, evictedSlotEvent, event, "EvictionEvent should return the evictedSlotEvent for a slot lower than lastEvictedSlot")
}

// TestEvict tests the Evict method
func TestEvict(t *testing.T) {
	state := newEvictionState[int]()

	{
		// Setup a slot to be evicted
		slotToEvict := 0
		state.evictionEvents.Set(slotToEvict, NewEvent())

		// Evict the slot and verify
		state.Evict(slotToEvict)
		_, exists := state.evictionEvents.Get(slotToEvict)
		require.False(t, exists, "Evicted slot should no longer exist in evictionEvents")
	}

	{
		// Setup a slot to be evicted
		slotToEvict := 1
		state.evictionEvents.Set(slotToEvict, NewEvent())

		// Evict the slot and verify
		state.Evict(slotToEvict)
		_, exists := state.evictionEvents.Get(slotToEvict)
		require.False(t, exists, "Evicted slot should no longer exist in evictionEvents")
	}

}

// TestEvictPrivate tests the private evict method
func TestEvictTriggered(t *testing.T) {
	state := newEvictionState[int]()

	{
		var expectedEventsToTrigger []Event
		expectedEventsToTrigger = append(expectedEventsToTrigger, state.EvictionEvent(0))
		expectedEventsToTrigger = append(expectedEventsToTrigger, state.EvictionEvent(1))
		expectedEventsToTrigger = append(expectedEventsToTrigger, state.EvictionEvent(3))

		actualEventsToTrigger := state.evict(5)
		require.Equal(t, expectedEventsToTrigger, actualEventsToTrigger, "evict should return the correct events to trigger")
	}

	// Test with a slot smaller than current lastEvictedSlot
	{
		// Try to evict slot 0, which is less than the last evicted slot
		require.Len(t, state.evict(0), 0, "evict should not return any events when slot is less than or equal to lastEvictedSlotIndex")
		require.Len(t, state.evict(4), 0, "evict should not return any events when slot is less than or equal to lastEvictedSlotIndex")
		require.Len(t, state.evict(5), 0, "evict should not return any events when slot is less than or equal to lastEvictedSlotIndex")
	}

	{
		var expectedEventsToTrigger []Event
		expectedEventsToTrigger = append(expectedEventsToTrigger, state.EvictionEvent(6))
		expectedEventsToTrigger = append(expectedEventsToTrigger, state.EvictionEvent(8))

		actualEventsToTrigger := state.evict(10)
		require.Equal(t, expectedEventsToTrigger, actualEventsToTrigger, "evict should return the correct events to trigger")
	}

	// Requesting a EvictionEvent of an evicted slot should return the evictedSlotEvent
	{
		require.Equal(t, evictedSlotEvent, state.EvictionEvent(0), "EvictionEvent should return the evictedSlotEvent for a slot lower than lastEvictedSlot")
		require.Equal(t, evictedSlotEvent, state.EvictionEvent(10), "EvictionEvent should return the evictedSlotEvent for a slot lower than lastEvictedSlot")
	}
}
