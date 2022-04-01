package workerpool

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/panjf2000/ants/v2"
)

// BlockingQueuedWorkerPool implements a non-blocking goroutine pool backed by a queue.
type BlockingQueuedWorkerPool struct {
	workerFunc func(Task)
	options    *Options

	pool         *ants.PoolWithFunc
	tasks        chan Task
	shutdown     bool
	shutdownOnce sync.Once

	shutdownMutex syncutils.RWMutex
	tasksWg       sync.WaitGroup
}

// NewBlockingQueuedWorkerPool creates and starts a new worker pool for the supplied function, with the supplied options.
func NewBlockingQueuedWorkerPool(workerFunc func(Task), optionalOptions ...Option) (result *BlockingQueuedWorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	result = &BlockingQueuedWorkerPool{
		workerFunc: workerFunc,
		options:    options,
	}

	if newPool, err := ants.NewPoolWithFunc(options.WorkerCount, result.workerFuncWrapper, ants.WithNonblocking(true)); err != nil {
		panic(err)
	} else {
		result.pool = newPool
		result.tasks = make(chan Task, options.QueueSize)
	}

	return
}

func (wp *BlockingQueuedWorkerPool) workerFuncWrapper(t interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic in WorkerPool: %s %s", r, debug.Stack())
		}

		close(t.(Task).resultChan)
		wp.tasksWg.Done()
	}()

	wp.workerFunc(t.(Task))
}

// Submits a task to this pool, blocks if the queue is full.
func (wp *BlockingQueuedWorkerPool) Submit(params ...interface{}) (result chan interface{}) {
	wp.shutdownMutex.RLock()
	defer wp.shutdownMutex.RUnlock()

	if wp.shutdown {
		return nil
	}

	result = make(chan interface{}, 1)
	t := Task{
		params:     params,
		resultChan: result,
	}

	wp.tasksWg.Add(1)
	antsErr := wp.pool.Invoke(t)

	if antsErr == nil {
		return
	}

	if antsErr != ants.ErrPoolOverload {
		close(result)
		panic(antsErr)
	}

	// This blocks if queue size reached
	wp.tasks <- t

	return
}

// Stop closes this pool. If FlushTasksAtShutdown was set, it allows currently running and pending tasks to complete.
func (wp *BlockingQueuedWorkerPool) Stop() {
	wp.shutdownOnce.Do(func() {
		wp.shutdownMutex.Lock()
		defer wp.shutdownMutex.Unlock()
		wp.shutdown = true

		if wp.pool != nil {
			go func() {
				if wp.options.FlushTasksAtShutdown {
					wp.tasksWg.Wait()
				} else {
					for {
						select {
						case <-wp.tasks:
							wp.tasksWg.Done()
						default:
							break
						}
					}
				}
				close(wp.tasks)
				wp.pool.Release()
			}()
		}
	})
}

// StopAndWait closes the pool and waits for tasks to complete.
func (wp *BlockingQueuedWorkerPool) StopAndWait() {
	wp.Stop()
	wp.tasksWg.Wait()
}

// GetWorkerCount gets the configured worker count.
func (wp *BlockingQueuedWorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}
