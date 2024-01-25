package workerpool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebounceFunc(t *testing.T) {
	wp := New(t.Name()).Start()
	debounce := wp.DebounceFunc()

	var latestValue atomic.Uint64
	var callCount atomic.Uint64

	for i := 0; i < 100; i++ {
		iCaptured := i

		debounce(func() {
			callCount.Add(1)

			latestValue.Store(uint64(iCaptured))

			time.Sleep(1 * time.Second)
		})
	}

	wp.PendingTasksCounter.WaitIsZero()

	require.Less(t, callCount.Load(), uint64(3))
	require.Equal(t, uint64(99), latestValue.Load())
}

func Test_NonBlockingNoFlush(t *testing.T) {
	const workerCount = 2

	wp := New(t.Name(), WithWorkerCount(workerCount), WithCancelPendingTasksOnShutdown(true))

	wp.Start()

	var atomicCounter atomic.Int64

	slowFunc := func() {
		atomicCounter.Add(1)
		time.Sleep(50 * time.Millisecond)
	}

	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	// These ones should never be executed
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	assert.Eventually(t, func() bool {
		return atomicCounter.Load() >= 2
	}, 1*time.Second, 1*time.Millisecond)

	assert.LessOrEqual(t, atomicCounter.Load(), int64(12))
}

func Test_NonBlockingFlush(t *testing.T) {
	const workerCount = 2

	wp := New(t.Name(), WithWorkerCount(workerCount))

	wp.Start()

	var atomicCounter atomic.Int64

	slowFunc := func() {
		atomicCounter.Add(1)
		time.Sleep(50 * time.Millisecond)
	}

	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	assert.Eventually(t, func() bool {
		return atomicCounter.Load() >= 2
	}, 10*time.Second, 1*time.Millisecond)

	wp.Shutdown().ShutdownComplete.Wait()

	assert.EqualValues(t, atomicCounter.Load(), 8)
}

func Test_QueueWaitSizeIsBelow(t *testing.T) {
	const workerCount = 2

	wp := New(t.Name(), WithWorkerCount(workerCount), WithCancelPendingTasksOnShutdown(true))

	wp.Start()

	var atomicCounter atomic.Int64

	slowFunc := func() {
		time.Sleep(50 * time.Millisecond)
		atomicCounter.Add(1)
	}

	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	// This one should never be executed
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)
	wp.Submit(slowFunc)

	wp.Queue.WaitSizeIsBelow(4)

	assert.LessOrEqual(t, wp.Queue.Size(), 4)

	wp.Shutdown()
}

func Test_EmptyPoolStartupAndShutdown(t *testing.T) {
	const workerCount = 2

	wp := New(t.Name(), WithWorkerCount(workerCount))

	wp.Start()

	time.Sleep(50 * time.Millisecond)

	wp.Shutdown()
}
