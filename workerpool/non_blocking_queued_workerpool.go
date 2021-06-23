package workerpool

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/iotaledger/hive.go/datastructure/queue"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/panjf2000/ants/v2"
)

// NonBlockingQueuedWorkerPool implements a non-blocking goroutine pool backed by a queue.
type NonBlockingQueuedWorkerPool struct {
	workerFunc func(Task)
	options    *Options

	pool         *ants.PoolWithFunc
	queue        *queue.Queue
	shutdown     typeutils.AtomicBool
	running      typeutils.AtomicBool
	shutdownOnce sync.Once

	mutex   syncutils.RWMutex
	tasksWg sync.WaitGroup
}

// NewNonBlockingQueuedWorkerPool creates and starts a new worker pool for the supplied function, with the supplied options.
func NewNonBlockingQueuedWorkerPool(workerFunc func(Task), optionalOptions ...Option) (result *NonBlockingQueuedWorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	result = &NonBlockingQueuedWorkerPool{
		workerFunc: workerFunc,
		options:    options,
	}

	workerCount := options.WorkerCount

	if newPool, err := ants.NewPoolWithFunc(workerCount, result.workerFuncWrapper, ants.WithNonblocking(true)); err != nil {
		panic(err)
	} else {
		result.running.Set()
		result.pool = newPool
		result.queue = queue.New(options.QueueSize)
	}

	return
}

func (wp *NonBlockingQueuedWorkerPool) workerFuncWrapper(t interface{}) {
	// always execute at least 1 task: the one that was invoked via wp.pool.Invoke
	taskAvailable := true
	for taskAvailable {
		// wrap into inner function to continue execution with worker even if there's a panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("recovered from panic in WorkerPool: %s %s", r, debug.Stack())
				}

				wp.tasksWg.Done()
			}()

			wp.workerFunc(t.(Task))
		}()

		// reuse worker as long as there are tasks in the queue
		wp.mutex.RLock()
		t, taskAvailable = wp.queue.Poll()
		wp.mutex.RUnlock()
	}
}

func (wp *NonBlockingQueuedWorkerPool) doSubmit(t Task) bool {
	if antsErr := wp.pool.Invoke(t); antsErr != nil && antsErr != ants.ErrPoolOverload {
		panic(antsErr)
	} else {
		if antsErr == nil {
			return true
		}

		if wp.queue.Offer(t) {
			return true
		}

		// Queue cannot accommodate more tasks, dropping
		return false
	}
}

// Submit is an alias for TrySubmit
func (wp *NonBlockingQueuedWorkerPool) Submit(params ...interface{}) (chan interface{}, bool) {
	return wp.TrySubmit(params...)
}

// TrySubmit submits a task to this pool (it drops the task if not enough workers are available and the queue is full).
// It returns a channel to obtain the task result, and a boolean if the task was successfully submitted to the queue.
func (wp *NonBlockingQueuedWorkerPool) TrySubmit(params ...interface{}) (result chan interface{}, added bool) {
	if wp.shutdown.IsSet() {
		return nil, false
	}

	result = make(chan interface{}, 1)
	t := Task{
		params:     params,
		resultChan: result,
	}

	wp.mutex.Lock()
	defer wp.mutex.Unlock()
	if !wp.doSubmit(t) {
		close(result)
		return nil, false
	}

	wp.tasksWg.Add(1)

	return result, true
}

// Stop closes this pool. If FlushTasksAtShutdown was set, it allows currently running and pending tasks to complete.
func (wp *NonBlockingQueuedWorkerPool) Stop() {
	wp.shutdownOnce.Do(func() {
		wp.shutdown.Set()

		if wp.pool != nil {
			go func() {
				if wp.options.FlushTasksAtShutdown {
					wp.tasksWg.Wait()
				} else {
					for {
						_, polled := wp.queue.Poll()
						if !polled {
							break
						}
						wp.tasksWg.Done()
					}
				}
				wp.pool.Release()
			}()
		}
	})
}

// StopAndWait closes the pool and waits for tasks to complete.
func (wp *NonBlockingQueuedWorkerPool) StopAndWait() {
	wp.Stop()
	wp.tasksWg.Wait()
}

// GetWorkerCount gets the configured worker count.
func (wp *NonBlockingQueuedWorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}

// GetPendingQueueSize gets the current amount of pending tasks in the queue.
func (wp *NonBlockingQueuedWorkerPool) GetPendingQueueSize() int {
	return wp.queue.Size()
}
