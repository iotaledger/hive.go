package workerpool

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/runtime/debug"
	"github.com/iotaledger/hive.go/runtime/timeutil"
)

// Task is a task that is executed by a WorkerPool.
type Task struct {
	// workerFunc is the function that is executed by the WorkerPool.
	workerFunc func()

	// doneCallback is called when the Task is done.
	doneCallback func()

	// stackTrace is the stack trace of the Task.
	stackTrace string

	// doneChan is used to indicate that the Task is done.
	doneChan chan types.Empty
}

// newTask creates a new Task.
func newTask(workerFunc func(), doneCallback func(), stackTrace string) *Task {
	if debug.GetEnabled() && stackTrace == "" {
		stackTrace = debug.ClosureStackTrace(workerFunc)
	}

	return &Task{
		workerFunc:   workerFunc,
		doneCallback: doneCallback,
		stackTrace:   stackTrace,
		doneChan:     make(chan types.Empty),
	}
}

// run executes the Task.
func (t *Task) run() {
	if debug.GetEnabled() {
		go t.detectDeadlock()
	}

	t.workerFunc()
	t.markDone()
}

// markDone marks the Task as done.
func (t *Task) markDone() {
	close(t.doneChan)

	t.doneCallback()
}

// detectDeadlock detects if the Task is stuck and prints the stack trace.
func (t *Task) detectDeadlock() {
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
