package module

import (
	"sync"

	"github.com/iotaledger/hive.go/runtime/promise"
)

// Module is a trait that exposes a lifecycle related API, that can be used to create a modular architecture where
// different modules can listen and wait for each other to reach certain states.
type Module struct {
	// constructed is triggered when the module was constructed.
	constructed *promise.Event

	// initialized is triggered when the module was initialized.
	initialized *promise.Event

	// shutdown is triggered when the module begins its shutdown process.
	shutdown *promise.Event

	// stopped is triggered when the module finishes its shutdown process.
	stopped *promise.Event

	// initOnce is used to ensure that the init function is only called once.
	initOnce sync.Once
}

// TriggerConstructed triggers the constructed event.
func (m *Module) TriggerConstructed() {
	m.initOnce.Do(m.init)

	m.constructed.Trigger()
}

// WasConstructed returns true if the constructed event was triggered.
func (m *Module) WasConstructed() bool {
	m.initOnce.Do(m.init)

	return m.constructed.WasTriggered()
}

// HookConstructed registers a callback for the constructed event.
func (m *Module) HookConstructed(callback func()) (unsubscribe func()) {
	m.initOnce.Do(m.init)

	return m.constructed.OnTrigger(callback)
}

// TriggerInitialized triggers the initialized event.
func (m *Module) TriggerInitialized() {
	m.initOnce.Do(m.init)

	m.initialized.Trigger()
}

// WasInitialized returns true if the initialized event was triggered.
func (m *Module) WasInitialized() bool {
	m.initOnce.Do(m.init)

	return m.initialized.WasTriggered()
}

// HookInitialized registers a callback for the initialized event.
func (m *Module) HookInitialized(callback func()) (unsubscribe func()) {
	m.initOnce.Do(m.init)

	return m.initialized.OnTrigger(callback)
}

// TriggerShutdown triggers the shutdown event.
func (m *Module) TriggerShutdown() {
	m.initOnce.Do(m.init)

	m.shutdown.Trigger()
}

// WasShutdown returns true if the shutdown event was triggered.
func (m *Module) WasShutdown() bool {
	m.initOnce.Do(m.init)

	return m.shutdown.WasTriggered()
}

// HookShutdown registers a callback for the shutdown event.
func (m *Module) HookShutdown(callback func()) (unsubscribe func()) {
	m.initOnce.Do(m.init)

	return m.shutdown.OnTrigger(callback)
}

// TriggerStopped triggers the stopped event.
func (m *Module) TriggerStopped() {
	m.initOnce.Do(m.init)

	m.stopped.Trigger()
}

// WasStopped returns true if the stopped event was triggered.
func (m *Module) WasStopped() bool {
	m.initOnce.Do(m.init)

	return m.stopped.WasTriggered()
}

// HookStopped registers a callback for the stopped event.
func (m *Module) HookStopped(callback func()) (unsubscribe func()) {
	m.initOnce.Do(m.init)

	return m.stopped.OnTrigger(callback)
}

// init initializes the module.
func (m *Module) init() {
	m.constructed = promise.NewEvent()
	m.initialized = promise.NewEvent()
	m.shutdown = promise.NewEvent()
	m.stopped = promise.NewEvent()
}
