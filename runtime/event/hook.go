package event

import (
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// Hook is a container that holds a trigger function and its trigger settings.
type Hook[TriggerFunc any] struct {
	id      uint64
	event   *event[TriggerFunc]
	trigger TriggerFunc

	// preTriggerFunc is a function that is called before each Hook is executed.
	preTriggerFunc TriggerFunc

	*triggerSettings
}

// newHook creates a new Hook.
func newHook[TriggerFunc any](id uint64, event *event[TriggerFunc], trigger TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	h := &Hook[TriggerFunc]{
		id:              id,
		event:           event,
		trigger:         trigger,
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}

	if !IsInterfaceNil(h.triggerSettings.preTriggerFunc) {
		//nolint:forcetypeassert // false positive, we know that preTriggerFunc is of type TriggerFunc
		h.preTriggerFunc = h.triggerSettings.preTriggerFunc.(TriggerFunc)
	}

	return h
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
