package async

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/iotaledger/hive.go/datastructure/queue"
	"github.com/iotaledger/hive.go/typeutils"
)

// Implements a non-blocking (it enqueues the tasks in a queue of fixed capacity if the pool is full) goroutine pool
// with fixed capacity, managing and recycling a massive number of goroutines, allowing developers to limit the number
// of goroutines in your concurrent programs.
type NonBlockingQueueWorkerPool struct {
	pool         *ants.Pool
	queue        *queue.Queue
	stopped      typeutils.AtomicBool
	tasksWg      sync.WaitGroup
	initOnce     sync.Once
	shutdownOnce sync.Once
}

// Submits a task to this pool (it drops the task if not enough workers are available).
// It returns false if the task could not be submitted because the pool is full.
func (workerPool *NonBlockingQueueWorkerPool) Submit(f func()) bool {
	if workerPool.stopped.IsSet() {
		return false
	}

	workerPool.tasksWg.Add(1)

	if antsErr := workerPool.get().Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("recovered from panic in WorkerPool: %s %s", r, debug.Stack())
			}

			workerPool.tasksWg.Done()

			if queuedTask, someInQueue := workerPool.queue.Poll(); someInQueue {
				workerPool.Submit(queuedTask.(func()))
			}
		}()

		f()
	}); antsErr != nil && antsErr != ants.ErrPoolOverload {
		panic(antsErr)
	} else {
		if antsErr == nil {
			return true
		} else {
			workerPool.tasksWg.Done()

			return workerPool.queue.Offer(f)
		}
	}
}

// Returns the capacity (number of workers) of this pool.
func (workerPool *NonBlockingQueueWorkerPool) Capacity() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Cap()
}

// Changes the capacity of this pool.
func (workerPool *NonBlockingQueueWorkerPool) Tune(capacity int) *NonBlockingQueueWorkerPool {
	if workerPool.stopped.IsSet() {
		return workerPool
	}

	// Queue's capacity is not altered
	workerPool.get().Tune(capacity)

	return workerPool
}

// Returns the available (idle) workers.
func (workerPool *NonBlockingQueueWorkerPool) IdleWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Free()
}

// Returns the number of the currently running workers.
func (workerPool *NonBlockingQueueWorkerPool) RunningWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Running()
}

// Immediately closes this pool (see ShutdownGracefully for a method that waits for the running workers to finish).
func (workerPool *NonBlockingQueueWorkerPool) Shutdown() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Closes this pool and waits for the currently running goroutines to finish.
func (workerPool *NonBlockingQueueWorkerPool) ShutdownGracefully() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()
		workerPool.tasksWg.Wait()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Internal utility function that initiates and returns the pool.
func (workerPool *NonBlockingQueueWorkerPool) get() *ants.Pool {
	workerPool.initOnce.Do(func() {
		if newPool, err := ants.NewPool(runtime.GOMAXPROCS(0), ants.WithNonblocking(true)); err != nil {
			panic(err)
		} else {
			workerPool.pool = newPool
			workerPool.queue = queue.New(runtime.GOMAXPROCS(0) * 2)
		}
	})

	return workerPool.pool
}
