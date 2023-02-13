package event

import (
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/generics/orderedmap"
	"github.com/iotaledger/hive.go/core/typeutils"
	"github.com/iotaledger/hive.go/core/workerpool"
)

// base is the generic base type for all events.
type base[TriggerFunc any] struct {
	// hooks is a dictionary of all hooks that are currently hooked to the event.
	hooks *orderedmap.OrderedMap[uint64, *Hook[TriggerFunc]]

	// hooksCounter is used to assign a unique ID to each hook.
	hooksCounter atomic.Uint64

	// link is the Hook to another event.
	link *Hook[TriggerFunc]

	// linkMutex is used to prevent concurrent access to the link.
	linkMutex sync.Mutex

	// triggerSettings is the settings that are used to trigger the event.
	*triggerSettings
}

// newBase creates a new base instance.
func newBase[TriggerFunc any](opts ...Option) *base[TriggerFunc] {
	return &base[TriggerFunc]{
		hooks:           orderedmap.New[uint64, *Hook[TriggerFunc]](),
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}
}

// Hook adds a new hook to the event and returns it.
func (e *base[TriggerFunc]) Hook(triggerFunc TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	hookID := e.hooksCounter.Add(1)
	hook := newHook(triggerFunc, func() { e.hooks.Delete(hookID) }, opts...)

	e.hooks.Set(hookID, hook)

	return hook
}

// linkTo links the trigger function to the given target event.
func (e *base[TriggerFunc]) linkTo(target eventInterface[TriggerFunc], triggerFunc TriggerFunc) {
	e.linkMutex.Lock()
	defer e.linkMutex.Unlock()

	if e.link != nil {
		e.link.Unhook()
	}

	if typeutils.IsInterfaceNil(target) {
		e.link = nil
	} else {
		e.link = target.Hook(triggerFunc)
	}
}

// targetWorkerPool returns the worker pool of the given hook or the base's worker pool if the hook does not have one.
func (e *base[TriggerFunc]) targetWorkerPool(hook *Hook[TriggerFunc]) *workerpool.UnboundedWorkerPool {
	if hook.workerPool != nil {
		if hook.workerPool == noWorkerPool {
			return nil
		}

		return hook.workerPool
	}

	return e.workerPool
}

// eventInterface is an interface that is used to match the Hook interface of events for linking.
type eventInterface[TriggerFunc any] interface {
	// Hook adds a new hook to the event and returns it.
	Hook(callback TriggerFunc, opts ...Option) *Hook[TriggerFunc]
}
