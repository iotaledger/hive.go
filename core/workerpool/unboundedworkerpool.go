package workerpool

import (
	"sync"

	"github.com/iotaledger/hive.go/core/syncutils"
)

type UnboundedWorkerPool struct {
	PendingTasksCounter *syncutils.Counter
	Queue               *syncutils.Stack[*WorkerPoolTask]
	ShutdownComplete    sync.WaitGroup

	workerCount    int
	isRunning      bool
	dispatcherChan chan *WorkerPoolTask
	shutdownSignal chan bool
	mutex          syncutils.RWMutex
}

func NewUnboundedWorkerPool(workerCount int) (newUnboundedWorkerPool *UnboundedWorkerPool) {
	return &UnboundedWorkerPool{
		PendingTasksCounter: syncutils.NewCounter(),
		Queue:               syncutils.NewStack[*WorkerPoolTask](),
		workerCount:         workerCount,
		shutdownSignal:      make(chan bool, workerCount),
	}
}

func (u *UnboundedWorkerPool) Start() (self *UnboundedWorkerPool) {
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

func (u *UnboundedWorkerPool) Shutdown(cancelPendingTasks ...bool) (self *UnboundedWorkerPool) {
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

func (u *UnboundedWorkerPool) Submit(task func(), optStackTrace ...string) {
	if !u.IsRunning() {
		panic("worker pool is not running")
	}

	if u.PendingTasksCounter.Increase(); len(optStackTrace) >= 1 {
		u.Queue.Push(newWorkerPoolTask(u.PendingTasksCounter.Decrease, task, optStackTrace[0]))
	} else {
		u.Queue.Push(newWorkerPoolTask(u.PendingTasksCounter.Decrease, task, ""))
	}
}

func (u *UnboundedWorkerPool) IsRunning() (isRunning bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.isRunning
}

func (u *UnboundedWorkerPool) WorkerCount() (workerCount int) {
	return u.workerCount
}

func (u *UnboundedWorkerPool) startDispatcher() {
	u.dispatcherChan = make(chan *WorkerPoolTask, u.workerCount)

	go u.dispatcher()
}

func (u *UnboundedWorkerPool) dispatcher() {
	for u.IsRunning() || u.Queue.Size() > 0 {
		if task, success := u.Queue.PopOrWait(u.IsRunning); success {
			u.dispatcherChan <- task
		}
	}

	u.PendingTasksCounter.WaitIsZero()

	close(u.dispatcherChan)
}

func (u *UnboundedWorkerPool) startWorkers() {
	u.ShutdownComplete = sync.WaitGroup{}
	for i := 0; i < u.workerCount; i++ {
		u.ShutdownComplete.Add(1)

		go u.worker()
	}
}

func (u *UnboundedWorkerPool) worker() {
	defer u.ShutdownComplete.Done()

readNextElement:
	select {
	case cancelPendingTasks := <-u.shutdownSignal:
		u.handleShutdown(cancelPendingTasks)
	default:
		select {
		case cancelPendingTasks := <-u.shutdownSignal:
			u.handleShutdown(cancelPendingTasks)
		case element, success := <-u.dispatcherChan:
			if success {
				element.run()
				goto readNextElement
			}
		}
	}
}

func (u *UnboundedWorkerPool) handleShutdown(cancelPendingTasks bool) {
	for task, success := <-u.dispatcherChan; success; task, success = <-u.dispatcherChan {
		if cancelPendingTasks {
			task.markDone()
		} else {
			task.run()
		}
	}
}
