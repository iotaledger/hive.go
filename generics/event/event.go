package event

import (
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/generics/orderedmap"
)

// Event represents an object that is triggered to notify code of "interesting updates" that may affect its behavior.
type Event[T any] struct {
	beforeCallbacks *orderedmap.OrderedMap[uint64, func(T)]
	callbacks       *orderedmap.OrderedMap[uint64, func(T)]
	afterCallbacks  *orderedmap.OrderedMap[uint64, func(T)]
	asyncCallbacks  *orderedmap.OrderedMap[uint64, func(T)]
}

func New[T any]() (newEvent *Event[T]) {
	return &Event[T]{
		beforeCallbacks: orderedmap.New[uint64, func(T)](),
		callbacks:       orderedmap.New[uint64, func(T)](),
		afterCallbacks:  orderedmap.New[uint64, func(T)](),
		asyncCallbacks:  orderedmap.New[uint64, func(T)](),
	}
}

// AttachSyncBefore allows to register a Closure that is executed before the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) Attach(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.asyncCallbacks, closure, triggerMaxCount...)
}

// AttachSyncBefore allows to register a Closure that is executed before the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) AttachSyncBefore(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.beforeCallbacks, closure, triggerMaxCount...)
}

// AttachSync allows to register a Closure that is executed when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) AttachSync(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.callbacks, closure, triggerMaxCount...)
}

// AttachSyncAfter allows to register a Closure that is executed after the Event triggered.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (e *Event[T]) AttachSyncAfter(closure *Closure[T], triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	e.attachCallback(e.afterCallbacks, closure, triggerMaxCount...)
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
	e.beforeCallbacks.Clear()
	e.callbacks.Clear()
	e.afterCallbacks.Clear()
}

// Trigger calls the registered callbacks with the given parameters.
func (e *Event[T]) Trigger(event T) {
	for _, queue := range []*orderedmap.OrderedMap[uint64, func(T)]{e.beforeCallbacks, e.callbacks, e.afterCallbacks} {
		queue.ForEach(func(closureID uint64, callback func(T)) bool {
			callback(event)

			return true
		})
	}

	e.triggerAsyncCallbacks(event)
}

func (e *Event[T]) triggerAsyncCallbacks(event T) {
	e.asyncCallbacks.ForEach(func(closureID uint64, callback func(T)) bool {
		// TODO: submit jobs to some worker pool
		go callback(event)

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
	e.beforeCallbacks.Delete(closureID)
	e.callbacks.Delete(closureID)
	e.afterCallbacks.Delete(closureID)
}
