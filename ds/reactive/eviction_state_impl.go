package reactive

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/lo"
)

// evictionState is the default implementation of the EvictionState interface.
type evictionState[Type EvictionStateSlotType] struct {
	mutex sync.RWMutex

	// lastEvictedSlot is the index of the last evicted slot.
	lastEvictedSlot *Type

	// evictionEvents is the map of all eviction events that were not evicted yet.
	evictionEvents *shrinkingmap.ShrinkingMap[Type, Event]
}

// newEvictionState creates a new evictionState instance.
func newEvictionState[Type EvictionStateSlotType]() *evictionState[Type] {
	return &evictionState[Type]{
		evictionEvents: shrinkingmap.New[Type, Event](),
	}
}

func (e *evictionState[Type]) LastEvictedSlot() Type {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.lastEvictedSlot == nil {
		return 0
	}

	return *e.lastEvictedSlot
}

// EvictionEvent returns the event that is triggered when the given slot was evicted.
func (e *evictionState[Type]) EvictionEvent(slot Type) Event {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.lastEvictedSlot == nil || slot > *e.lastEvictedSlot {
		return lo.Return1(e.evictionEvents.GetOrCreate(slot, NewEvent))
	}

	return evictedSlotEvent
}

// Evict evicts the given slot and triggers the corresponding eviction events.
func (e *evictionState[Type]) Evict(slot Type) {
	for _, slotEvictedEvent := range e.evict(slot) {
		slotEvictedEvent.Trigger()
	}
}

// evict advances the lastEvictedSlot to the given slot and returns the events that shall be triggered.
func (e *evictionState[Type]) evict(slot Type) []Event {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.lastEvictedSlot != nil && slot <= *e.lastEvictedSlot {
		return nil
	}

	var startingSlot Type
	if e.lastEvictedSlot == nil {
		startingSlot = 0
	} else {
		startingSlot = *e.lastEvictedSlot + Type(1)
	}

	var eventsToTrigger []Event
	for i := startingSlot; i <= slot; i++ {
		if slotEvictedEvent, exists := e.evictionEvents.Get(i); exists {
			eventsToTrigger = append(eventsToTrigger, slotEvictedEvent)
			e.evictionEvents.Delete(i)
		}
	}

	e.lastEvictedSlot = &slot

	return eventsToTrigger
}

var evictedSlotEvent = func() Event {
	e := NewEvent()
	e.Trigger()

	return e
}()
