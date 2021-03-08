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
	defer ev.mutex.Unlock()
	ev.callbacks[closure.Id] = closure.Fnc
}

func (ev *Event) Detach(closure *Closure) {
	if closure == nil {
		return
	}
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	delete(ev.callbacks, closure.Id)
}

func (ev *Event) Trigger(params ...interface{}) {
	ev.mutex.RLock()
	defer ev.mutex.RUnlock()
	for _, handler := range ev.callbacks {
		ev.triggerFunc(handler, params...)
	}
}

func (ev *Event) DetachAll() {
	ev.mutex.Lock()
	defer ev.mutex.Unlock()
	ev.callbacks = make(map[uintptr]interface{})
}

func NewEvent(triggerFunc func(handler interface{}, params ...interface{})) *Event {
	return &Event{
		triggerFunc: triggerFunc,
		callbacks:   make(map[uintptr]interface{}),
	}
}
