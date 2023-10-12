package reactive

import "time"

// Clock is a reactive variable that automatically updates its value to the current time with the given granularity.
type Clock interface {
	// Variable is the variable that holds the current time.
	Variable[time.Time]

	// OnTick registers a callback that gets called when the clock ticks.
	OnTick(handler func(now time.Time)) (unsubscribe func())

	// Shutdown shuts down the Clock.
	Shutdown()
}

// NewClock creates a new Clock.
func NewClock(granularity time.Duration) Clock {
	return newClock(granularity)
}
