package workerpool

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/debug"
	"github.com/iotaledger/hive.go/core/syncutils"
	"github.com/iotaledger/hive.go/core/types"
)

// region BlockingQueuedWorkerPool /////////////////////////////////////////////////////////////////////////////////////

// BlockingQueuedWorkerPool represents a set of workers with a blocking queue of pending tasks.
type BlockingQueuedWorkerPool struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	options   *Options
	tasks     chan *BlockingQueueWorkerPoolTask
	running   bool
	shutdown  bool

	pendingTasksCounter        uint64
	pendingTasksMutex          sync.Mutex
	waitUntilAllTasksProcessed *sync.Cond
	mutex                      syncutils.RWMutex
	workers                    sync.WaitGroup
}

// NewBlockingQueuedWorkerPool returns a new stopped WorkerPool.
func NewBlockingQueuedWorkerPool(optionalOptions ...Option) (result *BlockingQueuedWorkerPool) {
	options := defaultOptions.Override(optionalOptions...)
	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &BlockingQueuedWorkerPool{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		options:   options,
		tasks:     make(chan *BlockingQueueWorkerPoolTask, options.QueueSize),
	}
	result.waitUntilAllTasksProcessed = sync.NewCond(&result.pendingTasksMutex)

	return result
}

// Submit submits a handler function to the queue and blocks if the queue is full.
func (b *BlockingQueuedWorkerPool) Submit(handler func()) {
	b.SubmitTask(b.CreateTask(handler))
}

// TrySubmit tries to queue the execution of the handler function and ignores the handler if there is no capacity for it
// to be added.
func (b *BlockingQueuedWorkerPool) TrySubmit(f func()) (added bool) {
	return b.TrySubmitTask(b.CreateTask(f))
}

// CreateTask creates a new BlockingQueueWorkerPoolTask with the given handler and optional ClosureStackTrace.
func (b *BlockingQueuedWorkerPool) CreateTask(f func(), optionalStackTrace ...string) *BlockingQueueWorkerPoolTask {
	b.IncreasePendingTasksCounter()

	var stackTrace string
	if len(optionalStackTrace) > 1 {
		stackTrace = optionalStackTrace[0]
	}

	return newBlockingQueueWorkerPoolTask(b, f, stackTrace)
}

// SubmitTask submits a task to the queue and blocks if the queue is full (it should only be used instead of Submit if
// manually handling the task is necessary to create better debug outputs).
func (b *BlockingQueuedWorkerPool) SubmitTask(task *BlockingQueueWorkerPoolTask) {
	if !b.IsRunning() {
		task.markDone()

		return
	}
	b.tasks <- task
}

// TrySubmitTask tries to queue the execution of the task and ignores the task if there is no capacity for it to
// be added (it should only be used instead of TrySubmit if manually handling the task is necessary to create better
// debug outputs).
func (b *BlockingQueuedWorkerPool) TrySubmitTask(task *BlockingQueueWorkerPoolTask) (added bool) {
	if !b.IsRunning() {
		task.markDone()

		return false
	}

	select {
	case b.tasks <- task:
		return true
	default:
		task.markDone()

		return false
	}
}

// Start starts the WorkerPool (non-blocking).
func (b *BlockingQueuedWorkerPool) Start() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.running {
		if b.shutdown {
			panic("Worker was already used before")
		}
		b.running = true

		b.startWorkers()
	}
}

// Run starts the WorkerPool and waits for its shutdown.
func (b *BlockingQueuedWorkerPool) Run() {
	b.Start()

	b.workers.Wait()
}

// Stop stops the WorkerPool.
func (b *BlockingQueuedWorkerPool) Stop() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.running {
		b.shutdown = true
		b.running = false

		b.ctxCancel()
	}
}

// StopAndWait stops the WorkerPool and waits for its shutdown.
func (b *BlockingQueuedWorkerPool) StopAndWait() {
	b.Stop()
	b.workers.Wait()
}

// GetWorkerCount returns the worker count for the WorkerPool.
func (b *BlockingQueuedWorkerPool) GetWorkerCount() int {
	return b.options.WorkerCount
}

// GetPendingQueueSize returns the amount of tasks pending to the processed.
func (b *BlockingQueuedWorkerPool) GetPendingQueueSize() int {
	return len(b.tasks)
}

// WaitUntilAllTasksProcessed waits until all tasks are processed.
func (b *BlockingQueuedWorkerPool) WaitUntilAllTasksProcessed() {
	b.pendingTasksMutex.Lock()
	defer b.pendingTasksMutex.Unlock()
	if b.pendingTasksCounter == 0 {
		return
	}

	for b.pendingTasksCounter > 0 {
		b.waitUntilAllTasksProcessed.Wait()
	}
}

// IsRunning returns true if the WorkerPool is running.
func (b *BlockingQueuedWorkerPool) IsRunning() (isRunning bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return !b.shutdown
}

func (b *BlockingQueuedWorkerPool) alias() string {
	if b.options.Alias != "" {
		return b.options.Alias
	}

	return "BlockingQueuedWorkerPool"
}

func (b *BlockingQueuedWorkerPool) startWorkers() {
	for i := 0; i < b.options.WorkerCount; i++ {
		b.workers.Add(1)

		go func() {
			for aborted := false; !aborted; {
				select {
				case <-b.ctx.Done():
					aborted = true

					if b.options.FlushTasksAtShutdown {
					terminateLoop:
						// process all waiting tasks after shutdown signal
						for {
							select {
							case currentTask := <-b.tasks:
								currentTask.run()
							default:
								break terminateLoop
							}
						}
					}
				case currentTask := <-b.tasks:
					currentTask.run()
				}
			}

			b.workers.Done()
		}()
	}
}

// IncreasePendingTasksCounter increases the pending task counter.
func (b *BlockingQueuedWorkerPool) IncreasePendingTasksCounter() {
	b.pendingTasksMutex.Lock()
	defer b.pendingTasksMutex.Unlock()

	b.pendingTasksCounter++
}

// DecreasePendingTasksCounter decreases the pending task counter.
func (b *BlockingQueuedWorkerPool) DecreasePendingTasksCounter() {
	b.pendingTasksMutex.Lock()

	b.pendingTasksCounter--
	if b.pendingTasksCounter != 0 {
		b.pendingTasksMutex.Unlock()

		return
	}
	b.pendingTasksMutex.Unlock()
	b.waitUntilAllTasksProcessed.Broadcast()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region BlockingQueueWorkerPoolTask /////////////////////////////////////////////////////////////////////////////////

// BlockingQueueWorkerPoolTask is a task that is executed by a BlockingQueuedWorkerPool.
type BlockingQueueWorkerPoolTask struct {
	workerPool *BlockingQueuedWorkerPool
	workerFunc func()
	doneChan   chan types.Empty
	stackTrace string
}

// newBlockingQueueWorkerPoolTask creates a new BlockingQueueWorkerPoolTask.
func newBlockingQueueWorkerPoolTask(workerPool *BlockingQueuedWorkerPool, workerFunc func(), stackTrace string) *BlockingQueueWorkerPoolTask {
	if debug.GetEnabled() && stackTrace == "" {
		stackTrace = debug.ClosureStackTrace(workerFunc)
	}

	return &BlockingQueueWorkerPoolTask{
		workerPool: workerPool,
		workerFunc: workerFunc,
		doneChan:   make(chan types.Empty),
		stackTrace: stackTrace,
	}
}

// run executes the task.
func (t *BlockingQueueWorkerPoolTask) run() {
	if debug.GetEnabled() {
		go t.detectedHangingTasks()
	}

	t.workerFunc()
	t.markDone()
}

// markDone marks the task as done.
func (t *BlockingQueueWorkerPoolTask) markDone() {
	close(t.doneChan)
	t.workerPool.DecreasePendingTasksCounter()
}

// detectedHangingTasks is a debug method that is used to print information about possibly hanging task executions.
func (t *BlockingQueueWorkerPoolTask) detectedHangingTasks() {
	select {
	case <-t.doneChan:
		return
	case <-time.After(debug.DeadlockDetectionTimeout):
		fmt.Println("task in " + t.workerPool.alias() + " seems to hang (" + debug.DeadlockDetectionTimeout.String() + ") ...")
		fmt.Println("\n" + strings.Replace(strings.Replace(t.stackTrace, "closure:", "task:", 1), "called by", "queued by", 1))
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
