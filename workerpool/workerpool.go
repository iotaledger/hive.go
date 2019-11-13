package workerpool

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/syncutils"
	"sync"
)

type WorkerPool struct {
	workerFnc func(Task)
	options   *Options

	calls       chan Task
	terminate   chan int
	barrierIn   chan struct{}
	barrierOut  chan struct{}
	barrierWait sync.WaitGroup
	barrierLock syncutils.Mutex

	BarrierEvent *events.Event

	running bool
	mutex   syncutils.RWMutex
	wait    sync.WaitGroup
}

func New(workerFnc func(Task), optionalOptions ...Option) (result *WorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	result = &WorkerPool{
		workerFnc: workerFnc,
		options:   options,
	}

	result.resetChannels()

	return
}

func NewWithBarrier(workerFnc func(Task), barrierEvent *events.Event, optionalOptions ...Option) (result *WorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	result = &WorkerPool{
		workerFnc:    workerFnc,
		options:      options,
		BarrierEvent: barrierEvent,
	}

	result.resetChannels()

	return
}

func (wp *WorkerPool) submit(barrier bool, params ...interface{}) (result chan interface{}) {
	result = make(chan interface{}, 1)

	wp.mutex.RLock()

	wp.calls <- Task{
		params:     params,
		resultChan: result,
		barrier:    barrier,
	}

	wp.mutex.RUnlock()

	return
}

func (wp *WorkerPool) trySubmit(params ...interface{}) (result chan interface{}, added bool) {
	result = make(chan interface{}, 1)

	wp.mutex.RLock()

	select {
	case wp.calls <- Task{
		params:     params,
		resultChan: result,
		barrier:    false,
	}:
		added = true
	default:
		// Queue full => drop the task
		added = false
	}

	wp.mutex.RUnlock()

	return
}

func (wp *WorkerPool) Submit(params ...interface{}) (result chan interface{}) {
	return wp.submit(false, params...)
}

func (wp *WorkerPool) TrySubmit(params ...interface{}) (result chan interface{}, added bool) {
	return wp.trySubmit(params...)
}

func (wp *WorkerPool) SubmitBarrier(params ...interface{}) (result chan interface{}) {
	return wp.submit(true, params...)
}

func (wp *WorkerPool) Start() {
	wp.mutex.Lock()

	if !wp.running {
		wp.running = true

		wp.startWorkers()
	}

	wp.mutex.Unlock()
}

func (wp *WorkerPool) Run() {
	wp.Start()

	wp.wait.Wait()
}

func (wp *WorkerPool) Stop() {
	go wp.StopAndWait()
}

func (wp *WorkerPool) StopAndWait() {
	wp.mutex.Lock()

	if wp.running {
		wp.running = false

		close(wp.terminate)
	}

	wp.wait.Wait()
	wp.resetChannels()

	wp.mutex.Unlock()
}

func (wp *WorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}

func (wp *WorkerPool) GetPendingQueueSize() int {
	return len(wp.calls)
}

func (wp *WorkerPool) signalBarrier(params ...interface{}) {
	if !wp.running {
		return
	}

	wp.barrierLock.Lock()
	wp.barrierWait.Add(1)

	// signal all threads to enter the barrier after processing their last task or being idle
	for i := 0; i < wp.options.WorkerCount; i++ {
		wp.barrierIn <- struct{}{}
	}

	// wait for all threads to enter the barrier
	for i := 0; i < wp.options.WorkerCount; i++ {
		<-wp.barrierOut
	}

	wp.barrierWait.Done()
	wp.BarrierEvent.Trigger(params...)
	wp.barrierLock.Unlock()
}

func (wp *WorkerPool) resetChannels() {
	wp.calls = make(chan Task, wp.options.QueueSize)
	wp.terminate = make(chan int, 1)
	wp.barrierIn = make(chan struct{}, wp.options.WorkerCount)
	wp.barrierOut = make(chan struct{}, wp.options.WorkerCount)
}

func (wp *WorkerPool) startWorkers() {
	calls := wp.calls
	terminate := wp.terminate

	for i := 0; i < wp.options.WorkerCount; i++ {
		wp.wait.Add(1)

		go func() {
			aborted := false

			for !aborted {
				select {

				case <-terminate:
					aborted = true

				case <-wp.barrierIn:
					wp.barrierOut <- struct{}{}
					wp.barrierWait.Wait()

				case batchTask := <-calls:
					if !batchTask.barrier {
						wp.workerFnc(batchTask)
					} else {
						// barrier detected, signal all workers
						go wp.signalBarrier(batchTask.params...)
					}
				}
			}

			wp.wait.Done()
		}()
	}
}
