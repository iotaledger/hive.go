package workerpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/debug"
)

func Test_SimpleCounter(t *testing.T) {
	const queueSize = 10
	const incCount = 100

	el := NewBlockingQueuedWorkerPool(QueueSize(queueSize), FlushTasksAtShutdown(true))

	var counter uint64
	incAtomic := func() {
		atomic.AddUint64(&counter, 1)
	}

	for i := 0; i < incCount; i++ {
		added := el.TrySubmit(incAtomic)

		if i < queueSize {
			assert.True(t, added)
		} else {
			assert.False(t, added)
		}
	}
	assert.Equal(t, queueSize, el.GetPendingQueueSize())
	el.Start()

	for i := 0; i < incCount-queueSize; i++ {
		el.Submit(incAtomic)
	}
	el.StopAndWait()
	assert.Equal(t, uint64(incCount), counter)
}

func Test_ShutdownNotAdded(t *testing.T) {
	el := NewBlockingQueuedWorkerPool()
	el.Start()
	el.Stop()

	counter := 0
	inc := func() {
		counter++
	}

	el.Submit(inc)
	assert.False(t, el.TrySubmit(inc))

	assert.Equal(t, 0, counter)
}

func Test_NoFlush(t *testing.T) {
	const workerCount = 1
	const incCount = 100

	elNoFlush := NewBlockingQueuedWorkerPool(WorkerCount(workerCount), QueueSize(incCount), FlushTasksAtShutdown(false))
	elNoFlush.Start()

	assert.Equal(t, 1, elNoFlush.GetWorkerCount())

	var noFlushCounter uint64
	slowIncAtomic := func(counter *uint64) func() {
		return func() {
			atomic.AddUint64(counter, 1)
			time.Sleep(5 * time.Millisecond)
		}
	}

	for i := 0; i < incCount; i++ {
		elNoFlush.Submit(slowIncAtomic(&noFlushCounter))
	}

	elNoFlush.StopAndWait()

	assert.NotEqual(t, uint64(incCount), noFlushCounter)
}

func Test_NoFlushVsFlush(t *testing.T) {
	const workerCount = 1
	const incCount = 100

	elNoFlush := NewBlockingQueuedWorkerPool(WorkerCount(workerCount), QueueSize(incCount), FlushTasksAtShutdown(false))
	elFlush := NewBlockingQueuedWorkerPool(WorkerCount(workerCount), QueueSize(incCount), FlushTasksAtShutdown(true))
	elNoFlush.Start()
	elFlush.Start()

	assert.Equal(t, 1, elNoFlush.GetWorkerCount())
	assert.Equal(t, 1, elFlush.GetWorkerCount())

	var noFlushCounter uint64
	var flushCounter uint64
	slowIncAtomic := func(counter *uint64) func() {
		return func() {
			atomic.AddUint64(counter, 1)
			time.Sleep(5 * time.Millisecond)
		}
	}

	for i := 0; i < incCount; i++ {
		elNoFlush.Submit(slowIncAtomic(&noFlushCounter))
		elFlush.Submit(slowIncAtomic(&flushCounter))
	}

	elNoFlush.Stop()
	elFlush.Stop()

	elNoFlush.StopAndWait()
	elFlush.StopAndWait()

	assert.NotEqual(t, uint64(incCount), noFlushCounter)
	assert.Equal(t, uint64(incCount), flushCounter)
}

func Test_DeadlockDetection(t *testing.T) {
	debug.SetEnabled(true)

	el := NewBlockingQueuedWorkerPool(QueueSize(1000000), FlushTasksAtShutdown(true), WithAlias("event.Loop"))
	el.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	el.Submit(func() {
		time.Sleep(7 * time.Second)

		fmt.Println("DONE")

		wg.Done()
	})

	wg.Wait()
}

func Test_Events(t *testing.T) {
	el := NewBlockingQueuedWorkerPool(QueueSize(1000000), FlushTasksAtShutdown(true))
	el.Start()

	var counter uint64
	el.TrySubmit(func() {
		time.Sleep(500 * time.Millisecond)
		atomic.AddUint64(&counter, 1)
	})

	total := 10000
	for i := 0; i < total; i++ {
		el.Submit(func() {
			el.Submit(func() {
				atomic.AddUint64(&counter, 1)

				el.Submit(func() {
					atomic.AddUint64(&counter, 1)
				})
			})
		})
	}
	el.WaitUntilAllTasksProcessed()

	assert.EqualValues(t, total*2+1, counter)
}
