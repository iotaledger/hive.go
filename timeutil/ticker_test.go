package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func TestTicker_WaitForShutdown(t *testing.T) {
	counter := atomic.NewUint64(0)

	shutdownChan := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if counter.Load() > 10 {
				close(shutdownChan)
				return
			}
		}
	}()

	ticker := NewTicker(func() { counter.Inc() }, 100*time.Millisecond, shutdownChan)
	ticker.WaitForShutdown()

	assert.GreaterOrEqual(t, counter.Load(), uint64(10))
}
