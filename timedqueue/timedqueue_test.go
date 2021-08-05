package timedqueue

import (
	"container/heap"
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
	}, 10*time.Second, 100*time.Millisecond)

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

func TestQueueSize(t *testing.T) {
	const maxSize = 100
	t1 := time.Now().Add(1 * time.Hour)
	t2 := time.Now().Add(5 * time.Hour)

	timedQueue := New(WithMaxSize(maxSize))
	defer timedQueue.Shutdown()

	// start worker (will simply block because times are too far in the future)
	go func() {
		for currentElement := timedQueue.Poll(true); currentElement != nil; currentElement = timedQueue.Poll(true) {
			currentElement.(func())()
		}
	}()

	// add element at t2 (far in the future) to the queue
	timedQueue.Add(func() {}, t2)

	// now fill up queue
	for i := 0; i < maxSize; i++ {
		timedQueue.Add(func() {}, t1)
	}

	// verify that only maxSize elements are in queue
	assert.Equal(t, maxSize, timedQueue.Size())

	// verify that all elements in the queue have time t1
	for i := 0; i < timedQueue.Size(); i++ {
		e := heap.Remove(&timedQueue.heap, 0).(*QueueElement)
		assert.Equal(t, t1, e.ScheduledTime)
	}
}
