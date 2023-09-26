package workerpool

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/runtime/syncutils"
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

	// workerCount is the number of workers that are used to execute tasks.
	workerCount int

	// isRunning indicates if the WorkerPool is running.
	isRunning bool

	// dispatcherChan is the channel that is used to dispatch tasks to the workers.
	dispatcherChan chan *Task

	// shutdownSignal is the channel that is used to signal the workers to shut down.
	shutdownSignal chan bool

	// mutex is used to synchronize access to the WorkerPool.
	mutex syncutils.RWMutex
}

// New creates a new WorkerPool with the given name and returns it.
func New(name string, optsWorkerCount ...int) *WorkerPool {
	if len(optsWorkerCount) == 0 {
		optsWorkerCount = append(optsWorkerCount, 2*runtime.NumCPU())
	}

	return &WorkerPool{
		Name:                name,
		PendingTasksCounter: syncutils.NewCounter(),
		Queue:               syncutils.NewStack[*Task](),
		workerCount:         optsWorkerCount[0],
		shutdownSignal:      make(chan bool, optsWorkerCount[0]),
	}
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
		panic(fmt.Sprintf("worker pool '%s' is not running", w.Name))
	}

	w.increasePendingTasks()

	w.Queue.Push(newTask(workerFunc, w.decreasePendingTasks, lo.First(optStackTrace)))
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
func (w *WorkerPool) Shutdown(cancelPendingTasks ...bool) *WorkerPool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		w.isRunning = false

		for i := 0; i < w.workerCount; i++ {
			w.shutdownSignal <- len(cancelPendingTasks) >= 1 && cancelPendingTasks[0]
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
	for i := 0; i < w.workerCount; i++ {
		w.ShutdownComplete.Add(1)

		go w.worker()
	}
}

// worker is a worker that executes tasks.
func (w *WorkerPool) worker() {
	defer w.ShutdownComplete.Done()

	w.handleShutdown(w.workerReadLoop())
}

// workerReadLoop reads tasks from the dispatcherChan and executes them.
func (w *WorkerPool) workerReadLoop() bool {
	for {
		select {
		case shutdownSignal := <-w.shutdownSignal:
			return shutdownSignal
		default:
			select {
			case shutdownSignal := <-w.shutdownSignal:
				return shutdownSignal
			case element, success := <-w.dispatcherChan:
				if !success {
					return false
				}

				element.run()
			}
		}
	}
}

// handleShutdown handles the shutdown of the worker.
func (w *WorkerPool) handleShutdown(cancelPendingTasks bool) {
	for task, success := <-w.dispatcherChan; success; task, success = <-w.dispatcherChan {
		if cancelPendingTasks {
			task.markDone()
		} else {
			task.run()
		}
	}
}
