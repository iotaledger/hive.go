package reactive

import (
	"time"
)

// clock is the default implementation of the Clock interface.
type clock struct {
	// variable embeds variable that holds the current time.
	*variable[time.Time]

	// shutdown is used to signal the clock to shut down.
	shutdown chan bool
}

// newClock creates a new clock instance.
func newClock(granularity time.Duration) *clock {
	c := &clock{
		variable: newVariable[time.Time](),
		shutdown: make(chan bool, 1),
	}

	// set the initial value.
	c.variable.Set(time.Now())

	go func() {
		// align the ticker to the given granularity.
		time.Sleep(time.Duration(int64(granularity) - (time.Now().UnixNano() % int64(granularity))))
		ticker := time.NewTicker(granularity)

		// first tick after the initial value.
		c.variable.Set(time.Now().Truncate(granularity))

		for {
			select {
			case <-c.shutdown:
				return
			case t := <-ticker.C:
				c.variable.Set(t.Truncate(granularity))
			}
		}
	}()

	return c
}

// OnTick registers a callback that gets called when the clock ticks.
func (c *clock) OnTick(handler func(now time.Time)) (unsubscribe func()) {
	return c.OnUpdate(func(_, now time.Time) {
		handler(now)
	})
}

// Shutdown shuts down the Clock.
func (c *clock) Shutdown() {
	close(c.shutdown)
}
