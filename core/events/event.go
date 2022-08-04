package events

import (
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/core/datastructure/orderedmap"
	reflectpkg "github.com/iotaledger/hive.go/core/reflect"
)

// Event represents an object that is triggered to notify code of "interesting updates" that may affect its behavior.
type Event struct {
	triggerFunc func(handler interface{}, params ...interface{})
	beforeHooks *orderedmap.OrderedMap
	hooks       *orderedmap.OrderedMap
	afterHooks  *orderedmap.OrderedMap
}

// NewEvent is the constructor of an Event.
func NewEvent(triggerFunc func(handler interface{}, params ...interface{})) *Event {
	return &Event{
		triggerFunc: triggerFunc,
		beforeHooks: orderedmap.New(),
		hooks:       orderedmap.New(),
		afterHooks:  orderedmap.New(),
	}
}

func (ev *Event) attachCallback(callbacks *orderedmap.OrderedMap, closure *Closure, triggerMaxCount ...uint64) {

	callbackFunc := closure.Fnc

	if (len(triggerMaxCount) > 0) && (triggerMaxCount[0] > 0) {
		// a trigger limit was specified
		triggerCount := atomic.NewUint64(0)

		// wrap the Closure Function to automatically detach the Closure from the event after exceeding the trigger limit.
		callbackFunc = reflectpkg.FuncPostCallback(closure.Fnc, func() {
			if triggerCount.Inc() >= triggerMaxCount[0] {
				ev.DetachID(closure.ID)
			}
		})
	}

	callbacks.Set(closure.ID, callbackFunc)
}

// HookBefore allows to register a Closure that is executed before the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (ev *Event) HookBefore(closure *Closure, triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	ev.attachCallback(ev.beforeHooks, closure, triggerMaxCount...)
}

// Hook allows to register a Closure that is executed when the Event triggers.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (ev *Event) Hook(closure *Closure, triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	ev.attachCallback(ev.hooks, closure, triggerMaxCount...)
}

// HookAfter allows to register a Closure that is executed after the Event triggered.
// If 'triggerMaxCount' is >0, the Closure is automatically detached after exceeding the trigger limit.
func (ev *Event) HookAfter(closure *Closure, triggerMaxCount ...uint64) {
	if closure == nil {
		return
	}

	ev.attachCallback(ev.afterHooks, closure, triggerMaxCount...)
}

// DetachID allows to unregister a Closure ID that was previously registered.
func (ev *Event) DetachID(closureID uint64) {
	ev.beforeHooks.Delete(closureID)
	ev.hooks.Delete(closureID)
	ev.afterHooks.Delete(closureID)
}

// Detach allows to unregister a Closure that was previously registered.
func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}

	ev.DetachID(closure.ID)
}

// Trigger calls the registered callbacks with the given parameters.
func (ev *Event) Trigger(params ...interface{}) {
	ev.beforeHooks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
	ev.hooks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
	ev.afterHooks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
}

// DetachAll removes all registered callbacks.
func (ev *Event) DetachAll() {
	ev.beforeHooks.Clear()
	ev.hooks.Clear()
	ev.afterHooks.Clear()
}
