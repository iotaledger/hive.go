package workerpool

import (
	"container/list"
	"context"
	"sync"

	"github.com/iotaledger/hive.go/core/syncutils"
)

// region BlockingQueuedWorkerPool /////////////////////////////////////////////////////////////////////////////////////

// BlockingQueuedWorkerPool represents a set of workers with a blocking queue of pending tasks.
type NonBlockingWorkerPool struct {
	ctx                      context.Context
	ctxCancel                context.CancelFunc
	options                  *Options
	nextTaskChan             chan *WorkerPoolTask
	undispatchedTasks        *list.List
	undispatchedTasksMutex   sync.RWMutex
	pendingTasksCounter      uint64
	pendingTasksCounterMutex sync.RWMutex
	allTasksProcessed        *sync.Cond
	undispatchedTaskAdded    *sync.Cond
	undispatchedTaskRemoved  *sync.Cond
	running                  bool
	shutdown                 bool
	mutex                    syncutils.RWMutex
	workers                  sync.WaitGroup
}

// NewNonBlockingWorkerPool returns a new stopped WorkerPool.
func NewNonBlockingWorkerPool(optionalOptions ...Option) (result *NonBlockingWorkerPool) {
	options := defaultOptions.Override(optionalOptions...)
	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &NonBlockingWorkerPool{
		ctx:               ctx,
		ctxCancel:         ctxCancel,
		options:           options,
		nextTaskChan:      make(chan *WorkerPoolTask, options.WorkerCount),
		undispatchedTasks: list.New(),
	}

	result.undispatchedTaskAdded = sync.NewCond(&result.undispatchedTasksMutex)
	result.undispatchedTaskRemoved = sync.NewCond(&result.undispatchedTasksMutex)
	result.allTasksProcessed = sync.NewCond(&result.pendingTasksCounterMutex)

	return result
}

// Submit submits a handler function to the queue and blocks if the queue is full.
func (b *NonBlockingWorkerPool) Submit(handler func()) {
	b.SubmitTask(b.CreateTask(handler))
}

func (b *NonBlockingWorkerPool) WaitForUndispatchedTasksBelowThreshold(threshold int) {
	b.undispatchedTasksMutex.Lock()
	defer b.undispatchedTasksMutex.Unlock()

	for b.undispatchedTasks.Len() > threshold {
		b.undispatchedTaskRemoved.Wait()
	}
}

// CreateTask creates a new BlockingQueueWorkerPoolTask with the given handler and optional ClosureStackTrace.
func (b *NonBlockingWorkerPool) CreateTask(f func(), optionalStackTrace ...string) *WorkerPoolTask {
	var stackTrace string
	if len(optionalStackTrace) > 1 {
		stackTrace = optionalStackTrace[0]
	}

	return newWorkerPoolTask(b.decreasePendingTasksCounter, f, stackTrace)
}

// SubmitTask submits a task to the queue and blocks if the queue is full (it should only be used instead of Submit if
// manually handling the task is necessary to create better debug outputs).
func (b *NonBlockingWorkerPool) SubmitTask(task *WorkerPoolTask) {
	b.increasePendingTasksCounter()
	if !b.IsRunning() {
		task.markDone()

		return
	}

	b.appendPendingTask(task)
	b.undispatchedTaskAdded.Signal()
}

// Start starts the WorkerPool (non-blocking).
func (b *NonBlockingWorkerPool) Start() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.running {
		if b.shutdown {
			panic("Worker was already used before")
		}
		b.running = true

		go b.dispatchTasksLoop()
		b.startWorkers()
	}
}

// Run starts the WorkerPool and waits for its shutdown.
func (b *NonBlockingWorkerPool) Run() {
	b.Start()

	b.workers.Wait()
}

// Stop stops the WorkerPool.
func (b *NonBlockingWorkerPool) Stop() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.running {
		b.shutdown = true
		b.running = false

		b.ctxCancel()
		b.undispatchedTaskAdded.Signal()
	}
}

// StopAndWait stops the WorkerPool and waits for its shutdown.
func (b *NonBlockingWorkerPool) StopAndWait() {
	b.Stop()
	b.workers.Wait()
}

// WorkerCount returns the worker count for the WorkerPool.
func (b *NonBlockingWorkerPool) WorkerCount() int {
	return b.options.WorkerCount
}

// GetPendingQueueSize returns the amount of tasks pending to the processed.
func (b *NonBlockingWorkerPool) PendingTasksCount() uint64 {
	b.pendingTasksCounterMutex.RLock()
	defer b.pendingTasksCounterMutex.RUnlock()

	return b.pendingTasksCounter
}

// WaitUntilAllTasksProcessed waits until all tasks are processed.
func (b *NonBlockingWorkerPool) WaitUntilAllTasksProcessed() {
	b.pendingTasksCounterMutex.Lock()
	defer b.pendingTasksCounterMutex.Unlock()

	for b.pendingTasksCounter > 0 {
		b.allTasksProcessed.Wait()
	}
}

// IsRunning returns true if the WorkerPool is running.
func (b *NonBlockingWorkerPool) IsRunning() (isRunning bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return !b.shutdown
}

func (b *NonBlockingWorkerPool) UndispatchedTaskCount() int {
	b.undispatchedTasksMutex.RLock()
	defer b.undispatchedTasksMutex.RUnlock()

	return b.undispatchedTasks.Len()
}

func (b *NonBlockingWorkerPool) appendPendingTask(task *WorkerPoolTask) {
	b.undispatchedTasksMutex.Lock()
	defer b.undispatchedTasksMutex.Unlock()

	b.undispatchedTasks.PushBack(task)
}

func (b *NonBlockingWorkerPool) dispatchTasksLoop() {
	for b.IsRunning() || (b.options.FlushTasksAtShutdown && b.UndispatchedTaskCount() > 0) {
		b.dispatchTask()
	}
	b.WaitUntilAllTasksProcessed()
	close(b.nextTaskChan)
}

func (b *NonBlockingWorkerPool) dispatchTask() {
	b.undispatchedTasksMutex.Lock()
	defer b.undispatchedTasksMutex.Unlock()

	for b.undispatchedTasks.Len() == 0 {
		if !b.IsRunning() {
			return
		}
		b.undispatchedTaskAdded.Wait()
	}

	b.nextTaskChan <- b.undispatchedTasks.Remove(b.undispatchedTasks.Front()).(*WorkerPoolTask)
	b.undispatchedTaskRemoved.Broadcast()
}

func (b *NonBlockingWorkerPool) startWorkers() {
	for i := 0; i < b.options.WorkerCount; i++ {
		b.workers.Add(1)

		go func() {
		consumeLoop:
			for {
				select {
				case <-b.ctx.Done():
					b.flushTasks()

					break consumeLoop
				default:
					select {
					case <-b.ctx.Done():
						b.flushTasks()

						break consumeLoop
					case currentTask, ok := <-b.nextTaskChan:
						if !ok {
							break consumeLoop
						}

						currentTask.run()
					}
				}
			}

			b.workers.Done()
		}()
	}
}

func (b *NonBlockingWorkerPool) flushTasks() {
	if b.options.FlushTasksAtShutdown {
		for task, ok := <-b.nextTaskChan; ok; task, ok = <-b.nextTaskChan {
			task.run()
		}
	}
}

// IncreasePendingTasksCounter increases the pending task counter.
func (b *NonBlockingWorkerPool) increasePendingTasksCounter() {
	b.pendingTasksCounterMutex.Lock()
	defer b.pendingTasksCounterMutex.Unlock()

	b.pendingTasksCounter++
}

// DecreasePendingTasksCounter decreases the pending task counter.
func (b *NonBlockingWorkerPool) decreasePendingTasksCounter() {
	b.pendingTasksCounterMutex.Lock()

	b.pendingTasksCounter--
	if b.pendingTasksCounter != 0 {
		b.pendingTasksCounterMutex.Unlock()

		return
	}
	b.pendingTasksCounterMutex.Unlock()
	b.allTasksProcessed.Broadcast()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
