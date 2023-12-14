package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// ReactiveModule is a trait that exposes logging functionality and lifecycle related Events, that can be used to create
// a modular architecture where different modules can wait for each other to reach certain states.
type ReactiveModule struct {
	// Constructed is triggered when the module was constructed.
	Constructed reactive.Event

	// Initialized is triggered when the module was initialized.
	Initialized reactive.Event

	// Shutdown is triggered when the module begins its shutdown process.
	Shutdown reactive.Event

	// Stopped is triggered when the module finishes its shutdown process.
	Stopped reactive.Event

	// Logger is the logger of the module.
	log.Logger
}

// NewReactiveModule creates a new reactive module with the given logger.
func NewReactiveModule(logger log.Logger) *ReactiveModule {
	return &ReactiveModule{
		Constructed: reactive.NewEvent(),
		Initialized: reactive.NewEvent(),
		Shutdown:    reactive.NewEvent(),
		Stopped:     reactive.NewEvent(),
		Logger:      logger,
	}
}

// NewReactiveSubModule creates a new reactive submodule with the given name.
func (r *ReactiveModule) NewReactiveSubModule(name string) *ReactiveModule {
	childLogger := r.NewChildLogger(name)
	r.Shutdown.OnTrigger(childLogger.UnsubscribeFromParentLogger)

	return NewReactiveModule(childLogger)
}
