package event

import (
	"github.com/iotaledger/hive.go/core/generics/options"

	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// Hook is a container that holds a trigger function and its trigger settings.
type Hook[TriggerFunc any] struct {
	id      uint64
	event   *event[TriggerFunc]
	trigger TriggerFunc

	*triggerSettings
}

// newHook creates a new Hook.
func newHook[TriggerFunc any](id uint64, event *event[TriggerFunc], trigger TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	return &Hook[TriggerFunc]{
		id:              id,
		event:           event,
		trigger:         trigger,
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}
}

// WorkerPool returns the worker pool that is used to trigger the callback.
func (h *Hook[TriggerFunc]) WorkerPool() *workerpool.WorkerPool {
	if has, workerPool := h.hasWorkerPool(); has {
		return workerPool
	}

	return h.event.WorkerPool()
}

// Unhook removes the callback from the event.
func (h *Hook[TriggerFunc]) Unhook() {
	h.event.hooks.Delete(h.id)
}
