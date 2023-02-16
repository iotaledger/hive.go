package timeutil

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/types"
)

// region PrecisionTicker //////////////////////////////////////////////////////////////////////////////////////////////

// PrecisionTicker is a ticker that can be used on systems (like windows) that do not offer a high enough time
// resolution for very fast ticker intervals.
type PrecisionTicker struct {
	callback   func()
	rate       time.Duration
	iterations int
	options    *precisionTickerOptions

	iterationsMutex sync.RWMutex
	shutdownWG      sync.WaitGroup
	shutdownChan    chan types.Empty
	shutdownOnce    sync.Once
}

// NewPrecisionTicker creates a new PrecisionTicker instance that executes the given callback function at the given
// rate.
func NewPrecisionTicker(callback func(), rate time.Duration, options ...PrecisionTickerOption) (precisionTicker *PrecisionTicker) {
	precisionTicker = &PrecisionTicker{
		callback:     callback,
		rate:         rate,
		options:      newPrecisionTickerOptions(options...),
		shutdownChan: make(chan types.Empty, 1),
	}

	precisionTicker.shutdownWG.Add(1)
	go precisionTicker.run()

	return precisionTicker
}

// Iterations returns the number of iterations that the ticker has performed.
func (p *PrecisionTicker) Iterations() (iterations int) {
	p.iterationsMutex.RLock()
	defer p.iterationsMutex.RUnlock()

	return p.iterations
}

// WaitForShutdown waits for the ticker to shut down.
func (p *PrecisionTicker) WaitForShutdown() {
	<-p.shutdownChan
}

// WaitForGracefulShutdown waits for the ticker to shut down gracefully.
func (p *PrecisionTicker) WaitForGracefulShutdown() {
	p.shutdownWG.Wait()
}

// Shutdown shuts down the ticker.
func (p *PrecisionTicker) Shutdown() {
	p.shutdownOnce.Do(func() {
		close(p.shutdownChan)
	})
}

// run is the main loop of the ticker.
func (p *PrecisionTicker) run() {
	defer p.shutdownWG.Done()
	defer p.Shutdown()

	start := time.Now()
	for p.options.maxIterations == 0 || p.iterations < p.options.maxIterations {
		select {
		case <-p.shutdownChan:
			return
		default:
			p.callback()

			p.waitIfNecessary(start)
		}
	}
}

// waitIfNecessary waits for the next tick if the observed execution time is smaller than the expected execution time by
// a margin that is larger than the minimum time precision.
func (p *PrecisionTicker) waitIfNecessary(start time.Time) {
	if tickerOffset := time.Until(start.Add(time.Duration(p.increaseIterations()) * p.rate)); tickerOffset > p.options.minTimePrecision {
		time.Sleep(tickerOffset)
	}
}

// incrementIterations increases the number of iterations by one. It returns the number of iterations after the
// increment.
func (p *PrecisionTicker) increaseIterations() (newIterations int) {
	p.iterationsMutex.Lock()
	defer p.iterationsMutex.Unlock()

	p.iterations++

	return p.iterations
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region PrecisionTickerOption ////////////////////////////////////////////////////////////////////////////////////////

// PrecisionTickerOption is a function that can be used to configure a PrecisionTicker.
type PrecisionTickerOption func(*precisionTickerOptions)

// WithMinTimePrecision sets the assumed minimum time precision of the system.
func WithMinTimePrecision(minTimePrecision time.Duration) PrecisionTickerOption {
	return func(options *precisionTickerOptions) {
		options.minTimePrecision = minTimePrecision
	}
}

// WithMaxIterations sets the maximum number of iterations that the ticker will perform.
func WithMaxIterations(maxIterations int) PrecisionTickerOption {
	return func(options *precisionTickerOptions) {
		options.maxIterations = maxIterations
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region precisionTickerOptions ///////////////////////////////////////////////////////////////////////////////////////

// precisionTickerOptions is a struct that contains the options for a PrecisionTicker.
type precisionTickerOptions struct {
	minTimePrecision time.Duration
	maxIterations    int
}

// newPrecisionTickerOptions creates a new precisionTickerOptions instance.
func newPrecisionTickerOptions(options ...PrecisionTickerOption) (newPrecisionTickerOptions *precisionTickerOptions) {
	newPrecisionTickerOptions = &precisionTickerOptions{
		minTimePrecision: 16 * time.Millisecond,
		maxIterations:    0,
	}

	for _, option := range options {
		option(newPrecisionTickerOptions)
	}

	return newPrecisionTickerOptions
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
