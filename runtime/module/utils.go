package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
)

// TriggerAll triggers the given event on all given modules.
func TriggerAll(event func(Module) reactive.Event, modules ...Module) {
	for _, module := range modules {
		event(module).Trigger()
	}
}

// WaitAll waits until all given modules have triggered the given event.
func WaitAll(event func(Module) reactive.Event, modules ...Module) reactive.WaitGroup[Module] {
	wg := reactive.NewWaitGroup(modules...)
	for _, module := range modules {
		event(module).OnTrigger(func() {
			wg.Done(module)
		})
	}

	return wg
}

// InitSimpleLifecycle is a helper function that sets up the default lifecycle for the given module. If no shutdown
// function is provided, then the default shutdown function will be used, which automatically triggers the StoppedEvent.
//
// Note: If a more complex lifecycle is required, then it needs to be implemented manually.
func InitSimpleLifecycle[T Module](m T, optShutdown ...func(T)) T {
	if len(optShutdown) == 0 {
		m.ShutdownEvent().OnTrigger(func() {
			m.StoppedEvent().Trigger()
		})
	} else {
		m.ShutdownEvent().OnTrigger(func() {
			optShutdown[0](m)
		})
	}

	m.ConstructedEvent().Trigger()
	m.InitializedEvent().Trigger()

	return m
}
