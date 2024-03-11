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
