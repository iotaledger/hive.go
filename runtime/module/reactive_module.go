package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// ReactiveModule defines the interface of a reactive module.
type ReactiveModule interface {
	// ConstructedEvent is the getter for an Event that is triggered when the module was constructed.
	ConstructedEvent() reactive.Event

	// InitializedEvent is the getter for an Event that is triggered when the module was initialized.
	InitializedEvent() reactive.Event

	// ShutdownEvent is the getter for an Event that is triggered when the module begins its shutdown process.
	ShutdownEvent() reactive.Event

	// StoppedEvent is the getter for an Event that is triggered when the module finishes its shutdown process.
	StoppedEvent() reactive.Event

	// NewSubModule creates a new reactive submodule with the given name.
	NewSubModule(name string) ReactiveModule

	// Logger is the logger of the module.
	log.Logger
}

// NewReactiveModule creates a new ReactiveModule with the given logger.
func NewReactiveModule(logger log.Logger) ReactiveModule {
	return newReactiveModule(logger)
}
