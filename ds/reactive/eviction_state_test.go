package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewEvictionState tests the newEvictionState function
func TestNewEvictionState(t *testing.T) {
	state := newEvictionState[int]()
	require.NotNil(t, state, "newEvictionState should not return nil")
	require.NotNil(t, state.evictionEvents, "Eviction events map should be initialized")
}

// TestLastEvictedSlot tests the LastEvictedSlot method
func TestLastEvictedSlot(t *testing.T) {
	state := newEvictionState[int]()
	slot := state.LastEvictedSlot()
	require.NotNil(t, slot, "LastEvictedSlot should not return nil")
}

// TestEvictionEvent tests the EvictionEvent method
func TestEvictionEvent(t *testing.T) {
	state := newEvictionState[int]()

	// Test with a slot that should create a new event
	event := state.EvictionEvent(1)
	require.NotNil(t, event, "EvictionEvent should not return nil for a new slot")

	// Test with a slot that should not create a new event
	state.lastEvictedSlot.Set(1)
	event = state.EvictionEvent(0)
	require.Equal(t, evictedSlotEvent, event, "EvictionEvent should return the evictedSlotEvent for a slot lower than lastEvictedSlot")
}

// TestEvict tests the Evict method
func TestEvict(t *testing.T) {
	state := newEvictionState[int]()

	// Setup a slot to be evicted
	slotToEvict := 1
	state.evictionEvents.Set(slotToEvict, NewEvent())

	// Evict the slot and verify
	state.Evict(slotToEvict)
	_, exists := state.evictionEvents.Get(slotToEvict)
	require.False(t, exists, "Evicted slot should no longer exist in evictionEvents")
}

// TestEvictPrivate tests the private evict method
func TestEvictTriggered(t *testing.T) {
	state := newEvictionState[int]()

	// Test with a slot less than the current lastEvictedSlotIndex
	state.lastEvictedSlot.Set(2)                  // Set the last evicted slot to 2
	state.evictionEvents.Set(1, evictedSlotEvent) // Set an event for slot 1
	eventsToTrigger := state.evict(1)             // Try to evict slot 1, which is less than the last evicted slot
	require.Len(t, eventsToTrigger, 0, "evict should not return any events when slot is less than or equal to lastEvictedSlotIndex")

	// Test with a slot equal to the current lastEvictedSlotIndex
	eventsToTrigger = state.evict(2) // Try to evict slot 2, which is equal to the last evicted slot
	require.Len(t, eventsToTrigger, 0, "evict should not return any events when slot is less than or equal to lastEvictedSlotIndex")

	// Test with a slot greater than the current lastEvictedSlotIndex
	slotToEvict := 3
	state.evictionEvents.Set(slotToEvict, NewEvent())
	eventToTrigger := state.EvictionEvent(slotToEvict)
	require.False(t, eventToTrigger.WasTriggered(), "evicted event should not have been triggered")
	state.Evict(slotToEvict)
	require.True(t, eventToTrigger.WasTriggered(), "evicted event should have been triggered")
}
