package reactive

// EvictionState is a reactive component that implements a slot based eviction mechanism.
type EvictionState[Type EvictionStateSlotType] interface {
	// LastEvictedSlot returns a reactive variable that contains the index of the last evicted slot.
	LastEvictedSlot() Type

	// EvictionEvent returns the event that is triggered when the given slot was evicted. It returns a triggered event
	// as default if the slot was already evicted.
	EvictionEvent(slot Type) Event

	// Evict evicts the given slot and triggers the corresponding eviction events.
	Evict(slot Type)
}

// NewEvictionState creates a new EvictionState instance.
func NewEvictionState[Type EvictionStateSlotType]() EvictionState[Type] {
	return newEvictionState[Type]()
}

// EvictionStateSlotType represents a constraint for the slot type of EvictionState.
type EvictionStateSlotType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}
