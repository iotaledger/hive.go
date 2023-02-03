package workerpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func Test_NonBlockingNoFlush(t *testing.T) {
	const workerCount = 2

	wp := NewUnboundedWorkerPool(t.Name(), workerCount)

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

	wp.Shutdown(true)

	assert.LessOrEqual(t, atomicCounter.Load(), int64(12))
}

func Test_NonBlockingFlush(t *testing.T) {
	const workerCount = 2

	wp := NewUnboundedWorkerPool(t.Name(), workerCount)

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

	wp := NewUnboundedWorkerPool(t.Name(), workerCount)

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

	wp.Shutdown(true)
}

func Test_EmptyPoolStartupAndShutdown(t *testing.T) {
	const workerCount = 2

	wp := NewUnboundedWorkerPool(t.Name(), workerCount)

	wp.Start()

	time.Sleep(50 * time.Millisecond)

	wp.Shutdown()
}
