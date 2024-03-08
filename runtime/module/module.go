package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// Module is a trait that exposes logging and lifecycle related API capabilities, that can be used to create a modular
// architecture where different modules can listen and wait for each other to reach certain states.
type Module interface {
	// ConstructedEvent is the getter for an Event that is triggered when the moduleImpl was constructed.
	ConstructedEvent() reactive.Event

	// InitializedEvent is the getter for an Event that is triggered when the moduleImpl was initialized.
	InitializedEvent() reactive.Event

	// ShutdownEvent is the getter for an Event that is triggered when the moduleImpl begins its shutdown process.
	ShutdownEvent() reactive.Event

	// StoppedEvent is the getter for an Event that is triggered when the moduleImpl finishes its shutdown process.
	StoppedEvent() reactive.Event

	// NewSubModule creates a new reactive submodule with the given name.
	NewSubModule(name string) Module

	// Logger is the logger of the moduleImpl.
	log.Logger
}

// New creates a new ReactiveModule with the given logger.
func New(logger log.Logger) Module {
	return newModuleImpl(logger)
}
