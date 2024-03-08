package module

import (
	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// module is the default implementation of a Module.
type module struct {
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

// newModule creates a new reactive module with the given logger.
func newModule(logger log.Logger) *module {
	return &module{
		constructed: reactive.NewEvent(),
		initialized: reactive.NewEvent(),
		shutdown:    reactive.NewEvent(),
		stopped:     reactive.NewEvent(),
		Logger:      logger,
	}
}

// ConstructedEvent is the getter for an Event that is triggered when the module was constructed.
func (m *module) ConstructedEvent() reactive.Event {
	return m.constructed
}

// InitializedEvent is the getter for an Event that is triggered when the module was initialized.
func (m *module) InitializedEvent() reactive.Event {
	return m.initialized
}

// ShutdownEvent is the getter for an Event that is triggered when the module begins its shutdown process.
func (m *module) ShutdownEvent() reactive.Event {
	return m.shutdown
}

// StoppedEvent is the getter for an Event that is triggered when the module finishes its shutdown process.
func (m *module) StoppedEvent() reactive.Event {
	return m.stopped
}

// NewSubModule creates a new reactive submodule with the given name.
func (m *module) NewSubModule(name string) Module {
	childLogger := m.NewChildLogger(name)
	m.shutdown.OnTrigger(childLogger.UnsubscribeFromParentLogger)

	return New(childLogger)
}
