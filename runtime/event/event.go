package event

import (
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/generics/orderedmap"
	"github.com/iotaledger/hive.go/core/typeutils"
)

// event is the generic base type for all events.
type event[TriggerFunc any] struct {
	// hooks is a dictionary of callbacks that are currently registered with the event.
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

// newEvent creates a new event instance with the given options.
func newEvent[TriggerFunc any](opts ...Option) *event[TriggerFunc] {
	return &event[TriggerFunc]{
		hooks:           orderedmap.New[uint64, *Hook[TriggerFunc]](),
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}
}

// Hook adds a new callback to the event and returns the corresponding Hook.
func (e *event[TriggerFunc]) Hook(triggerFunc TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	hook := newHook(e.hooksCounter.Add(1), e, triggerFunc, opts...)
	e.hooks.Set(hook.id, hook)

	return hook
}

// linkTo hooks the trigger function to the given target event.
func (e *event[TriggerFunc]) linkTo(target eventInterface[TriggerFunc], triggerFunc TriggerFunc) {
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

// eventInterface is an interface that is used to match the generic Hook function of events during the linking process.
type eventInterface[TriggerFunc any] interface {
	Hook(callback TriggerFunc, opts ...Option) *Hook[TriggerFunc]
}
