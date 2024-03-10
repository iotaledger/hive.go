package module

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/lo"
)

// OnAllConstructed registers a callback that gets executed when all given modules have been constructed.
func OnAllConstructed(callback func(), modules ...Module) (unsubscribe func()) {
	return onModulesTriggered(Module.ConstructedEvent, callback, modules...)
}

// OnAllInitialized registers a callback that gets executed when all given modules have been initialized.
func OnAllInitialized(callback func(), modules ...Module) (unsubscribe func()) {
	return onModulesTriggered(Module.InitializedEvent, callback, modules...)
}

// OnAllShutdown registers a callback that gets executed when all given modules have been shutdown.
func OnAllShutdown(callback func(), modules ...Module) (unsubscribe func()) {
	return onModulesTriggered(Module.ShutdownEvent, callback, modules...)
}

// OnAllStopped registers a callback that gets executed when all given modules have been stopped.
func OnAllStopped(callback func(), modules ...Module) (unsubscribe func()) {
	return onModulesTriggered(Module.StoppedEvent, callback, modules...)
}

// WaitAll waits until all given modules have triggered the given event.
func WaitAll(event func(Module) reactive.Event, modules ...Module) {
	var wg sync.WaitGroup

	wg.Add(len(modules))
	for _, module := range modules {
		event(module).OnTrigger(wg.Done)
	}

	wg.Wait()
}

// TriggerAll triggers the given event on all given modules.
func TriggerAll(event func(Module) reactive.Event, modules ...Module) {
	for i, module := range modules {
		fmt.Println(i, "TriggerAll", module)
		fmt.Println(i, "TriggerAll", module)
		event(module).Trigger()
	}
}

// onModulesTriggered registers a callback that gets executed when all given modules have triggered the given hook.
func onModulesTriggered(targetEvent func(Module) reactive.Event, callback func(), modulesToWaitFor ...Module) (unsubscribe func()) {
	var (
		expectedTriggerCount = int64(len(modulesToWaitFor))
		actualTriggerCount   atomic.Int64
	)

	unsubscribeFunctions := make([]func(), len(modulesToWaitFor))
	for i, module := range modulesToWaitFor {
		unsubscribeFunctions[i] = targetEvent(module).OnTrigger(func() {
			if actualTriggerCount.Add(1) == expectedTriggerCount {
				callback()
			}
		})
	}

	return lo.Batch(unsubscribeFunctions...)
}
