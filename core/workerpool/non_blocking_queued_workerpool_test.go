package workerpool

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func slowFunc(t Task) {
	time.Sleep(50 * time.Millisecond)
	t.Return(true)
}

func fastFunc(t Task) {
	t.Return(true)
}

func panicsFunc(t Task) {
	panic("croak!")
}

func assertParemetersIntByteArr(t Task) {
	test := t.Param(0).(*testing.T)
	_, ok := t.Param(1).(int)
	assert.True(test, ok)
	_, ok = t.Param(2).([]byte)
	assert.True(test, ok)
}

func TestNonBlockingQueue_Basic(t *testing.T) {
	capacity := 1
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))

	assert.Equal(t, wp.GetPendingQueueSize(), 0)

	enqueuedAndExecuted, added := wp.Submit()
	assert.True(t, added)

	assert.True(t, (<-enqueuedAndExecuted).(bool))
}

func TestNonBlockingQueue_HandlesPanic(t *testing.T) {
	capacity := 1
	wp := NewNonBlockingQueuedWorkerPool(panicsFunc, WorkerCount(capacity), QueueSize(capacity*2))

	wontResult, added := wp.Submit()
	assert.True(t, added)

	wp.StopAndWait()

	select {
	case <-wontResult:
		assert.Fail(t, "Task should have been paniced.")
	default:
	}
}

func TestNonBlockingQueue_Params(t *testing.T) {
	capacity := 1
	wp := NewNonBlockingQueuedWorkerPool(assertParemetersIntByteArr, WorkerCount(capacity), QueueSize(capacity*2))

	_, added := wp.Submit(t, 11, []byte{255, 31, 1})
	assert.True(t, added)

	wp.StopAndWait()
}

func TestNonBlockingQueue_Stopped(t *testing.T) {
	capacity := 1
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))
	wp.Stop()

	_, added := wp.Submit()
	assert.False(t, added)
}

func TestNonBlockingQueue_SingleProcessor(t *testing.T) {
	capacity := 1
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))

	assert.Equal(t, wp.GetWorkerCount(), capacity)
	assert.Equal(t, wp.GetPendingQueueSize(), 0)

	_, added := wp.Submit(struct{}{})
	assert.True(t, added)

	enqueuedAndExecuted, added := wp.Submit()
	assert.True(t, added)

	assert.True(t, (<-enqueuedAndExecuted).(bool))
}

func TestNonBlockingQueue_Queued(t *testing.T) {
	capacity := 2
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))

	assert.Equal(t, wp.GetWorkerCount(), capacity)
	assert.Equal(t, wp.GetPendingQueueSize(), 0)

	for i := 0; i < capacity*3-1; i++ {
		_, added := wp.Submit(struct{}{})
		assert.True(t, added)
	}

	enqueuedAndExecuted, added := wp.Submit()
	assert.True(t, added)

	assert.True(t, (<-enqueuedAndExecuted).(bool))
}

func TestNonBlockingQueue_QueueDropping(t *testing.T) {
	capacity := runtime.GOMAXPROCS(0)
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))

	assert.Equal(t, wp.GetWorkerCount(), capacity)
	assert.Equal(t, wp.GetPendingQueueSize(), 0)

	for i := 0; i < capacity; i++ {
		_, added := wp.Submit()
		assert.True(t, added)
	}

	for i := 0; i < capacity*2-1; i++ {
		assert.Equal(t, wp.GetPendingQueueSize(), i)
		_, added := wp.Submit()
		assert.True(t, added)
	}

	enqueuedAndExecuted, addedEnqueued := wp.Submit()
	assert.True(t, addedEnqueued)

	// Reached queue capacity
	_, added := wp.Submit()
	assert.False(t, added)
	_, added = wp.Submit()
	assert.False(t, added)

	wp.tasksWg.Wait()
	if addedEnqueued {
		assert.True(t, (<-enqueuedAndExecuted).(bool))
	}
}

func TestNonBlockingQueue_Concurrency(t *testing.T) {
	goroutines := 10000
	capacity := 10
	numTasks := 100
	wp := NewNonBlockingQueuedWorkerPool(fastFunc, WorkerCount(capacity), QueueSize(capacity*numTasks), FlushTasksAtShutdown(true))
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for k := 0; k < numTasks; k++ {
				wp.Submit()
			}
		}()
	}
	wg.Wait()
	wp.StopAndWait()
}

func TestNonBlockingQueue_NoFlushing(t *testing.T) {
	capacity := 2
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2))

	for i := 0; i < capacity*3-1; i++ {
		_, added := wp.Submit()
		assert.True(t, added)
	}

	wontResult, addedEnqueued := wp.Submit()
	assert.True(t, addedEnqueued)

	// Reached queue capacity
	_, added := wp.Submit()
	assert.False(t, added)
	_, added = wp.Submit()
	assert.False(t, added)

	wp.StopAndWait()

	select {
	case <-wontResult:
		assert.Fail(t, "Task should not have been executed.")
	default:
	}
}

func TestNonBlockingQueue_Flushing(t *testing.T) {
	capacity := 2
	wp := NewNonBlockingQueuedWorkerPool(slowFunc, WorkerCount(capacity), QueueSize(capacity*2), FlushTasksAtShutdown(true))

	for i := 0; i < capacity*3-1; i++ {
		_, added := wp.Submit()
		assert.True(t, added)
	}

	enqueuedAndExecuted, addedEnqueued := wp.Submit()
	assert.True(t, addedEnqueued)

	// Reached queue capacity
	_, added := wp.Submit()
	assert.False(t, added)
	_, added = wp.Submit()
	assert.False(t, added)

	wp.StopAndWait()

	assert.True(t, (<-enqueuedAndExecuted).(bool))
}
