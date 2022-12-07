package timeutil

import "time"

// CleanupTimer stops the timer and drains the Timer's channel.
// This cannot be done concurrent to other receives from the Timer's
// channel or other calls to the Timer's Stop method.
func CleanupTimer(t *time.Timer) {
	// prevent the timer from firing
	t.Stop()

	select {
	case <-t.C:
		// drain the channel in case the timer fired
	default:
		// do not block if channel is already empty
	}
}
