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
	// hooks is a dictionary of all hooks that are currently registered with the event.
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

// newBase creates a new base instance with the given options.
func newBase[TriggerFunc any](opts ...Option) *base[TriggerFunc] {
	return &base[TriggerFunc]{
		hooks:           orderedmap.New[uint64, *Hook[TriggerFunc]](),
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}
}

// Hook adds a new callback to the event and returns the corresponding Hook.
func (b *base[TriggerFunc]) Hook(triggerFunc TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	hookID := b.hooksCounter.Add(1)
	hook := newHook(triggerFunc, func() { b.hooks.Delete(hookID) }, opts...)

	b.hooks.Set(hookID, hook)

	return hook
}

// linkTo hooks the trigger function to the given target event.
func (b *base[TriggerFunc]) linkTo(target hookable[TriggerFunc], triggerFunc TriggerFunc) {
	b.linkMutex.Lock()
	defer b.linkMutex.Unlock()

	if b.link != nil {
		b.link.Unhook()
	}

	if typeutils.IsInterfaceNil(target) {
		b.link = nil
	} else {
		b.link = target.Hook(triggerFunc)
	}
}

// targetWorkerPool returns the worker pool of the given hook or the base's worker pool if the hook does not have one.
func (b *base[TriggerFunc]) targetWorkerPool(hook *Hook[TriggerFunc]) *workerpool.UnboundedWorkerPool {
	if hook.workerPool != nil {
		if hook.workerPool == noWorkerPool {
			return nil
		}

		return hook.workerPool
	}

	if b.workerPool == noWorkerPool {
		return nil
	}

	return b.workerPool
}

// hookable is an interface that is used to match the generic Hook function of events during the linking process.
type hookable[TriggerFunc any] interface {
	Hook(callback TriggerFunc, opts ...Option) *Hook[TriggerFunc]
}
