package timed

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/runtime/options"
)

// region TaskExecutor ////////////////////////////////////////////////////////////////////////////////////////////

// TaskExecutor is a TimedExecutor that internally manages the scheduled callbacks as tasks with a unique
// identifier. It allows to replace existing scheduled tasks and cancel them using the same identifier.
type TaskExecutor[T comparable] struct {
	*Executor
	queuedElements      *shrinkingmap.ShrinkingMap[T, *QueueElement[func()]]
	queuedElementsMutex sync.Mutex
}

// NewTaskExecutor is the constructor of the TaskExecutor.
func NewTaskExecutor[T comparable](workerCount int, opts ...options.Option[Executor]) *TaskExecutor[T] {
	return &TaskExecutor[T]{
		Executor:       NewExecutor(workerCount, opts...),
		queuedElements: shrinkingmap.New[T, *QueueElement[func()]](),
	}
}

// ExecuteAfter executes the given function after the given delay.
func (t *TaskExecutor[T]) ExecuteAfter(identifier T, callback func(), delay time.Duration) *ScheduledTask {
	return t.ExecuteAt(identifier, callback, time.Now().Add(delay))
}

// ExecuteAt executes the given function at the given time.
func (t *TaskExecutor[T]) ExecuteAt(identifier T, callback func(), executionTime time.Time) *ScheduledTask {
	t.queuedElementsMutex.Lock()
	defer t.queuedElementsMutex.Unlock()

	if queuedElement, queuedElementExists := t.queuedElements.Get(identifier); queuedElementExists {
		queuedElement.Cancel()
	}

	scheduledTask := t.Executor.ExecuteAt(func() {
		callback()

		t.queuedElementsMutex.Lock()
		defer t.queuedElementsMutex.Unlock()

		t.queuedElements.Delete(identifier)
	}, executionTime)

	if scheduledTask != nil {
		t.queuedElements.Set(identifier, scheduledTask)
	}

	return scheduledTask
}

// Cancel cancels a queued task.
func (t *TaskExecutor[T]) Cancel(identifier T) (canceled bool) {
	t.queuedElementsMutex.Lock()
	defer t.queuedElementsMutex.Unlock()

	queuedElement, queuedElementExists := t.queuedElements.Get(identifier)
	if !queuedElementExists {
		return false
	}

	queuedElement.Cancel()
	t.queuedElements.Delete(identifier)

	return true
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
