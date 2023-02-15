package event

import (
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/core/debug"
	"github.com/iotaledger/hive.go/core/generics/orderedmap"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// Event represents an object that is triggered to notify code of "interesting updates" that may affect its behavior.
type Event[T any] struct {
	beforeHooks   *orderedmap.OrderedMap[uint64, func(T)]
	hooks         *orderedmap.OrderedMap[uint64, func(T)]
	afterHooks    *orderedmap.OrderedMap[uint64, func(T)]
	asyncHandlers *orderedmap.OrderedMap[uint64, *handler[T]]
}

// New creates a new Event.
func New[T any]() (newEvent *Event[T]) {
	return &Event[T]{
		beforeHooks:   orderedmap.New[uint64, func(T)](),
		hooks:         orderedmap.New[uint64, func(T)](),
		afterHooks:    orderedmap.New[uint64, func(T)](),
		asyncHandlers: orderedmap.New[uint64, *handler[T]](),
	}
}

// Attach allows to register a Closure that is executed asynchronously when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) Attach(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	// By default, we use the global worker pool.
	e.asyncHandlers.Set(closure.ID, newHandler[T](e.callbackFromClosure(closure, triggerMaxCount...), Loop))
}

// AttachWithWorkerPool allows to register a Closure that is executed asynchronously in the specified worker pool when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) AttachWithWorkerPool(closure *Closure[T], wp *workerpool.UnboundedWorkerPool, triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.asyncHandlers.Set(closure.ID, newHandler[T](e.callbackFromClosure(closure, triggerMaxCount...), wp))
}

// HookBefore allows to register a Closure that is executed before the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) HookBefore(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.beforeHooks.Set(closure.ID, e.callbackFromClosure(closure, triggerMaxCount...))
}

// Hook allows to register a Closure that is executed when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) Hook(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.hooks.Set(closure.ID, e.callbackFromClosure(closure, triggerMaxCount...))
}

// HookAfter allows to register a Closure that is executed after the Event triggered.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) HookAfter(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.afterHooks.Set(closure.ID, e.callbackFromClosure(closure, triggerMaxCount...))
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
	e.asyncHandlers.Clear()
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
	e.asyncHandlers.ForEach(func(closureID uint64, h *handler[T]) bool {
		var closureStackTrace string
		if debug.GetEnabled() {
			closureStackTrace = debug.ClosureStackTrace(h.callback)
		}

		h.wp.Submit(func() { h.callback(event) }, closureStackTrace)

		return true
	})
}

func (e *Event[T]) callbackFromClosure(closure *Closure[T], triggerMaxCount ...uint64) func(T) {
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

	return callbackFunc
}

func (e *Event[T]) detachID(closureID uint64) {
	e.beforeHooks.Delete(closureID)
	e.hooks.Delete(closureID)
	e.afterHooks.Delete(closureID)
	e.asyncHandlers.Delete(closureID)
}
