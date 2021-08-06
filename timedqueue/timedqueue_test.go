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

func TestTimedQueue(t *testing.T) {
	timedQ := New()
	popCount := 0

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(15 * time.Second)
	for i := 0; i < 10; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
	}

	// insert elements to timedQs
	for i := 0; i < 10; i++ {
		timedQ.Add(i, elements[i])
	}
	assert.Equal(t, len(elements), timedQ.Size())

	// wait and Poll all elements, check the popped time is correct.
	for {
		topElement := timedQ.Poll(false).(int)
		popTime := time.Now()
		assert.True(t, (popTime.Sub(elements[topElement]) < time.Duration(1*time.Millisecond)))
		popCount++

		if timedQ.Size() == 0 {
			break
		}
	}

	assert.Equal(t, 0, timedQ.Size())
	assert.Equal(t, len(elements), popCount)
	timedQ.Shutdown()
}

func TestTimedQueuePollWaitIfEmpty(t *testing.T) {
	timedQ := New()
	end := make(chan struct{})
	popCount := 0

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(15 * time.Second)
	for i := 0; i < 10; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
	}

	go func() {
		for {
			topElement := timedQ.Poll(true).(int)
			popTime := time.Now()
			assert.True(t, (popTime.Sub(elements[topElement]) < time.Duration(1*time.Millisecond)))
			popCount++

			if timedQ.Size() == 0 {
				close(end)
				break
			}
		}
	}()

	// let timedQ reader to wait for a second
	time.Sleep(time.Duration(1 * time.Second))

	// insert elements to timedQ
	for i := 0; i < 10; i++ {
		timedQ.Add(i, elements[i])
	}
	assert.Equal(t, len(elements), timedQ.Size())

	// wait all element is clear
	<-end
	assert.Equal(t, 0, timedQ.Size())
	assert.Equal(t, len(elements), popCount)
	timedQ.Shutdown()
}

func TestTimedQueueCancel(t *testing.T) {
	timedQ := New()
	end := make(chan struct{})
	popCount := 0

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(15 * time.Second)
	for i := 0; i < 10; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
	}

	go func() {
		for {
			topElement := timedQ.Poll(false).(int)
			popTime := time.Now()
			assert.True(t, (popTime.Sub(elements[topElement]) < time.Duration(1*time.Millisecond)))
			popCount++

			if timedQ.Size() == 0 {
				close(end)
				break
			}
		}
	}()

	// insert elements to timedQ
	var timeElements []*QueueElement
	for i := 0; i < 10; i++ {
		timeElements = append(timeElements, timedQ.Add(i, elements[i]))
	}
	assert.Equal(t, len(elements), timedQ.Size())

	// cancel the first and the last element
	timeElements[0].Cancel()
	timeElements[len(elements)-1].Cancel()

	// wait all element is clear
	<-end
	assert.Equal(t, 0, timedQ.Size())
	assert.Equal(t, len(elements)-2, popCount)
	timedQ.Shutdown()
}
