package events

import (
	"github.com/iotaledger/hive.go/syncutils"
)

type Event struct {
	triggerFunc     func(handler interface{}, params ...interface{})
	beforeCallbacks map[uint64]interface{}
	callbacks       map[uint64]interface{}
	afterCallbacks  map[uint64]interface{}
	mutex           syncutils.RWMutex
}

// AttachBefore allows to register a Closure that is executed before the Event triggers.
func (ev *Event) AttachBefore(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.beforeCallbacks == nil {
		ev.beforeCallbacks = make(map[uint64]interface{})
	}
	ev.beforeCallbacks[closure.ID] = closure.Fnc
}

func (ev *Event) Attach(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.callbacks == nil {
		ev.callbacks = make(map[uint64]interface{})
	}
	ev.callbacks[closure.ID] = closure.Fnc
}

// AttachAfter allows to register a Closure that is executed after the Event triggered.
func (ev *Event) AttachAfter(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.afterCallbacks == nil {
		ev.afterCallbacks = make(map[uint64]interface{})
	}
	ev.afterCallbacks[closure.ID] = closure.Fnc
}

func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	delete(ev.beforeCallbacks, closure.ID)
	if len(ev.beforeCallbacks) == 0 {
		ev.beforeCallbacks = nil
	}
	delete(ev.callbacks, closure.ID)
	if len(ev.callbacks) == 0 {
		ev.callbacks = nil
	}
	delete(ev.afterCallbacks, closure.ID)
	if len(ev.afterCallbacks) == 0 {
		ev.afterCallbacks = nil
	}
}

func (ev *Event) Trigger(params ...interface{}) {
	ev.mutex.RLock()
	defer ev.mutex.RUnlock()
	if ev.beforeCallbacks != nil {
		for _, handler := range ev.beforeCallbacks {
			ev.triggerFunc(handler, params...)
		}
	}
	if ev.callbacks != nil {
		for _, handler := range ev.callbacks {
			ev.triggerFunc(handler, params...)
		}
	}
	if ev.afterCallbacks != nil {
		for _, handler := range ev.afterCallbacks {
			ev.triggerFunc(handler, params...)
		}
	}
}

func (ev *Event) DetachAll() {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	ev.beforeCallbacks = nil
	ev.callbacks = nil
	ev.afterCallbacks = nil
}

func NewEvent(triggerFunc func(handler interface{}, params ...interface{})) *Event {
	return &Event{
		triggerFunc: triggerFunc,
	}
}
