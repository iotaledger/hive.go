package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// reactiveModule is the default implementation of a ReactiveModule.
type reactiveModule struct {
	// constructed is an Event that is triggered when the module was constructed.
	constructed reactive.Event

	// initialized is an Event that is triggered when the module was initialized.
	initialized reactive.Event

	// shutdown is an Event that is triggered when the module begins its shutdown process.
	shutdown reactive.Event

	// stopped is an Event that is triggered when the module finishes its shutdown process.
	stopped reactive.Event

	// Logger is the logger of the module.
	log.Logger
}

// newReactiveModule creates a new reactive module with the given logger.
func newReactiveModule(logger log.Logger) *reactiveModule {
	return &reactiveModule{
		constructed: reactive.NewEvent(),
		initialized: reactive.NewEvent(),
		shutdown:    reactive.NewEvent(),
		stopped:     reactive.NewEvent(),
		Logger:      logger,
	}
}

// ConstructedEvent is the getter for an Event that is triggered when the module was constructed.
func (r *reactiveModule) ConstructedEvent() reactive.Event {
	return r.constructed
}

// InitializedEvent is the getter for an Event that is triggered when the module was initialized.
func (r *reactiveModule) InitializedEvent() reactive.Event {
	return r.stopped
}

// ShutdownEvent is the getter for an Event that is triggered when the module begins its shutdown process.
func (r *reactiveModule) ShutdownEvent() reactive.Event {
	return r.shutdown
}

// StoppedEvent is the getter for an Event that is triggered when the module finishes its shutdown process.
func (r *reactiveModule) StoppedEvent() reactive.Event {
	return r.stopped
}

// NewSubModule creates a new reactive submodule with the given name.
func (r *reactiveModule) NewSubModule(name string) ReactiveModule {
	childLogger := r.NewChildLogger(name)
	r.shutdown.OnTrigger(childLogger.UnsubscribeFromParentLogger)

	return NewReactiveModule(childLogger)
}
