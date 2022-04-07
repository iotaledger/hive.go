package workerpool

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
)

// EventLoop represents a set of workers with a blocking queue of pending tasks.
type EventLoop struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	options *Options

	calls chan func()

	running  bool
	shutdown bool

	mutex syncutils.RWMutex
	wait  sync.WaitGroup
}

// NewEventLoop returns a new stopped EventLoop.
func NewEventLoop(optionalOptions ...Option) (result *EventLoop) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &EventLoop{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		options:   options,
		calls:     make(chan func(), options.QueueSize),
	}

	return
}

// Submit submits a task to the loop, if the queue is full the call blocks until the task is succesfully submitted.
func (e *EventLoop) Submit(f func()) {
	e.mutex.RLock()

	if e.shutdown {
		e.mutex.RUnlock()
		return
	}
	e.mutex.RUnlock()

	e.calls <- f
}

// TrySubmit submits a task to the loop without blocking, it returns false if the queue is full and the task was not
// succesfully submitted.
func (e *EventLoop) TrySubmit(f func()) (added bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if !e.shutdown {
		select {
		case e.calls <- f:
			return true
		default:
			return false
		}
	}

	return false
}

// Run starts the EventLoop.
func (e *EventLoop) Start() {
	e.mutex.Lock()

	if !e.running {
		if e.shutdown {
			panic("Worker was already used before")
		}
		e.running = true

		e.startWorkers()
	}

	e.mutex.Unlock()
}

// Run starts the EventLoop and waits for its shutdown.
func (e *EventLoop) Run() {
	e.Start()

	e.wait.Wait()
}

// Run stops the EventLoop.
func (e *EventLoop) Stop() {
	e.mutex.Lock()

	if e.running {
		e.shutdown = true
		e.running = false

		e.ctxCancel()
	}

	e.mutex.Unlock()
}

// Run stops the EventLoop and waits for its shutdown.
func (e *EventLoop) StopAndWait() {
	e.Stop()
	e.wait.Wait()
}

// GetWorkerCount returns the worker count for the EventLoop.
func (e *EventLoop) GetWorkerCount() int {
	return e.options.WorkerCount
}

// GetPendingQueueSize returns the amount of tasks pending to the processed.
func (e *EventLoop) GetPendingQueueSize() int {
	return len(e.calls)
}

func (e *EventLoop) startWorkers() {
	for i := 0; i < e.options.WorkerCount; i++ {
		e.wait.Add(1)

		go func() {
			aborted := false

			for !aborted {
				select {

				case <-e.ctx.Done():
					aborted = true

					if e.options.FlushTasksAtShutdown {
					terminateLoop:
						// process all waiting tasks after shutdown signal
						for {
							select {
							case f := <-e.calls:
								f()

							default:
								break terminateLoop
							}
						}
					}
				case f := <-e.calls:
					f()
				}
			}

			e.wait.Done()
		}()

	}
}
