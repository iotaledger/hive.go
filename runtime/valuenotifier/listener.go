package valuenotifier

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Notifier[T comparable] struct {
	listeners map[T]*listener
	mutex     sync.RWMutex
}

type listener struct {
	channel chan struct{}
	count   int
}

func New[T comparable]() *Notifier[T] {
	return &Notifier[T]{
		listeners: make(map[T]*listener),
	}
}

func (v *Notifier[T]) removeListener(value T) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	valueListeners, exists := v.listeners[value]
	if !exists {
		return
	}
	valueListeners.count--

	if valueListeners.count == 0 {
		// No one is listening anymore, so we can close the channel and clean up
		close(valueListeners.channel)
		delete(v.listeners, value)
	}
}

// Listener creates a unique listener that can be used to wait until Notify is called for the given value.
func (v *Notifier[T]) Listener(value T) *Listener {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if valueListeners, exists := v.listeners[value]; exists {
		valueListeners.count++
		return newListener(valueListeners.channel, func() {
			v.removeListener(value)
		})
	}

	msgProcessedChan := make(chan struct{})
	v.listeners[value] = &listener{msgProcessedChan, 1}

	return newListener(msgProcessedChan, func() {
		v.removeListener(value)
	})
}

func (v *Notifier[T]) Notify(value T) {
	v.mutex.RLock()
	// check if the key was registered
	if _, exists := v.listeners[value]; !exists {
		v.mutex.RUnlock()

		return
	}
	v.mutex.RUnlock()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	// check again if the key is still registered
	l, exists := v.listeners[value]
	if !exists {
		return
	}

	// trigger the event by closing the channel
	close(l.channel)

	delete(v.listeners, value)
}

var (
	ErrListenerDeregistered = errors.New("listener was deregistered")
)

type Listener struct {
	channel          chan struct{}
	deregisteredChan chan struct{}

	deregistered atomic.Bool
	deregister   func()
}

func newListener(channel chan struct{}, deregister func()) *Listener {
	return &Listener{
		channel:          channel,
		deregisteredChan: make(chan struct{}),
		deregister:       deregister,
	}
}

func (l *Listener) Deregister() {
	if !l.deregistered.Swap(true) {
		close(l.deregisteredChan)
		l.deregister()
	}
}

// Wait waits until the listener is notified or the context is done.
// If the context was done, the listener de-registers automatically to clean up memory.
func (l *Listener) Wait(ctx context.Context) error {
	if l.deregistered.Load() {
		return ErrListenerDeregistered
	}

	// we wait either until the channel got closed or the context is done
	select {
	case <-l.channel:
		return nil
	case <-l.deregisteredChan:
		return ErrListenerDeregistered
	case <-ctx.Done():
		l.deregister()
		return ctx.Err()
	}
}
