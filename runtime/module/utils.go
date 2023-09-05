package module

import (
	"sync/atomic"

	"github.com/iotaledger/hive.go/lo"
)

// OnAllConstructed registers a callback that gets executed when all given modules have been constructed.
func OnAllConstructed(callback func(), modules ...Interface) (unsubscribe func()) {
	return onModulesTriggered(Interface.HookConstructed, callback, modules...)
}

// OnAllInitialized registers a callback that gets executed when all given modules have been initialized.
func OnAllInitialized(callback func(), modulesToWaitFor ...Interface) (unsubscribe func()) {
	return onModulesTriggered(Interface.HookInitialized, callback, modulesToWaitFor...)
}

// OnAllShutdown registers a callback that gets executed when all given modules have been shutdown.
func OnAllShutdown(callback func(), modulesToWaitFor ...Interface) (unsubscribe func()) {
	return onModulesTriggered(Interface.HookShutdown, callback, modulesToWaitFor...)
}

// OnAllStopped registers a callback that gets executed when all given modules have been stopped.
func OnAllStopped(callback func(), modulesToWaitFor ...Interface) (unsubscribe func()) {
	return onModulesTriggered(Interface.HookStopped, callback, modulesToWaitFor...)
}

// onModulesTriggered registers a callback that gets executed when all given modules have triggered the given hook.
func onModulesTriggered(targetHook func(Interface, func()) func(), callback func(), modulesToWaitFor ...Interface) (unsubscribe func()) {
	var (
		expectedTriggerCount = int64(len(modulesToWaitFor))
		actualTriggerCount   atomic.Int64
	)

	unsubscribeFunctions := make([]func(), len(modulesToWaitFor))
	for i, module := range modulesToWaitFor {
		unsubscribeFunctions[i] = targetHook(module, func() {
			if actualTriggerCount.Add(1) == expectedTriggerCount {
				callback()
			}
		})
	}

	return lo.Batch(unsubscribeFunctions...)
}
