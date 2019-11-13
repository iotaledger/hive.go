package events

import (
	"github.com/iotaledger/hive.go/syncutils"
)

type Event struct {
	triggerFunc func(handler interface{}, params ...interface{})
	callbacks   map[uintptr]interface{}
	mutex       syncutils.RWMutex
}

func (ev *Event) Attach(closure *Closure) {
	ev.mutex.Lock()
	ev.callbacks[closure.Id] = closure.Fnc
	ev.mutex.Unlock()
}

func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}
	ev.mutex.Lock()
	delete(ev.callbacks, closure.Id)
	ev.mutex.Unlock()
}

func (ev *Event) Trigger(params ...interface{}) {
	ev.mutex.RLock()
	for _, handler := range ev.callbacks {
		ev.triggerFunc(handler, params...)
	}
	ev.mutex.RUnlock()
}

func (ev *Event) DetachAll() {
	ev.mutex.Lock()
	ev.callbacks = make(map[uintptr]interface{})
	ev.mutex.Unlock()
}

func NewEvent(triggerFunc func(handler interface{}, params ...interface{})) *Event {
	return &Event{
		triggerFunc: triggerFunc,
		callbacks:   make(map[uintptr]interface{}),
	}
}
