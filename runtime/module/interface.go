package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// Interface defines the interface of a module.
type Interface interface {
	// TriggerConstructed triggers the constructed event.
	TriggerConstructed()

	// WasConstructed returns true if the constructed event was triggered.
	WasConstructed() bool

	// HookConstructed registers a callback for the constructed event.
	HookConstructed(func()) (unsubscribe func())

	// TriggerInitialized triggers the initialized event.
	TriggerInitialized()

	// WasInitialized returns true if the initialized event was triggered.
	WasInitialized() bool

	// HookInitialized registers a callback for the initialized event.
	HookInitialized(func()) (unsubscribe func())

	// Shutdown shuts down the module, should finally call TriggerStopped.
	Shutdown()

	// TriggerShutdown triggers the shutdown event.
	TriggerShutdown()

	// WasShutdown returns true if the shutdown event was triggered.
	WasShutdown() bool

	// HookShutdown registers a callback for the shutdown event.
	HookShutdown(func()) (unsubscribe func())

	// TriggerStopped triggers the stopped event.
	TriggerStopped()

	// WasStopped returns true if the stopped event was triggered.
	WasStopped() bool

	// HookStopped registers a callback for the stopped event.
	HookStopped(func()) (unsubscribe func())
}

type ReactiveInterface interface {
	// ConstructedEvent returns a reactive.Event that is triggered when the module was constructed.
	ConstructedEvent() reactive.Event

	// InitializedEvent returns a reactive.Event that is triggered when the module was initialized.
	InitializedEvent() reactive.Event

	// ShutdownEvent returns a reactive.Event that is triggered when the module begins its shutdown process.
	ShutdownEvent() reactive.Event

	// StoppedEvent returns a reactive.Event that is triggered when the module finishes its shutdown process.
	StoppedEvent() reactive.Event

	// Logger is the logger of the module.
	log.Logger
}
