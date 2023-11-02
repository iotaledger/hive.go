package reactive

import (
	"sync"
)

// WithValue is a utility function that allows to set up dynamic behavior based on the current value of a
// ReadableVariable which is automatically torn down once the value of the ReadableVariable changes again.
func WithValue[T comparable](variable ReadableVariable[T], setup func(value T) (teardown func())) (teardown func()) {
	return variable.OnUpdateWithContext(func(_, value T, unsubscribeOnUpdate func(setup func() (teardown func()))) {
		unsubscribeOnUpdate(func() func() { return setup(value) })
	})
}

// WithNonEmptyValue is a utility function that allows to set up dynamic behavior based on the current (non-empty) value
// of a ReadableVariable which is automatically torn down once the value of the ReadableVariable changes again.
func WithNonEmptyValue[T comparable](setup func(value T) (teardown func()), variable ReadableVariable[T]) (teardown func()) {
	return variable.OnUpdateWithContext(func(_, value T, unsubscribeOnUpdate func(setup func() (teardown func()))) {
		if value != *new(T) {
			unsubscribeOnUpdate(func() func() { return setup(value) })
		}
	})
}

// AssignDynamicValue is a utility function that allows to assign a dynamic value to a given Variable. It returns a tear
// down function that can be used to unsubscribe the Variable from the dynamic value.
func AssignDynamicValue[T comparable](variable Variable[T], dynamicValue DerivedVariable[T]) (teardown func()) {
	// no need to unsubscribe variable from dynamic value (it will no longer change and get garbage collected after
	// unsubscribing itself)
	_ = variable.InheritFrom(dynamicValue)

	return dynamicValue.Unsubscribe
}

// region callback /////////////////////////////////////////////////////////////////////////////////////////////////////

// callback is an internal wrapper for a callback function that is extended by an ID and a mutex (for call order
// synchronization).
type callback[FuncType any] struct {
	// Invoke is the callback function that is invoked when the callback is triggered.
	Invoke FuncType

	// unsubscribed is a flag that indicates whether the callback was unsubscribed.
	unsubscribed bool

	// lastUpdate is the last update that was applied to the callback.
	lastUpdate uniqueID

	// executionMutex is the mutex that is used to synchronize the execution of the callback.
	executionMutex sync.Mutex
}

// newCallback is the constructor for the callback type.
func newCallback[FuncType any](invoke FuncType) *callback[FuncType] {
	return &callback[FuncType]{
		Invoke: invoke,
	}
}

// LockExecution locks the callback for the given update and returns true if the callback was locked successfully.
func (c *callback[FuncType]) LockExecution(updateID uniqueID) bool {
	c.executionMutex.Lock()

	if c.unsubscribed || updateID != 0 && updateID == c.lastUpdate {
		c.executionMutex.Unlock()

		return false
	}

	c.lastUpdate = updateID

	return true
}

// UnlockExecution unlocks the callback.
func (c *callback[FuncType]) UnlockExecution() {
	c.executionMutex.Unlock()
}

// MarkUnsubscribed marks the callback as unsubscribed (it will no longer trigger).
func (c *callback[FuncType]) MarkUnsubscribed() {
	c.executionMutex.Lock()
	defer c.executionMutex.Unlock()

	c.unsubscribed = true
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region uniqueID /////////////////////////////////////////////////////////////////////////////////////////////////////

// UniqueID is a unique identifier.
type uniqueID uint64

// Next returns the next unique identifier.
func (u *uniqueID) Next() uniqueID {
	*u++

	return *u
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
