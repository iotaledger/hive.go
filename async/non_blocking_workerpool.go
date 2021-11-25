package async

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/iotaledger/hive.go/v2/typeutils"
)

// Implements a non-blocking (it drops tasks if the pool is full) goroutine pool with fixed capacity, managing and
// recycling a massive number of goroutines, allowing developers to limit the number of goroutines in your concurrent
// programs.
type NonBlockingWorkerPool struct {
	pool         *ants.Pool
	stopped      typeutils.AtomicBool
	tasksWg      sync.WaitGroup
	initOnce     sync.Once
	shutdownOnce sync.Once
}

// Submits a task to this pool (it drops the task if not enough workers are available).
// It returns false if the task could not be submitted because the pool is full.
func (workerPool *NonBlockingWorkerPool) Submit(f func()) bool {
	if workerPool.stopped.IsSet() {
		return false
	}

	workerPool.tasksWg.Add(1)

	if antsErr := workerPool.get().Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				workerPool.tasksWg.Done()

				fmt.Printf("recovered from panic in WorkerPool: %s %s", r, debug.Stack())
			}
		}()

		f()

		workerPool.tasksWg.Done()
	}); antsErr != nil && antsErr != ants.ErrPoolOverload {
		panic(antsErr)
	} else {
		if antsErr == nil {
			return true
		} else {
			workerPool.tasksWg.Done()

			return false
		}
	}
}

// Returns the capacity (number of workers) of this pool.
func (workerPool *NonBlockingWorkerPool) Capacity() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Cap()
}

// Changes the capacity of this pool.
func (workerPool *NonBlockingWorkerPool) Tune(capacity int) *NonBlockingWorkerPool {
	if workerPool.stopped.IsSet() {
		return workerPool
	}

	workerPool.get().Tune(capacity)

	return workerPool
}

// Returns the available (idle) workers.
func (workerPool *NonBlockingWorkerPool) IdleWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Free()
}

// Returns the number of the currently running workers.
func (workerPool *NonBlockingWorkerPool) RunningWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Running()
}

// Immediately closes this pool (see ShutdownGracefully for a method that waits for the running workers to finish).
func (workerPool *NonBlockingWorkerPool) Shutdown() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Closes this pool and waits for the currently running goroutines to finish.
func (workerPool *NonBlockingWorkerPool) ShutdownGracefully() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()
		workerPool.tasksWg.Wait()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Internal utility function that initiates and returns the pool.
func (workerPool *NonBlockingWorkerPool) get() *ants.Pool {
	workerPool.initOnce.Do(func() {
		if newPool, err := ants.NewPool(runtime.GOMAXPROCS(0), ants.WithNonblocking(true)); err != nil {
			panic(err)
		} else {
			workerPool.pool = newPool
		}
	})

	return workerPool.pool
}
