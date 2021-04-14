package events

import (
	"github.com/iotaledger/hive.go/syncutils"
)

type Event struct {
	triggerFunc     func(handler interface{}, params ...interface{})
	beforeCallbacks map[uintptr]interface{}
	callbacks       map[uintptr]interface{}
	afterCallbacks  map[uintptr]interface{}
	mutex           syncutils.RWMutex
}

func (ev *Event) AttachBefore(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.beforeCallbacks == nil {
		ev.beforeCallbacks = make(map[uintptr]interface{})
	}
	ev.beforeCallbacks[closure.Id] = closure.Fnc
}

func (ev *Event) Attach(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.callbacks == nil {
		ev.callbacks = make(map[uintptr]interface{})
	}
	ev.callbacks[closure.Id] = closure.Fnc
}

func (ev *Event) AttachAfter(closure *Closure) {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	if ev.afterCallbacks == nil {
		ev.afterCallbacks = make(map[uintptr]interface{})
	}
	ev.afterCallbacks[closure.Id] = closure.Fnc
}

func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	delete(ev.beforeCallbacks, closure.Id)
	if len(ev.beforeCallbacks) == 0 {
		ev.beforeCallbacks = nil
	}
	delete(ev.callbacks, closure.Id)
	if len(ev.callbacks) == 0 {
		ev.callbacks = nil
	}
	delete(ev.afterCallbacks, closure.Id)
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
