package module

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/reactive"
)

// WaitAll waits until all given modules have triggered the given event.
func WaitAll(event func(Module) reactive.Event, modules ...Module) reactive.WaitGroup[T] {
	pendingModules := reportPendingModules(modules...)

	var wg sync.WaitGroup

	wg.Add(len(modules))
	for _, module := range modules {
		event(module).OnTrigger(func() {
			if pendingModules != nil {
				pendingModules.Delete(module)
			}

			wg.Done()
		})
	}

	wg.Wait()
}
