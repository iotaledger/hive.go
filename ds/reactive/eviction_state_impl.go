package reactive

import (
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

// evictionState is the default implementation of the EvictionState interface.
type evictionState[Type EvictionStateSlotType] struct {
	// lastEvictedSlot is the index of the last evicted slot.
	lastEvictedSlot Variable[Type]

	// evictionEvents is the map of all eviction events that were not evicted yet.
	evictionEvents *shrinkingmap.ShrinkingMap[Type, Event]
}

// newEvictionState creates a new evictionState instance.
func newEvictionState[Type EvictionStateSlotType]() *evictionState[Type] {
	return &evictionState[Type]{
		lastEvictedSlot: NewVariable[Type](),
		evictionEvents:  shrinkingmap.New[Type, Event](),
	}
}

// LastEvictedSlot returns a reactive variable that contains the index of the last evicted slot.
func (e *evictionState[Type]) LastEvictedSlot() Variable[Type] {
	return e.lastEvictedSlot
}

// EvictionEvent returns the event that is triggered when the given slot was evicted.
func (e *evictionState[Type]) EvictionEvent(slot Type) Event {
	evictionEvent := evictedSlotEvent

	e.lastEvictedSlot.Read(func(lastEvictedSlotIndex Type) {
		var zeroValue Type

		if slot > lastEvictedSlotIndex || (slot == zeroValue && lastEvictedSlotIndex == zeroValue) {
			evictionEvent, _ = e.evictionEvents.GetOrCreate(slot, NewEvent)
		}
	})

	return evictionEvent
}

// Evict evicts the given slot and triggers the corresponding eviction events.
func (e *evictionState[Type]) Evict(slot Type) {
	for _, slotEvictedEvent := range e.evict(slot) {
		slotEvictedEvent.Trigger()
	}
}

// evict advances the lastEvictedSlot to the given slot and returns the events that shall be triggered.
func (e *evictionState[Type]) evict(slot Type) (eventsToTrigger []Event) {
	e.lastEvictedSlot.Compute(func(lastEvictedSlotIndex Type) Type {
		if slot <= lastEvictedSlotIndex {
			return lastEvictedSlotIndex
		}

		for i := lastEvictedSlotIndex + Type(1); i <= slot; i++ {
			if slotEvictedEvent, exists := e.evictionEvents.Get(i); exists {
				eventsToTrigger = append(eventsToTrigger, slotEvictedEvent)
				e.evictionEvents.Delete(i)
			}
		}

		return slot
	})

	return eventsToTrigger
}

var evictedSlotEvent = func() Event {
	e := NewEvent()
	e.Trigger()

	return e
}()
