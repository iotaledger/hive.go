package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCallback_LockExecution tests the LockExecution method of the callback type
func TestCallback_LockExecution(t *testing.T) {
	// Create a new callback
	cb := newCallback(func() {})
	require.NotNil(t, cb)

	// Test locking the execution when not unsubscribed and updateID is different
	locked := cb.LockExecution(1)
	require.True(t, locked)
	cb.UnlockExecution()

	// Try to lock again with the same updateID, should fail
	locked = cb.LockExecution(1)
	require.False(t, locked)

	// Set the callback as unsubscribed and try to lock
	cb.unsubscribed = true
	locked = cb.LockExecution(2)
	require.False(t, locked)

	// Reset unsubscribed for next test
	cb.unsubscribed = false

	// Test locking with a different updateID
	locked = cb.LockExecution(3)
	require.True(t, locked)
	cb.UnlockExecution()
}
