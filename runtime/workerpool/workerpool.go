package workerpool

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/iotaledger/hive.go/core/syncutils"
)

type WorkerPool struct {
	Name                string
	PendingTasksCounter *syncutils.Counter
	Queue               *syncutils.Stack[*Task]
	ShutdownComplete    sync.WaitGroup

	workerCount    int
	isRunning      bool
	dispatcherChan chan *Task
	shutdownSignal chan bool
	mutex          syncutils.RWMutex
}

func New(name string, optsWorkerCount ...int) (newWorkerPool *WorkerPool) {
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

func (u *WorkerPool) Start() (self *WorkerPool) {
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

func (u *WorkerPool) Submit(task func(), optStackTrace ...string) {
	if !u.IsRunning() {
		panic(fmt.Sprintf("worker pool '%s' is not running", u.Name))
	}

	if u.PendingTasksCounter.Increase(); len(optStackTrace) >= 1 {
		u.Queue.Push(newTask(func() { u.PendingTasksCounter.Decrease() }, task, optStackTrace[0]))
	} else {
		u.Queue.Push(newTask(func() { u.PendingTasksCounter.Decrease() }, task, ""))
	}
}

func (u *WorkerPool) Shutdown(cancelPendingTasks ...bool) (self *WorkerPool) {
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

func (u *WorkerPool) IsRunning() (isRunning bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.isRunning
}

func (u *WorkerPool) WorkerCount() (workerCount int) {
	return u.workerCount
}

func (u *WorkerPool) startDispatcher() {
	u.dispatcherChan = make(chan *Task, u.workerCount)

	go u.dispatcher()
}

func (u *WorkerPool) dispatcher() {
	for u.IsRunning() || u.Queue.Size() > 0 {
		if task, success := u.Queue.PopOrWait(u.IsRunning); success {
			u.dispatcherChan <- task
		}
	}

	u.PendingTasksCounter.WaitIsZero()

	close(u.dispatcherChan)
}

func (u *WorkerPool) startWorkers() {
	for i := 0; i < u.workerCount; i++ {
		u.ShutdownComplete.Add(1)

		go u.worker()
	}
}

func (u *WorkerPool) worker() {
	defer u.ShutdownComplete.Done()

	u.handleShutdown(u.workerReadLoop())
}

func (u *WorkerPool) workerReadLoop() (shutdownSignal bool) {
	for {
		select {
		case shutdownSignal = <-u.shutdownSignal:
			return
		default:
			select {
			case shutdownSignal = <-u.shutdownSignal:
				return
			case element, success := <-u.dispatcherChan:
				if !success {
					return
				}

				element.run()
			}
		}
	}
}

func (u *WorkerPool) handleShutdown(cancelPendingTasks bool) {
	for task, success := <-u.dispatcherChan; success; task, success = <-u.dispatcherChan {
		if cancelPendingTasks {
			task.markDone()
		} else {
			task.run()
		}
	}
}
