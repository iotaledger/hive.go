package timed

import (
	"sync"
	"time"

	"github.com/izuc/zipp.foundation/ds/shrinkingmap"
)

// region TaskExecutor ////////////////////////////////////////////////////////////////////////////////////////////

// TaskExecutor is a TimedExecutor that internally manages the scheduled callbacks as tasks with a unique
// identifier. It allows to replace existing scheduled tasks and cancel them using the same identifier.
type TaskExecutor[T comparable] struct {
	*Executor
	queuedElements      *shrinkingmap.ShrinkingMap[T, *QueueElement]
	queuedElementsMutex sync.Mutex
}

// NewTaskExecutor is the constructor of the TaskExecutor.
func NewTaskExecutor[T comparable](workerCount int) *TaskExecutor[T] {
	return &TaskExecutor[T]{
		Executor:       NewExecutor(workerCount),
		queuedElements: shrinkingmap.New[T, *QueueElement](),
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

	queuedElement, queuedElementExists := t.queuedElements.Get(identifier)
	if queuedElementExists {
		queuedElement.Cancel()
	}

	t.queuedElements.Set(identifier, t.Executor.ExecuteAt(func() {
		callback()

		t.queuedElementsMutex.Lock()
		defer t.queuedElementsMutex.Unlock()

		t.queuedElements.Delete(identifier)
	}, executionTime))

	queuedElement, _ = t.queuedElements.Get(identifier)
	return queuedElement
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
