package events

import (
	"github.com/iotaledger/hive.go/v2/datastructure/orderedmap"
)

// Event represents an object that is triggered to notify code of "interesting updates" that may affect its behavior.
type Event struct {
	triggerFunc     func(handler interface{}, params ...interface{})
	beforeCallbacks *orderedmap.OrderedMap
	callbacks       *orderedmap.OrderedMap
	afterCallbacks  *orderedmap.OrderedMap
}

// NewEvent is the constructor of an Event.
func NewEvent(triggerFunc func(handler interface{}, params ...interface{})) *Event {
	return &Event{
		triggerFunc:     triggerFunc,
		beforeCallbacks: orderedmap.New(),
		callbacks:       orderedmap.New(),
		afterCallbacks:  orderedmap.New(),
	}
}

// AttachBefore allows to register a Closure that is executed before the Event triggers.
func (ev *Event) AttachBefore(closure *Closure) {
	if closure == nil {
		return
	}

	ev.beforeCallbacks.Set(closure.ID, closure.Fnc)
}

// Attach allows to register a Closure that is executed when the Event triggers.
func (ev *Event) Attach(closure *Closure) {
	if closure == nil {
		return
	}

	ev.callbacks.Set(closure.ID, closure.Fnc)
}

// AttachAfter allows to register a Closure that is executed after the Event triggered.
func (ev *Event) AttachAfter(closure *Closure) {
	if closure == nil {
		return
	}

	ev.afterCallbacks.Set(closure.ID, closure.Fnc)
}

// Detach allows to unregister a Closure that was previously registered.
func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}

	ev.beforeCallbacks.Delete(closure.ID)
	ev.callbacks.Delete(closure.ID)
	ev.afterCallbacks.Delete(closure.ID)
}

// Trigger calls the registered callbacks with the given parameters.
func (ev *Event) Trigger(params ...interface{}) {
	ev.beforeCallbacks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
	ev.callbacks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
	ev.afterCallbacks.ForEach(func(_, handler interface{}) bool {
		ev.triggerFunc(handler, params...)

		return true
	})
}

// DetachAll removes all registered callbacks.
func (ev *Event) DetachAll() {
	ev.beforeCallbacks.Clear()
	ev.callbacks.Clear()
	ev.afterCallbacks.Clear()
}
