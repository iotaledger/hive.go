package workerpool

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

// WorkerPool is a pool of workers that can execute tasks.
type WorkerPool struct {
	// Name is the name of the WorkerPool.
	Name string

	// PendingTasksCounter is the number of tasks that are currently pending.
	PendingTasksCounter *syncutils.Counter

	// Queue is the queue of tasks that are waiting to be executed.
	Queue *syncutils.Stack[*Task]

	// ShutdownComplete is a WaitGroup that is used to wait for the WorkerPool to shutdown.
	ShutdownComplete sync.WaitGroup

	// isRunning indicates if the WorkerPool is running.
	isRunning bool

	// dispatcherChan is the channel that is used to dispatch tasks to the workers.
	dispatcherChan chan *Task

	// shutdownSignal is the channel that is used to signal the workers to shut down.
	shutdownSignal chan struct{}

	// workerCount is the number of workers that are used to execute tasks, defaults to twice the amount of logical CPUs.
	workerCount int

	// optPanicOnSubmitAfterShutdown indicates if a panic should be triggered when a task is submitted after the WorkerPool was shut down.
	optPanicOnSubmitAfterShutdown bool

	// optCancelPendingTasksOnShutdown indicates if pending tasks should be canceled on shutdown.
	optCancelPendingTasksOnShutdown bool

	// mutex is used to synchronize access to the WorkerPool.
	mutex syncutils.RWMutex
}

// New creates a new WorkerPool with the given name and returns it.
func New(name string, opts ...options.Option[WorkerPool]) *WorkerPool {
	return options.Apply(&WorkerPool{
		Name:                name,
		PendingTasksCounter: syncutils.NewCounter(),
		Queue:               syncutils.NewStack[*Task](),
		workerCount:         2 * runtime.NumCPU(),

		optCancelPendingTasksOnShutdown: false,
		optPanicOnSubmitAfterShutdown:   false,
	}, opts, func(w *WorkerPool) {
		w.shutdownSignal = make(chan struct{}, w.workerCount)
	})
}

// Start starts the WorkerPool.
func (w *WorkerPool) Start() *WorkerPool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isRunning {
		w.ShutdownComplete.Wait()

		w.isRunning = true

		w.startDispatcher()
		w.startWorkers()
	}

	return w
}

// Submit submits a new task to the WorkerPool.
func (w *WorkerPool) Submit(workerFunc func(), optStackTrace ...string) {
	if !w.IsRunning() {
		if w.optPanicOnSubmitAfterShutdown {
			panic(fmt.Sprintf("worker pool '%s' is not running", w.Name))
		}

		return
	}

	w.increasePendingTasks()

	w.Queue.Push(newTask(workerFunc, w.decreasePendingTasks, lo.First(optStackTrace)))
}

// DebounceFunc returns a function that can be used to submit a task that is canceled if the function is called with a
// new task before the previous task was executed.
func (w *WorkerPool) DebounceFunc() (debounce func(workerFunc func(), optStackTrace ...string)) {
	// lastInvocation is used to determine if a new invocation was submitted before the previous one was executed
	var lastInvocation atomic.Uint64

	// execMutex is used to synchronize the execution of the workerFunc (only one should be executed at a time)
	var execMutex sync.Mutex

	return func(workerFunc func(), optStackTrace ...string) {
		// increment the invocation counter
		currentInvocation := lastInvocation.Add(1)

		// submit the task
		w.Submit(func() {
			// abort if the current invocation is not the last one anymore
			if currentInvocation != lastInvocation.Load() {
				return
			}

			// grab execution mutex
			execMutex.Lock()
			defer execMutex.Unlock()

			// abort if the current invocation is not the last one anymore (double-checked locking)
			if currentInvocation != lastInvocation.Load() {
				return
			}

			// execute the workerFunc
			workerFunc()
		}, optStackTrace...)
	}
}

// IsRunning returns true if the WorkerPool is running.
func (w *WorkerPool) IsRunning() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.isRunning
}

// WorkerCount returns the number of workers that are used to execute tasks.
func (w *WorkerPool) WorkerCount() int {
	return w.workerCount
}

// Shutdown shuts down the WorkerPool.
func (w *WorkerPool) Shutdown() *WorkerPool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		w.isRunning = false

		for range w.workerCount {
			w.shutdownSignal <- struct{}{}
		}

		w.Queue.SignalShutdown()
	}

	return w
}

// increasePendingTasks increases the number of pending tasks.
func (w *WorkerPool) increasePendingTasks() {
	w.PendingTasksCounter.Increase()
}

// decreasePendingTasks decreases the number of pending tasks.
func (w *WorkerPool) decreasePendingTasks() {
	w.PendingTasksCounter.Decrease()
}

// startDispatcher starts the dispatcher that dispatches tasks to the workers.
func (w *WorkerPool) startDispatcher() {
	w.dispatcherChan = make(chan *Task, w.workerCount)

	go w.dispatcher()
}

// dispatcher is the dispatcher that dispatches tasks to the workers.
func (w *WorkerPool) dispatcher() {
	for w.IsRunning() || w.Queue.Size() > 0 {
		if task, success := w.Queue.PopOrWait(w.IsRunning); success {
			w.dispatcherChan <- task
		}
	}

	w.PendingTasksCounter.WaitIsZero()

	close(w.dispatcherChan)
}

// startWorkers starts the workers that execute tasks.
func (w *WorkerPool) startWorkers() {
	for range w.workerCount {
		w.ShutdownComplete.Add(1)

		go w.worker()
	}
}

// worker is a worker that executes tasks.
func (w *WorkerPool) worker() {
	defer w.ShutdownComplete.Done()

	w.workerReadLoop()

	w.handleShutdown()
}

// workerReadLoop reads tasks from the dispatcherChan and executes them.
func (w *WorkerPool) workerReadLoop() {
	for {
		select {
		case <-w.shutdownSignal:
			return
		default:
			select {
			case <-w.shutdownSignal:
				return
			case element, success := <-w.dispatcherChan:
				if !success {
					return
				}

				element.run()
			}
		}
	}
}

// handleShutdown handles the shutdown of the worker.
func (w *WorkerPool) handleShutdown() {
	for task, success := <-w.dispatcherChan; success; task, success = <-w.dispatcherChan {
		if w.optCancelPendingTasksOnShutdown {
			task.markDone()
		} else {
			task.run()
		}
	}
}

// WithWorkerCount is an option for the WorkerPool that allows to set the number of workers that are used to execute tasks.
func WithWorkerCount(workerCount int) options.Option[WorkerPool] {
	return func(w *WorkerPool) {
		w.workerCount = workerCount
	}
}

// WithPanicOnSubmitAfterShutdown is an option for the WorkerPool that allows to set if a panic should be triggered when a task is submitted after the WorkerPool was shut down.
func WithPanicOnSubmitAfterShutdown(panicOnSubmitAfterShutdown bool) options.Option[WorkerPool] {
	return func(w *WorkerPool) {
		w.optPanicOnSubmitAfterShutdown = panicOnSubmitAfterShutdown
	}
}

// WithCancelPendingTasksOnShutdown is an option for the WorkerPool that allows to set if pending tasks should be canceled on shutdown.
func WithCancelPendingTasksOnShutdown(cancelPendingTasksOnShutdown bool) options.Option[WorkerPool] {
	return func(w *WorkerPool) {
		w.optCancelPendingTasksOnShutdown = cancelPendingTasksOnShutdown
	}
}
