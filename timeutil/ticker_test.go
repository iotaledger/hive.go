package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func TestTicker_ExternalShutdownSignal(t *testing.T) {
	// use counter to track execution state
	counter := atomic.NewUint64(0)

	// create "external" shutdown signal
	shutdownChan := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			if counter.Load() > 2 {
				close(shutdownChan)
				return
			}
		}
	}()

	// create ticker and wait for external shutdown
	ticker := NewTicker(func() {
		counter.Inc()
		time.Sleep(1 * time.Second)
		counter.Inc()
	}, 100*time.Millisecond, shutdownChan)

	// wait for the shutdown signal
	ticker.WaitForShutdown()

	// make sure we really waited for the external shutdown signal
	assert.GreaterOrEqual(t, counter.Load(), uint64(3))

	// wait for the handler to finish
	ticker.WaitForGracefulShutdown()

	// make sure we really waited for the handler to finish
	assert.GreaterOrEqual(t, counter.Load(), uint64(4))
}
