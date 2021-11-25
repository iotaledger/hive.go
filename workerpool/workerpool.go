package workerpool

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
)

type WorkerPool struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	workerFnc func(Task)
	options   *Options

	calls chan Task

	running  bool
	shutdown bool

	mutex syncutils.RWMutex
	wait  sync.WaitGroup
}

func voidCaller(handler interface{}, params ...interface{}) {
	handler.(func())()
}

func New(workerFnc func(Task), optionalOptions ...Option) (result *WorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &WorkerPool{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		workerFnc: workerFnc,
		options:   options,
		calls:     make(chan Task, options.QueueSize),
	}

	return
}

func (wp *WorkerPool) Submit(params ...interface{}) (result chan interface{}, added bool) {

	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.shutdown {
		result = make(chan interface{}, 1)

		wp.calls <- Task{
			params:     params,
			resultChan: result,
		}
		return result, true
	}

	return nil, false
}

func (wp *WorkerPool) TrySubmit(params ...interface{}) (result chan interface{}, added bool) {

	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.shutdown {
		result = make(chan interface{}, 1)

		select {
		case wp.calls <- Task{
			params:     params,
			resultChan: result,
		}:
			return result, true
		default:
			// Queue full => drop the task
			close(result)
			return nil, false
		}
	}

	return nil, false
}

func (wp *WorkerPool) Start() {
	wp.mutex.Lock()

	if !wp.running {
		if wp.shutdown {
			panic("Worker was already used before")
		}
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
	wp.mutex.Lock()

	if wp.running {
		wp.shutdown = true
		wp.running = false

		wp.ctxCancel()
	}

	wp.mutex.Unlock()
}

func (wp *WorkerPool) StopAndWait() {
	wp.Stop()
	wp.wait.Wait()
}

func (wp *WorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}

func (wp *WorkerPool) GetPendingQueueSize() int {
	return len(wp.calls)
}

func (wp *WorkerPool) startWorkers() {

	for i := 0; i < wp.options.WorkerCount; i++ {
		wp.wait.Add(1)

		go func() {
			aborted := false

			for !aborted {
				select {

				case <-wp.ctx.Done():
					aborted = true

					if wp.options.FlushTasksAtShutdown {
					terminateLoop:
						// process all waiting tasks after shutdown signal
						for {
							select {
							case batchTask := <-wp.calls:
								wp.workerFnc(batchTask)

							default:
								break terminateLoop
							}
						}
					}

				case batchTask := <-wp.calls:
					wp.workerFnc(batchTask)
				}
			}

			wp.wait.Done()
		}()
	}
}
