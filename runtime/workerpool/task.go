package workerpool

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/core/debug"
	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/hive.go/core/types"
)

// Task is a task that is executed by a WorkerPool.
type Task struct {
	decreasePendingCounterFunc func()
	workerFunc                 func()
	doneChan                   chan types.Empty
	stackTrace                 string
}

// newTask creates a new BlockingQueueWorkerPoolTask.
func newTask(decreasePendingCounterFunc func(), workerFunc func(), stackTrace string) *Task {
	if debug.GetEnabled() && stackTrace == "" {
		stackTrace = debug.ClosureStackTrace(workerFunc)
	}

	return &Task{
		decreasePendingCounterFunc: decreasePendingCounterFunc,
		workerFunc:                 workerFunc,
		doneChan:                   make(chan types.Empty),
		stackTrace:                 stackTrace,
	}
}

// run executes the task.
func (t *Task) run() {
	if debug.GetEnabled() {
		go t.detectedHangingTasks()
	}

	t.workerFunc()
	t.markDone()
}

// markDone marks the task as done.
func (t *Task) markDone() {
	close(t.doneChan)
	t.decreasePendingCounterFunc()
}

// detectedHangingTasks is a debug method that is used to print information about possibly hanging task executions.
func (t *Task) detectedHangingTasks() {
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
