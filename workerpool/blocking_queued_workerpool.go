package workerpool

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
)

// BlockingQueuedWorkerPool represents a set of workers with a blocking queue of pending tasks.
type BlockingQueuedWorkerPool struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	options *Options

	calls chan func()

	running  bool
	shutdown bool

	mutex syncutils.RWMutex
	wait  sync.WaitGroup
}

// NewBlockingQueuedWorkerPool returns a new stopped WorkerPool.
func NewBlockingQueuedWorkerPool(optionalOptions ...Option) (result *BlockingQueuedWorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &BlockingQueuedWorkerPool{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		options:   options,
		calls:     make(chan func(), options.QueueSize),
	}

	return
}

// Submit submits a task to the loop, if the queue is full the call blocks until the task is succesfully submitted.
func (wp *BlockingQueuedWorkerPool) Submit(f func()) {
	wp.mutex.RLock()

	if wp.shutdown {
		wp.mutex.RUnlock()
		return
	}
	wp.mutex.RUnlock()

	wp.calls <- f
}

// TrySubmit submits a task to the loop without blocking, it returns false if the queue is full and the task was not
// succesfully submitted.
func (wp *BlockingQueuedWorkerPool) TrySubmit(f func()) (added bool) {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.shutdown {
		select {
		case wp.calls <- f:
			return true
		default:
			return false
		}
	}

	return false
}

// Start starts the WorkerPool (non-blocking).
func (wp *BlockingQueuedWorkerPool) Start() {
	wp.mutex.Lock()

	if !wp.running {
		if wp.shutdown {
			panic("Worker was already used before")
		}
		wp.running = true

		wp.startWorkers()
	}

	wp.mutex.Unlock()
}

// Run starts the WorkerPool and waits for its shutdown.
func (wp *BlockingQueuedWorkerPool) Run() {
	wp.Start()

	wp.wait.Wait()
}

// Stop stops the WorkerPool.
func (wp *BlockingQueuedWorkerPool) Stop() {
	wp.mutex.Lock()

	if wp.running {
		wp.shutdown = true
		wp.running = false

		wp.ctxCancel()
	}

	wp.mutex.Unlock()
}

// StopAndWait stops the WorkerPool and waits for its shutdown.
func (wp *BlockingQueuedWorkerPool) StopAndWait() {
	wp.Stop()
	wp.wait.Wait()
}

// GetWorkerCount returns the worker count for the WorkerPool.
func (wp *BlockingQueuedWorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}

// GetPendingQueueSize returns the amount of tasks pending to the processed.
func (wp *BlockingQueuedWorkerPool) GetPendingQueueSize() int {
	return len(wp.calls)
}

func (wp *BlockingQueuedWorkerPool) startWorkers() {
	for i := 0; i < wp.options.WorkerCount; i++ {
		wp.wait.Add(1)

		go func() {
			aborted := false

			for !aborted {
				select {

				case <-wp.ctx.Done():
					aborted = true

					if wp.options.FlushTasksAtShutdown {
					terminateLoop:
						// process all waiting tasks after shutdown signal
						for {
							select {
							case f := <-wp.calls:
								f()

							default:
								break terminateLoop
							}
						}
					}
				case f := <-wp.calls:
					f()
				}
			}

			wp.wait.Done()
		}()

	}
}
