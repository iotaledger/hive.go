package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// reactiveImpl is the default implementation of a reactive module.
type reactiveImpl struct {
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

// newReactiveImpl creates a new reactive module with the given logger.
func newReactiveImpl(logger log.Logger) *reactiveImpl {
	return &reactiveImpl{
		constructed: reactive.NewEvent(),
		initialized: reactive.NewEvent(),
		shutdown:    reactive.NewEvent(),
		stopped:     reactive.NewEvent(),
		Logger:      logger,
	}
}

// Constructed is the getter for an Event that is triggered when the module was constructed.
func (r *reactiveImpl) Constructed() reactive.Event {
	return r.constructed
}

// Initialized is the getter for an Event that is triggered when the module was initialized.
func (r *reactiveImpl) Initialized() reactive.Event {
	return r.stopped
}

// Shutdown is the getter for an Event that is triggered when the module begins its shutdown process.
func (r *reactiveImpl) Shutdown() reactive.Event {
	return r.shutdown
}

// Stopped is the getter for an Event that is triggered when the module finishes its shutdown process.
func (r *reactiveImpl) Stopped() reactive.Event {
	return r.stopped
}

// NewSubModule creates a new reactive submodule with the given name.
func (r *reactiveImpl) NewSubModule(name string) Reactive {
	childLogger := r.NewChildLogger(name)
	r.shutdown.OnTrigger(childLogger.UnsubscribeFromParentLogger)

	return NewReactive(childLogger)
}
