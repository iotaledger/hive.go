package timedexecutor

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimedExecutor_MemLeak(t *testing.T) {
	const testCount = 100000

	timedExecutor := New(1)
	memStatsStart := memStats()

	var executionCounter uint64
	for i := 0; i < testCount; i++ {
		timedExecutor.ExecuteAfter(func() {
			atomic.AddUint64(&executionCounter, 1)
		}, 500*time.Millisecond)
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

func TestTimedExecutor(t *testing.T) {
	const workerCount = 4

	timedExecutor := New(workerCount)
	defer timedExecutor.Shutdown()
	assert.Equal(t, workerCount, timedExecutor.WorkerCount())

	// prepare a list of functions to schedule
	elements := make(map[time.Time]func())
	var expected, actual []int
	now := time.Now().Add(15 * time.Second)

	for i := 0; i < 10; i++ {
		str := i
		elements[now.Add(time.Duration(i)*time.Second)] = func() {
			actual = append(actual, str)
		}
		expected = append(expected, i)
	}

	// insert functions to timedExecutor
	for t, f := range elements {
		timedExecutor.ExecuteAt(f, t)
	}
	assert.Equal(t, len(elements), timedExecutor.Size())

	assert.Eventually(t, func() bool { return len(actual) == len(expected) }, 5*time.Minute, 100*time.Millisecond)
	assert.Equal(t, 0, timedExecutor.Size())
	assert.ElementsMatch(t, expected, actual)
}
