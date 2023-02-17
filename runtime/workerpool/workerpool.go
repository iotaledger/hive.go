package workerpool

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/syncutils"
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
func (u *WorkerPool) Start() *WorkerPool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if !u.isRunning {
		u.ShutdownComplete.Wait()

		u.isRunning = true

		u.startDispatcher()
		u.startWorkers()
	}

	return u
}

// Submit submits a new task to the WorkerPool.
func (u *WorkerPool) Submit(workerFunc func(), optStackTrace ...string) {
	if !u.IsRunning() {
		panic(fmt.Sprintf("worker pool '%s' is not running", u.Name))
	}

	u.increasePendingTasks()

	u.Queue.Push(newTask(workerFunc, u.decreasePendingTasks, lo.First(optStackTrace)))
}

// IsRunning returns true if the WorkerPool is running.
func (u *WorkerPool) IsRunning() bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.isRunning
}

// WorkerCount returns the number of workers that are used to execute tasks.
func (u *WorkerPool) WorkerCount() int {
	return u.workerCount
}

// Shutdown shuts down the WorkerPool.
func (u *WorkerPool) Shutdown(cancelPendingTasks ...bool) *WorkerPool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.isRunning {
		u.isRunning = false

		for i := 0; i < u.workerCount; i++ {
			u.shutdownSignal <- len(cancelPendingTasks) >= 1 && cancelPendingTasks[0]
		}

		u.Queue.SignalShutdown()
	}

	return u
}

// increasePendingTasks increases the number of pending tasks.
func (u *WorkerPool) increasePendingTasks() {
	u.PendingTasksCounter.Increase()
}

// decreasePendingTasks decreases the number of pending tasks.
func (u *WorkerPool) decreasePendingTasks() {
	u.PendingTasksCounter.Decrease()
}

// startDispatcher starts the dispatcher that dispatches tasks to the workers.
func (u *WorkerPool) startDispatcher() {
	u.dispatcherChan = make(chan *Task, u.workerCount)

	go u.dispatcher()
}

// dispatcher is the dispatcher that dispatches tasks to the workers.
func (u *WorkerPool) dispatcher() {
	for u.IsRunning() || u.Queue.Size() > 0 {
		if task, success := u.Queue.PopOrWait(u.IsRunning); success {
			u.dispatcherChan <- task
		}
	}

	u.PendingTasksCounter.WaitIsZero()

	close(u.dispatcherChan)
}

// startWorkers starts the workers that execute tasks.
func (u *WorkerPool) startWorkers() {
	for i := 0; i < u.workerCount; i++ {
		u.ShutdownComplete.Add(1)

		go u.worker()
	}
}

// worker is a worker that executes tasks.
func (u *WorkerPool) worker() {
	defer u.ShutdownComplete.Done()

	u.handleShutdown(u.workerReadLoop())
}

// workerReadLoop reads tasks from the dispatcherChan and executes them.
func (u *WorkerPool) workerReadLoop() bool {
	for {
		select {
		case shutdownSignal := <-u.shutdownSignal:
			return shutdownSignal
		default:
			select {
			case shutdownSignal := <-u.shutdownSignal:
				return shutdownSignal
			case element, success := <-u.dispatcherChan:
				if !success {
					return false
				}

				element.run()
			}
		}
	}
}

// handleShutdown handles the shutdown of the worker.
func (u *WorkerPool) handleShutdown(cancelPendingTasks bool) {
	for task, success := <-u.dispatcherChan; success; task, success = <-u.dispatcherChan {
		if cancelPendingTasks {
			task.markDone()
		} else {
			task.run()
		}
	}
}
