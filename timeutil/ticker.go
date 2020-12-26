package timeutil

import (
	"sync"
	"time"
)

// Ticker is task that gets executed repeatedly. It adjusts the intervals or drops ticks to make up for slow executions.
type Ticker struct {
	handler                func()
	interval               time.Duration
	internalShutdownSignal chan struct{}
	externalShutdownSignal <-chan struct{}
	handlerDone            sync.WaitGroup
	shutdownOnce           sync.Once
}

// NewTicker creates a new Ticker from the given details. The interval must be greater than zero; if not, NewTicker will
// panic.
func NewTicker(handler func(), interval time.Duration, optionalExternalShutdownSignal ...<-chan struct{}) (ticker *Ticker) {
	ticker = &Ticker{
		handler:                handler,
		interval:               interval,
		internalShutdownSignal: make(chan struct{}, 1),
	}

	if len(optionalExternalShutdownSignal) >= 1 && optionalExternalShutdownSignal[0] != nil {
		ticker.externalShutdownSignal = optionalExternalShutdownSignal[0]
	} else {
		ticker.externalShutdownSignal = ticker.internalShutdownSignal
	}

	go ticker.run()

	return
}

// Shutdown shuts down the Ticker.
func (t *Ticker) Shutdown() {
	t.shutdownOnce.Do(func() {
		close(t.internalShutdownSignal)
	})
}

// WaitForShutdown waits until the Ticker was shut down.
func (t *Ticker) WaitForShutdown() {
	<-t.internalShutdownSignal
	return
}

// WaitForGraceFullShutdown waits until the Ticker was shut down and the last handler has terminated.
func (t *Ticker) WaitForGraceFullShutdown() {
	<-t.internalShutdownSignal
	t.handlerDone.Wait()
	return
}

// run is an internal utility function that executes the ticker logic.
func (t *Ticker) run() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop() // prevent the ticker from leaking

	t.handlerDone.Add(1)
	defer t.handlerDone.Done()

	for {
		select {
		case <-t.externalShutdownSignal:
			if t.externalShutdownSignal != t.internalShutdownSignal {
				t.Shutdown()
			}
			return
		case <-t.internalShutdownSignal:
			return
		case <-ticker.C:
			t.handler()
		}
	}
}
