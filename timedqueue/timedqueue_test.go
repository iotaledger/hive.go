package timedqueue

import (
	"runtime"
	"runtime/debug"
	"sync"
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
	const elementsCount = 10

	tq := New()
	defer tq.Shutdown()

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(5 * time.Second)

	for i := 0; i < elementsCount; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
		tq.Add(i, elements[i])
	}

	assert.Equal(t, len(elements), tq.Size())

	// wait and Poll all elements, check the popped time is correct.
	consumed := 0
	for {
		topElement := tq.Poll(false).(int)

		// make sure elements are ready at their specified time.
		assert.True(t, time.Since(elements[topElement]) < 200*time.Millisecond)
		consumed++

		if tq.Size() == 0 {
			break
		}
	}

	// check that we consumed all elements
	assert.Equal(t, len(elements), consumed)
}

func TestTimedQueuePollWaitIfEmpty(t *testing.T) {
	const elementsCount = 10

	tq := New()
	defer tq.Shutdown()

	consumed := 0

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(5 * time.Second)
	for i := 0; i < elementsCount; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			topElement := tq.Poll(true).(int)

			// make sure elements are ready at their specified time.
			assert.True(t, time.Since(elements[topElement]) < 200*time.Millisecond)
			consumed++

			if tq.Size() == 0 {
				wg.Done()
				break
			}
		}
	}()

	// let worker wait for a second
	time.Sleep(1 * time.Second)

	// insert elements to tq
	for i := 0; i < 10; i++ {
		tq.Add(i, elements[i])
	}
	assert.Equal(t, len(elements), tq.Size())

	// wait all element is clear
	wg.Wait()
	assert.Equal(t, 0, tq.Size())
	assert.Equal(t, len(elements), consumed)
}

func TestTimedQueueCancel(t *testing.T) {
	const elementsCount = 10

	tq := New()
	defer tq.Shutdown()

	consumed := 0

	// prepare a list to insert
	var elements []time.Time
	var queueElements []*QueueElement

	now := time.Now().Add(5 * time.Second)
	for i := 0; i < elementsCount; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
		queueElements = append(queueElements, tq.Add(i, elements[i]))
	}
	assert.Equal(t, len(elements), tq.Size())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			topElement := tq.Poll(true).(int)

			// make sure elements are ready at their specified time.
			assert.True(t, time.Since(elements[topElement]) < 200*time.Millisecond)
			consumed++

			if tq.Size() == 0 {
				wg.Done()
				break
			}
		}
	}()

	// cancel the first and the last element
	queueElements[0].Cancel()
	queueElements[len(elements)-1].Cancel()

	// wait all element is clear
	wg.Wait()
	assert.Equal(t, 0, tq.Size())
	assert.Equal(t, len(elements)-2, consumed)
}
