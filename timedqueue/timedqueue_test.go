package timedqueue

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueueElement_MemLeak(t *testing.T) {
	const testCount = 100000

	timedQueue := New()
	memStatsStart := memStats()

	go func() {
		for currentElement := timedQueue.Poll(true); currentElement != nil; currentElement = timedQueue.Poll(true) {
			currentElement.(func())()
		}
	}()

	var executionCounter uint64
	for i := 0; i < testCount; i++ {
		timedQueue.Add(func() {
			atomic.AddUint64(&executionCounter, 1)
		}, time.Now().Add(500*time.Millisecond))
	}

	assert.Eventually(t, func() bool {
		return atomic.LoadUint64(&executionCounter) == uint64(testCount)
	}, 5*time.Second, 10*time.Millisecond)

	time.Sleep(5 * time.Second)

	memStatsEnd := memStats()
	assert.Less(t, float64(memStatsEnd.HeapObjects), 1.1*float64(memStatsStart.HeapObjects), "the objects in the heap should not grow by more than 10%")
}

func memStats() *runtime.MemStats {
	runtime.GC()
	debug.FreeOSMemory()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &memStats
}
