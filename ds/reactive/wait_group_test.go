package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWaitGroup(t *testing.T) {
	wg := NewWaitGroup(1, 2, 3)
	wg.Debug()

	require.True(t, wg.PendingElements().HasAll(NewSet(1, 2, 3)))
	require.False(t, wg.WasTriggered())

	wg.Done(1, 2)

	require.True(t, wg.PendingElements().HasAll(NewSet(3)))
	require.False(t, wg.WasTriggered())

	wg.Done(1, 2)

	require.True(t, wg.PendingElements().HasAll(NewSet(3)))
	require.False(t, wg.WasTriggered())

	wg.Add(4)

	require.True(t, wg.PendingElements().HasAll(NewSet(3, 4)))
	require.False(t, wg.WasTriggered())

	wg.Add(4)

	require.True(t, wg.PendingElements().HasAll(NewSet(3, 4)))
	require.False(t, wg.WasTriggered())

	wg.Done(3, 4)

	require.True(t, wg.PendingElements().IsEmpty())
	require.True(t, wg.WasTriggered())
}
