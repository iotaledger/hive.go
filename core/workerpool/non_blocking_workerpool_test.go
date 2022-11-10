package workerpool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NonBlockingNoFlush(t *testing.T) {
	const workerCount = 2

	wp := NewNonBlockingWorkerPool(FlushTasksAtShutdown(false), WorkerCount(workerCount))

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

	wp.StopAndWait()

	assert.LessOrEqual(t, atomicCounter.Load(), int64(12))
}

func Test_NonBlockingFlush(t *testing.T) {
	const workerCount = 2

	wp := NewNonBlockingWorkerPool(FlushTasksAtShutdown(true), WorkerCount(workerCount))

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
	}, 1*time.Second, 1*time.Millisecond)

	wp.StopAndWait()

	assert.EqualValues(t, atomicCounter.Load(), 8)
}

func Test_WaitForUndispatchedTasksBelowThreshold(t *testing.T) {
	const workerCount = 2

	wp := NewNonBlockingWorkerPool(FlushTasksAtShutdown(false), WorkerCount(workerCount))

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

	wp.WaitForUndispatchedTasksBelowThreshold(4)

	assert.LessOrEqual(t, wp.UndispatchedTaskCount(), 4)

	wp.Stop()
}
