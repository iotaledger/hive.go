package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	debug2 "github.com/iotaledger/hive.go/debug"
	"github.com/iotaledger/hive.go/syncutils"
)

// region BlockQueuedWorkerPool ////////////////////////////////////////////////////////////////////////////////////////

// BlockingQueuedWorkerPool represents a set of workers with a blocking queue of pending tasks.
type BlockingQueuedWorkerPool struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	options *Options

	calls chan task

	running  bool
	shutdown bool

	pendingTasksCounter        uint64
	pendingTasksMutex          sync.Mutex
	waitUntilAllTasksProcessed *sync.Cond
	closed                     uint64

	workerIDs      map[uint64]bool
	workerIDsMutex sync.RWMutex

	mutex syncutils.RWMutex
	wait  sync.WaitGroup
}

type task struct {
	f     func()
	lines []int
	files []string
}

func (t task) String() string {
	s := strings.Builder{}

	s.WriteString(fmt.Sprintf("%d ", debug2.GoroutineID()))
	s.WriteString(debug2.GetFunctionName(t.f))
	s.WriteString("\n")
	for i, file := range t.files {
		s.WriteString(fmt.Sprintf("\t %s:%d\n", file, t.lines[i]))
	}
	return s.String()
}

func (t task) run(wp *BlockingQueuedWorkerPool) {
	defer wp.decreasePendingTasksCounter()

	// fmt.Println(debug2.GoroutineID(), "EXECUTING", debug2.GetFunctionName(t.f))
	goroutine := debug2.GoroutineID()
	waiting := true
	go func() {
		time.Sleep(5 * time.Second)
		if waiting {
			fmt.Println(goroutine, "HANG!!", t)
		}
	}()
	t.f()

	waiting = false
	// fmt.Println(debug2.GoroutineID(), "EXECUTED", debug2.GetFunctionName(t.f))
}

// NewBlockingQueuedWorkerPool returns a new stopped WorkerPool.
func NewBlockingQueuedWorkerPool(optionalOptions ...Option) (result *BlockingQueuedWorkerPool) {
	options := DEFAULT_OPTIONS.Override(optionalOptions...)

	ctx, ctxCancel := context.WithCancel(context.Background())

	result = &BlockingQueuedWorkerPool{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		options:   options,
		calls:     make(chan task, options.QueueSize),
		workerIDs: make(map[uint64]bool),
	}
	result.waitUntilAllTasksProcessed = sync.NewCond(&result.pendingTasksMutex)

	return
}

// Submit submits a task to the loop, if the queue is full the call blocks until the task is succesfully submitted.
func (wp *BlockingQueuedWorkerPool) Submit(f func()) {
	if atomic.LoadUint64(&wp.closed) == 1 {
		wp.workerIDsMutex.RLock()
		if !wp.workerIDs[debug2.GoroutineID()] {
			fmt.Println(debug2.GoroutineID(), "WorkerPool was empty before submitting task")
			fmt.Println(wp.createTask(f))
		}
		wp.workerIDsMutex.RUnlock()
	}

	wp.mutex.RLock()
	if wp.shutdown {
		wp.mutex.RUnlock()
		return
	}
	wp.mutex.RUnlock()

	wp.increasePendingTaskCounter()
	// fmt.Println(debug2.GoroutineID(), "pendingTasks ADDED", pendingTasks)
	wp.calls <- wp.createTask(f)
}

func (wp *BlockingQueuedWorkerPool) createTask(f func()) task {
	t := task{
		f:     f,
		lines: make([]int, 0),
		files: make([]string, 0),
	}

	for i := 0; ; i++ {
		_, file, no, ok := runtime.Caller(i)
		if !ok {
			break
		}
		t.files = append(t.files, file)
		t.lines = append(t.lines, no)
	}
	return t
}

// TrySubmit submits a task to the loop without blocking, it returns false if the queue is full and the task was not
// succesfully submitted.
func (wp *BlockingQueuedWorkerPool) TrySubmit(f func()) (added bool) {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	if !wp.shutdown {
		wp.increasePendingTaskCounter()
		// fmt.Println(debug2.GoroutineID(), "pendingTasks ADDED", pendingTasks)
		select {
		case wp.calls <- wp.createTask(f):
			return true
		default:
			wp.decreasePendingTasksCounter()
			return false
		}
	}

	return false
}

// Start starts the WorkerPool (non-blocking).
func (wp *BlockingQueuedWorkerPool) Start() {
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

// Run starts the WorkerPool and waits for its shutdown.
func (wp *BlockingQueuedWorkerPool) Run() {
	wp.Start()

	wp.wait.Wait()
}

// Stop stops the WorkerPool.
func (wp *BlockingQueuedWorkerPool) Stop() {
	wp.mutex.Lock()

	if wp.running {
		wp.shutdown = true
		wp.running = false

		wp.ctxCancel()
	}

	wp.mutex.Unlock()
}

// StopAndWait stops the WorkerPool and waits for its shutdown.
func (wp *BlockingQueuedWorkerPool) StopAndWait() {
	wp.Stop()
	wp.wait.Wait()
}

// GetWorkerCount returns the worker count for the WorkerPool.
func (wp *BlockingQueuedWorkerPool) GetWorkerCount() int {
	return wp.options.WorkerCount
}

// GetPendingQueueSize returns the amount of tasks pending to the processed.
func (wp *BlockingQueuedWorkerPool) GetPendingQueueSize() int {
	return len(wp.calls)
}

func (wp *BlockingQueuedWorkerPool) startWorkers() {
	for i := 0; i < wp.options.WorkerCount; i++ {
		wp.wait.Add(1)

		go func() {
			wp.workerIDsMutex.Lock()
			wp.workerIDs[debug2.GoroutineID()] = true
			wp.workerIDsMutex.Unlock()

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
							case t := <-wp.calls:
								t.run(wp)
							default:
								break terminateLoop
							}
						}
					}
				case t := <-wp.calls:
					t.run(wp)
				}
			}

			wp.wait.Done()
		}()

	}
}

// WaitUntilAllTasksProcessed waits until all tasks are processed.
func (wp *BlockingQueuedWorkerPool) WaitUntilAllTasksProcessed() {
	wp.pendingTasksMutex.Lock()
	defer wp.pendingTasksMutex.Unlock()
	if wp.pendingTasksCounter == 0 {
		return
	}

	fmt.Println(debug2.GoroutineID(), "WAITING")
	for wp.pendingTasksCounter > 0 {
		wp.waitUntilAllTasksProcessed.Wait()
	}
	fmt.Println(debug2.GoroutineID(), "WAITING DONE")
	atomic.StoreUint64(&wp.closed, 0)
}

func (wp *BlockingQueuedWorkerPool) increasePendingTaskCounter() {
	wp.pendingTasksMutex.Lock()
	defer wp.pendingTasksMutex.Unlock()

	// fmt.Println(debug2.GoroutineID(), "increasePendingTasksCounter", wp.pendingTasksCounter)

	wp.pendingTasksCounter++
}

// decreasePendingTasksCounter decreases the pending tasks counter and closes the empty channel if necessary.
func (wp *BlockingQueuedWorkerPool) decreasePendingTasksCounter() {
	// fmt.Println(debug2.GoroutineID(), "decreasePendingTasksCounter")
	wp.pendingTasksMutex.Lock()

	// fmt.Println(debug2.GoroutineID(), "decreasePendingTasksCounter", wp.pendingTasksCounter)
	wp.pendingTasksCounter--
	if wp.pendingTasksCounter != 0 {
		wp.pendingTasksMutex.Unlock()
		return
	}
	atomic.StoreUint64(&wp.closed, 1)
	wp.pendingTasksMutex.Unlock()
	wp.waitUntilAllTasksProcessed.Broadcast()
	fmt.Println(debug2.GoroutineID(), "decreasePendingTasksCounter", "CLOSE")

	// fmt.Println(debug2.GoroutineID(), "pendingTasks", pendingTasks)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
