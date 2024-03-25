package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// moduleImpl is the default implementation of a Module.
type moduleImpl struct {
	// constructed is an Event that is triggered when the module was constructed.
	constructed reactive.Event

	// initialized is an Event that is triggered when the module was initialized.
	initialized reactive.Event

	// shutdown is an Event that is triggered when the module begins its shutdown process.
	shutdown reactive.Event

	// stopped is an Event that is triggered when the module finishes its shutdown process.
	stopped reactive.Event

	// Logger exposes the logging capabilities of the module.
	log.Logger
}

// newModuleImpl creates a new reactive moduleImpl with the given logger.
func newModuleImpl(logger log.Logger) *moduleImpl {
	return &moduleImpl{
		constructed: reactive.NewEvent(),
		initialized: reactive.NewEvent(),
		shutdown:    reactive.NewEvent(),
		stopped:     reactive.NewEvent(),
		Logger:      logger,
	}
}

// ConstructedEvent is the getter for an Event that is triggered when the module was constructed.
func (m *moduleImpl) ConstructedEvent() reactive.Event {
	return m.constructed
}

// InitializedEvent is the getter for an Event that is triggered when the module was initialized.
func (m *moduleImpl) InitializedEvent() reactive.Event {
	return m.initialized
}

// ShutdownEvent is the getter for an Event that is triggered when the module begins its shutdown process.
func (m *moduleImpl) ShutdownEvent() reactive.Event {
	return m.shutdown
}

// StoppedEvent is the getter for an Event that is triggered when the module finishes its shutdown process.
func (m *moduleImpl) StoppedEvent() reactive.Event {
	return m.stopped
}

// NewSubModule creates a new reactive submodule with the given name.
func (m *moduleImpl) NewSubModule(name string) Module {
	childLogger := m.NewChildLogger(name)
	m.shutdown.OnTrigger(childLogger.Shutdown)

	return New(childLogger)
}
