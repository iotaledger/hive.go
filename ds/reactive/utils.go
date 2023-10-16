package reactive

import (
	"sync"
)

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
