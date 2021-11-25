package async

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/iotaledger/hive.go/v2/typeutils"
)

// Implements a blocking goroutine pool with fixed capacity, managing and recycling a massive number of goroutines,
// allowing developers to limit the number of goroutines in your concurrent programs.
type WorkerPool struct {
	pool         *ants.Pool
	stopped      typeutils.AtomicBool
	tasksWg      sync.WaitGroup
	initOnce     sync.Once
	shutdownOnce sync.Once
}

// Submits a task to this pool (it waits if not enough workers are available).
func (workerPool *WorkerPool) Submit(f func()) {
	if workerPool.stopped.IsSet() {
		return
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
	}); antsErr != nil {
		panic(antsErr)
	}
}

// Returns the capacity (number of workers) of this pool.
func (workerPool *WorkerPool) Capacity() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Cap()
}

// Changes the capacity of this pool.
func (workerPool *WorkerPool) Tune(capacity int) *WorkerPool {
	if workerPool.stopped.IsSet() {
		return workerPool
	}

	workerPool.get().Tune(capacity)

	return workerPool
}

// Returns the available (idle) workers.
func (workerPool *WorkerPool) IdleWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Free()
}

// Returns the number of the currently running workers.
func (workerPool *WorkerPool) RunningWorkers() int {
	if workerPool.stopped.IsSet() {
		return 0
	}

	return workerPool.get().Running()
}

// Immediately closes this pool (see ShutdownGracefully for a method that waits for the scheduled workers to finish).
func (workerPool *WorkerPool) Shutdown() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Closes this pool and waits for the currently running goroutines to finish.
func (workerPool *WorkerPool) ShutdownGracefully() {
	workerPool.shutdownOnce.Do(func() {
		workerPool.stopped.Set()
		workerPool.tasksWg.Wait()

		if workerPool.pool != nil {
			go workerPool.pool.Release()
		}
	})
}

// Internal utility function that initiates and returns the pool.
func (workerPool *WorkerPool) get() *ants.Pool {
	workerPool.initOnce.Do(func() {
		if newPool, err := ants.NewPool(runtime.GOMAXPROCS(0), ants.WithNonblocking(false)); err != nil {
			panic(err)
		} else {
			workerPool.pool = newPool
		}
	})

	return workerPool.pool
}
