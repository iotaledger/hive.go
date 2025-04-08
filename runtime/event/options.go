package event

import (
	"sync/atomic"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// WithMaxTriggerCount sets the maximum number of times an entity shall be triggered.
func WithMaxTriggerCount(maxTriggerCount uint64) Option {
	return func(triggerSettings *triggerSettings) {
		triggerSettings.maxTriggerCount = maxTriggerCount
	}
}

// WithWorkerPool sets the worker pool that is used to process the trigger (nil forces execution in-place).
func WithWorkerPool(workerPool *workerpool.WorkerPool) Option {
	if workerPool == nil {
		return func(triggerSettings *triggerSettings) {
			triggerSettings.workerPool = noWorkerPool
		}
	}

	return func(triggerSettings *triggerSettings) {
		triggerSettings.workerPool = workerPool
	}
}

// WithPreTriggerFunc sets a function that is synchronously called before the trigger is executed.
func WithPreTriggerFunc(preTriggerFunc any) Option {
	return func(triggerSettings *triggerSettings) {
		triggerSettings.preTriggerFunc = preTriggerFunc
	}
}

// triggerSettings is a struct that contains trigger related settings and logic.
type triggerSettings struct {
	workerPool      *workerpool.WorkerPool
	triggerCount    atomic.Uint64
	maxTriggerCount uint64
	preTriggerFunc  any
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
func (t *triggerSettings) WorkerPool() *workerpool.WorkerPool {
	return lo.Return2(t.hasWorkerPool())
}

// hasWorkerPool returns if a worker pool (and which one) is set.
func (t *triggerSettings) hasWorkerPool() (bool, *workerpool.WorkerPool) {
	if t.workerPool == noWorkerPool {
		return true, nil
	}

	return t.workerPool != nil, t.workerPool
}

// currentTriggerExceedsMaxTriggerCount returns true if the maximum number of times Trigger shall be called was reached.
func (t *triggerSettings) currentTriggerExceedsMaxTriggerCount() bool {
	return t.triggerCount.Add(1) > t.maxTriggerCount && t.maxTriggerCount != 0
}

// noWorkerPool is a special value that indicates that no worker pool shall be used (forced).
var noWorkerPool = &workerpool.WorkerPool{}

// Option is a function that configures the triggerSettings.
type Option = options.Option[triggerSettings]
