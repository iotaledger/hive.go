package timeutil

import (
	"time"
)

// Ticker calls the handler with a period specified by the duration argument.
// It returns when the done channel is closed.
func Ticker(handler func(), interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop() // prevent the ticker from leaking

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			handler()
		}
	}
}
