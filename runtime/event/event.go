package event

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/runtime/options"
)

// event is the base type for all generic events.
type event[TriggerFunc any] struct {
	// hooks is a dictionary of callbacks that are currently registered with the event.
	hooks *orderedmap.OrderedMap[uint64, *Hook[TriggerFunc]]

	// hooksCounter is used to assign a unique ID to each hook.
	hooksCounter atomic.Uint64

	// link is the Hook to another event.
	link *Hook[TriggerFunc]

	// linkMutex is used to prevent concurrent access to the link.
	linkMutex sync.Mutex

	// preTriggerFunc is a function that is called before each Hook is executed.
	preTriggerFunc TriggerFunc

	// triggerSettings are the settings that keep track of the number of times the event was triggered.
	*triggerSettings
}

// newEvent creates a new event instance with the given options.
func newEvent[TriggerFunc any](opts ...Option) *event[TriggerFunc] {
	e := &event[TriggerFunc]{
		hooks:           orderedmap.New[uint64, *Hook[TriggerFunc]](),
		triggerSettings: options.Apply(new(triggerSettings), opts),
	}

	if !IsInterfaceNil(e.triggerSettings.preTriggerFunc) {
		//nolint:forcetypeassert // false positive, we know that preTriggerFunc is of type TriggerFunc
		e.preTriggerFunc = e.triggerSettings.preTriggerFunc.(TriggerFunc)
	}

	return e
}

// Hook adds a new callback to the event and returns the corresponding Hook.
func (e *event[TriggerFunc]) Hook(triggerFunc TriggerFunc, opts ...Option) *Hook[TriggerFunc] {
	hook := newHook(e.hooksCounter.Add(1), e, triggerFunc, opts...)
	e.hooks.Set(hook.id, hook)

	return hook
}

// linkTo unhooks the previous Hook and then hooks the trigger function to the given target event.
func (e *event[TriggerFunc]) linkTo(target eventInterface[TriggerFunc], triggerFunc TriggerFunc) {
	e.linkMutex.Lock()
	defer e.linkMutex.Unlock()

	if e.link != nil {
		e.link.Unhook()
	}

	if IsInterfaceNil(target) {
		e.link = nil
	} else {
		e.link = target.Hook(triggerFunc)
	}
}

func IsInterfaceNil(param interface{}) bool {
	return param == nil || (*[2]uintptr)(unsafe.Pointer(&param))[1] == 0
}

// eventInterface is an interface that is used to match the generic Hook function of events during the linking process.
type eventInterface[TriggerFunc any] interface {
	Hook(callback TriggerFunc, opts ...Option) *Hook[TriggerFunc]
}
