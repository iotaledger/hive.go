package promise

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

// region Event

// Event is an event that can be triggered exactly once. Consumers that register themselves after the event was
// triggered, will be called immediately.
type Event struct {
	// callbacks contains the callbacks that will be called when the event is triggered.
	callbacks *shrinkingmap.ShrinkingMap[uniqueID, func()]

	// callbackIDs is a counter for the callback IDs.
	callbackIDs uniqueID

	// mutex is used to synchronize access to the callbacks.
	mutex sync.RWMutex
}

// NewEvent creates a new Event with 0 generic parameters.
func NewEvent() *Event {
	return &Event{
		callbacks: shrinkingmap.New[uniqueID, func()](),
	}
}

// Trigger triggers the event. If the event was already triggered, this method does nothing.
func (e *Event) Trigger() (wasTriggered bool) {
	for _, callback := range func() []func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		callbacks := e.callbacks
		if wasTriggered = callbacks != nil; !wasTriggered {
			return nil
		}

		e.callbacks = nil

		return callbacks.Values()
	}() {
		callback()
	}

	return
}

// OnTrigger registers a callback that will be called when the event is triggered. If the event was already triggered,
// the callback will be called immediately.
func (e *Event) OnTrigger(callback func()) (unsubscribe func()) {
	registerCallback := func() (unsubscribe func(), subscribed bool) {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		if e.callbacks == nil {
			return void, false
		}

		callbackID := e.callbackIDs.Next()

		e.callbacks.Set(callbackID, callback)

		return func() {
			e.mutex.Lock()
			defer e.mutex.Unlock()

			if e.callbacks != nil {
				e.callbacks.Delete(callbackID)
			}
		}, true
	}

	unsubscribe, subscribed := registerCallback()
	if !subscribed {
		callback()
	}

	return unsubscribe
}

// WasTriggered returns true if the event was already triggered.
func (e *Event) WasTriggered() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.callbacks == nil
}

// endregion

// region Event1

// Event1 is an event with a single parameter that can be triggered exactly once. Consumers that register themselves
// after the event was triggered, will be called immediately.
type Event1[T any] struct {
	// callbacks contains the callbacks that will be called when the event is triggered.
	callbacks *shrinkingmap.ShrinkingMap[uniqueID, func(T)]

	// callbackIDs is a counter for the callback IDs.
	callbackIDs uniqueID

	// value is the value that was passed to the Trigger method.
	value *T

	// mutex is used to synchronize access to the callbacks.
	mutex sync.RWMutex
}

// NewEvent1 creates a new event with 1 generic parameter.
func NewEvent1[T any]() *Event1[T] {
	return &Event1[T]{
		callbacks: shrinkingmap.New[uniqueID, func(T)](),
	}
}

// Trigger triggers the event. If the event was already triggered, this method does nothing.
func (e *Event1[T]) Trigger(arg T) (wasTriggered bool) {
	for _, callback := range func() []func(T) {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		callbacks := e.callbacks
		if wasTriggered = callbacks != nil; !wasTriggered {
			return nil
		}

		e.callbacks = nil
		e.value = &arg

		return callbacks.Values()
	}() {
		callback(arg)
	}

	return
}

// OnTrigger registers a callback that will be called when the event is triggered. If the event was already triggered,
// the callback will be called immediately.
func (e *Event1[T]) OnTrigger(callback func(T)) (unsubscribe func()) {
	registerCallback := func() (unsubscribe func(), subscribed bool) {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		if e.callbacks == nil {
			return void, false
		}

		callbackID := e.callbackIDs.Next()

		e.callbacks.Set(callbackID, callback)

		return func() {
			e.mutex.Lock()
			defer e.mutex.Unlock()

			if e.callbacks != nil {
				e.callbacks.Delete(callbackID)
			}
		}, true
	}

	unsubscribe, subscribed := registerCallback()
	if !subscribed {
		callback(*e.value)
	}

	return unsubscribe
}

// WasTriggered returns true if the event was already triggered.
func (e *Event1[T]) WasTriggered() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.callbacks == nil
}

// endregion
