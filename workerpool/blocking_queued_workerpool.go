package workerpool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/types"
)

// region BlockQueuedWorkerPool ////////////////////////////////////////////////////////////////////////////////////////

// BlockingQueuedWorkerPool represents a set of workers with a blocking queue of pending tasks.
type BlockingQueuedWorkerPool struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	options *Options

	calls chan func()

	running  bool
	shutdown bool

	pendingTasksCounter uint64
	emptyChan           chan types.Empty
	emptyChanMutex      sync.RWMutex

	mutex syncutils.RWMutex
	wait  sync.WaitGroup
}

// NewBlockingQueuedWorkerPool returns a new stopped WorkerPool.
func NewBlockingQueuedWorkerPool(optionalOptions ...Option) (result *BlockingQueuedWorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &BlockingQueuedWorkerPool{
		emptyChan: make(chan types.Empty),
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

	atomic.AddUint64(&wp.pendingTasksCounter, 1)
	wp.calls <- f
}

// TrySubmit submits a task to the loop without blocking, it returns false if the queue is full and the task was not
// succesfully submitted.
func (wp *BlockingQueuedWorkerPool) TrySubmit(f func()) (added bool) {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.shutdown {
		atomic.AddUint64(&wp.pendingTasksCounter, 1)
		select {
		case wp.calls <- f:
			return true
		default:
			wp.decreasePendingTasksCounter()
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
								wp.decreasePendingTasksCounter()

							default:
								break terminateLoop
							}
						}
					}
				case f := <-wp.calls:
					f()
					wp.decreasePendingTasksCounter()
				}
			}

			wp.wait.Done()
		}()

	}
}

// WaitUntilAllTasksProcessed waits until all tasks are processed.
func (wp *BlockingQueuedWorkerPool) WaitUntilAllTasksProcessed() {
	if atomic.LoadUint64(&wp.pendingTasksCounter) == 0 {
		return
	}

	wp.emptyChanMutex.RLock()
	emptyChan := wp.emptyChan
	wp.emptyChanMutex.RUnlock()

	<-emptyChan
}

// decreasePendingTasksCounter decreases the pending tasks counter and closes the empty channel if necessary.
func (wp *BlockingQueuedWorkerPool) decreasePendingTasksCounter() {
	if atomic.AddUint64(&wp.pendingTasksCounter, ^uint64(0)) == 0 {
		wp.emptyChanMutex.Lock()
		defer wp.emptyChanMutex.Unlock()

		close(wp.emptyChan)
		wp.emptyChan = make(chan types.Empty)
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
