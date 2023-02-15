package event

import (
	"sync/atomic"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/workerpool"
)

// WithMaxTriggerCount sets the maximum number of times an event (or hook) shall be triggered.
func WithMaxTriggerCount(maxTriggerCount uint64) Option {
	return func(triggerSettings *triggerSettings) {
		triggerSettings.maxTriggerCount = maxTriggerCount
	}
}

// WithWorkerPool sets the worker pool that shall be used to execute the triggered function.
func WithWorkerPool(workerPool *workerpool.UnboundedWorkerPool) Option {
	return func(triggerSettings *triggerSettings) {
		triggerSettings.workerPool = workerPool
	}
}

// WithoutWorkerPool disables the usage of worker pools for the triggered element.
func WithoutWorkerPool() Option {
	return func(triggerSettings *triggerSettings) {
		triggerSettings.workerPool = noWorkerPool
	}
}

// triggerSettings is a struct that contains trigger related settings and logic.
type triggerSettings struct {
	workerPool      *workerpool.UnboundedWorkerPool
	triggerCount    atomic.Uint64
	maxTriggerCount uint64
}

// WasTriggered returns true if Trigger was called at least once.
func (t *triggerSettings) WasTriggered() bool {
	return t.triggerCount.Load() > 0
}

// TriggerCount returns the number of times Trigger was called.
func (t *triggerSettings) TriggerCount() int {
	return int(t.triggerCount.Load())
}

// MaxTriggerCount returns the maximum number of times Trigger can be called.
func (t *triggerSettings) MaxTriggerCount() int {
	return int(t.maxTriggerCount)
}

// MaxTriggerCountReached returns true if the maximum number of times Trigger can be called was reached.
func (t *triggerSettings) MaxTriggerCountReached() bool {
	return t.maxTriggerCount != 0 && t.triggerCount.Load() > t.maxTriggerCount
}

// WorkerPool returns the worker pool that shall be used to execute the triggered function.
func (t *triggerSettings) WorkerPool() *workerpool.UnboundedWorkerPool {
	return lo.Return2(t.hasWorkerPool())
}

// hasWorkerPool returns true if a worker pool is set
func (t *triggerSettings) hasWorkerPool() (bool, *workerpool.UnboundedWorkerPool) {
	if t.workerPool == noWorkerPool {
		return true, nil
	}

	return t.workerPool != nil, t.workerPool
}

// currentTriggerExceedsMaxTriggerCount returns true if the maximum number of times Trigger shall be called was reached.
func (t *triggerSettings) currentTriggerExceedsMaxTriggerCount() bool {
	return t.triggerCount.Add(1) > t.maxTriggerCount && t.maxTriggerCount != 0
}

// noWorkerPool is a special value that indicates that no worker pool shall be used.
var noWorkerPool = &workerpool.UnboundedWorkerPool{}

// Option is a function that configures the triggerSettings.
type Option = options.Option[triggerSettings]
