package timeutil

import (
	"context"
	"sync"
	"time"
)

// Ticker is task that gets executed repeatedly. It adjusts the intervals or drops ticks to make up for slow executions.
type Ticker struct {
	ctx              context.Context
	ctxCancel        context.CancelFunc
	handler          func()
	interval         time.Duration
	gracefulShutdown sync.WaitGroup
}

// NewTicker creates a new Ticker from the given details. The interval must be greater than zero; if not, NewTicker will
// panic.
func NewTicker(handler func(), interval time.Duration, ctx ...context.Context) (ticker *Ticker) {
	innerCtx := context.Background()
	if len(ctx) > 0 {
		innerCtx = ctx[0]
	}

	tickerCtx, tickerCtxCancel := context.WithCancel(innerCtx)

	ticker = &Ticker{
		ctx:       tickerCtx,
		ctxCancel: tickerCtxCancel,
		handler:   handler,
		interval:  interval,
	}

	go ticker.run()

	return
}

// Shutdown shuts down the Ticker.
func (t *Ticker) Shutdown() {
	t.ctxCancel()
}

// WaitForShutdown waits until the Ticker was shut down.
func (t *Ticker) WaitForShutdown() {
	<-t.ctx.Done()
}

// WaitForGracefulShutdown waits until the Ticker was shut down and the last handler has terminated.
func (t *Ticker) WaitForGracefulShutdown() {
	t.gracefulShutdown.Wait()
}

// run is an internal utility function that executes the ticker logic.
func (t *Ticker) run() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop() // prevent the ticker from leaking

	t.gracefulShutdown.Add(1)
	defer t.gracefulShutdown.Done()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.handler()
		}
	}
}
