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
	}, 10*time.Second, 10*time.Millisecond)

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
