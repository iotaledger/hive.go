package workerpool

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/core/debug"
	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/hive.go/core/types"
)

// region Task /////////////////////////////////////////////////////////////////////////////////////////////////////////

type Task struct {
	params     []interface{}
	resultChan chan interface{}
}

func (task *Task) Return(result interface{}) {
	task.resultChan <- result
	close(task.resultChan)
}

func (task *Task) Param(index int) interface{} {
	return task.params[index]
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region BlockingQueueWorkerPoolTask //////////////////////////////////////////////////////////////////////////////////

// WorkerPoolTask is a task that is executed by a BlockingQueuedWorkerPool.
type WorkerPoolTask struct {
	decreasePendingCounterFunc func()
	workerFunc                 func()
	doneChan                   chan types.Empty
	stackTrace                 string
}

// newWorkerPoolTask creates a new BlockingQueueWorkerPoolTask.
func newWorkerPoolTask(decreasePendingCounterFunc func(), workerFunc func(), stackTrace string) *WorkerPoolTask {
	if debug.GetEnabled() && stackTrace == "" {
		stackTrace = debug.ClosureStackTrace(workerFunc)
	}

	return &WorkerPoolTask{
		decreasePendingCounterFunc: decreasePendingCounterFunc,
		workerFunc:                 workerFunc,
		doneChan:                   make(chan types.Empty),
		stackTrace:                 stackTrace,
	}
}

// run executes the task.
func (t *WorkerPoolTask) run() {
	if debug.GetEnabled() {
		go t.detectedHangingTasks()
	}

	t.workerFunc()
	t.markDone()
}

// markDone marks the task as done.
func (t *WorkerPoolTask) markDone() {
	close(t.doneChan)
	t.decreasePendingCounterFunc()
}

// detectedHangingTasks is a debug method that is used to print information about possibly hanging task executions.
func (t *WorkerPoolTask) detectedHangingTasks() {
	timer := time.NewTimer(debug.DeadlockDetectionTimeout)
	defer timeutil.CleanupTimer(timer)

	select {
	case <-t.doneChan:
		return
	case <-timer.C:
		fmt.Println("task in workerpool seems to hang (" + debug.DeadlockDetectionTimeout.String() + ") ...")
		fmt.Println("\n" + strings.Replace(strings.Replace(t.stackTrace, "closure:", "task:", 1), "called by", "queued by", 1))
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
