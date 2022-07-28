package event

import (
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/core/debug"
	"github.com/iotaledger/hive.go/core/generics/orderedmap"
)

// Event represents an object that is triggered to notify code of "interesting updates" that may affect its behavior.
type Event[T any] struct {
	beforeHooks   *orderedmap.OrderedMap[uint64, func(T)]
	hooks         *orderedmap.OrderedMap[uint64, func(T)]
	afterHooks    *orderedmap.OrderedMap[uint64, func(T)]
	eventHandlers *orderedmap.OrderedMap[uint64, func(T)]
}

// New creates a new Event.
func New[T any]() (newEvent *Event[T]) {
	return &Event[T]{
		beforeHooks:   orderedmap.New[uint64, func(T)](),
		hooks:         orderedmap.New[uint64, func(T)](),
		afterHooks:    orderedmap.New[uint64, func(T)](),
		eventHandlers: orderedmap.New[uint64, func(T)](),
	}
}

// Attach allows to register a Closure that is executed asynchronously when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) Attach(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.eventHandlers, closure, triggerMaxCount...)
}

// HookBefore allows to register a Closure that is executed before the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) HookBefore(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.beforeHooks, closure, triggerMaxCount...)
}

// Hook allows to register a Closure that is executed when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) Hook(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.hooks, closure, triggerMaxCount...)
}

// HookAfter allows to register a Closure that is executed after the Event triggered.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) HookAfter(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.afterHooks, closure, triggerMaxCount...)
}

// Detach allows to unregister a Closure that was previously registered.
func (e *Event[T]) Detach(closure *Closure[T]) {
	if closure == nil {
		return
	}

	e.detachID(closure.ID)
}

// DetachAll removes all registered callbacks.
func (e *Event[T]) DetachAll() {
	e.beforeHooks.Clear()
	e.hooks.Clear()
	e.afterHooks.Clear()
	e.eventHandlers.Clear()
}

// Trigger calls the registered callbacks with the given parameters.
func (e *Event[T]) Trigger(event T) {
	for _, queue := range []*orderedmap.OrderedMap[uint64, func(T)]{e.beforeHooks, e.hooks, e.afterHooks} {
		queue.ForEach(func(closureID uint64, callback func(T)) bool {
			callback(event)

			return true
		})
	}
	e.triggerEventHandlers(event)
}

func (e *Event[T]) triggerEventHandlers(event T) {
	e.eventHandlers.ForEach(func(closureID uint64, callback func(T)) bool {
		var closureStackTrace string
		if debug.GetEnabled() {
			closureStackTrace = debug.ClosureStackTrace(callback)
		}
		// Create a new goroutine to submit the task to the queue to avoid deadlocks.
		// Deadlock could happen when all workers are processing tasks which submit a new task to the queue which is full.
		// Increasing pending task counter allows to successfully wait for all the tasks to be processed.
		Loop.IncreasePendingTasksCounter()
		go func() {
			defer Loop.DecreasePendingTasksCounter()
			Loop.SubmitTask(Loop.CreateTask(func() { callback(event) }, closureStackTrace))
		}()

		return true
	})
}

func (e *Event[T]) attachCallback(callbacks *orderedmap.OrderedMap[uint64, func(T)], closure *Closure[T], triggerMaxCount ...uint64) {
	callbackFunc := closure.Function
	if len(triggerMaxCount) > 0 && triggerMaxCount[0] > 0 {
		triggerCount := atomic.NewUint64(0)

		callbackFunc = func(event T) {
			closure.Function(event)

			if triggerCount.Inc() >= triggerMaxCount[0] {
				e.detachID(closure.ID)
			}
		}
	}

	callbacks.Set(closure.ID, callbackFunc)
}

func (e *Event[T]) detachID(closureID uint64) {
	e.beforeHooks.Delete(closureID)
	e.hooks.Delete(closureID)
	e.afterHooks.Delete(closureID)
	e.eventHandlers.Delete(closureID)
}
