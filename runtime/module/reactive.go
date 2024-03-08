package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// Reactive defines the interface of a reactive module.
type Reactive interface {
	// Constructed is the getter for an Event that is triggered when the module was constructed.
	Constructed() reactive.Event

	// Initialized is the getter for an Event that is triggered when the module was initialized.
	Initialized() reactive.Event

	// Shutdown is the getter for an Event that is triggered when the module begins its shutdown process.
	Shutdown() reactive.Event

	// Stopped is the getter for an Event that is triggered when the module finishes its shutdown process.
	Stopped() reactive.Event

	// NewSubModule creates a new reactive submodule with the given name.
	NewSubModule(name string) Reactive

	// Logger is the logger of the module.
	log.Logger
}

// NewReactive creates a new reactive module with the given logger.
func NewReactive(logger log.Logger) Reactive {
	return newReactiveImpl(logger)
}
