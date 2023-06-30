package walker_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds/walker"
)

func TestWalker(t *testing.T) {
	w := walker.New[int]()

	// Test Push and Next
	w.Push(1)
	w.Push(2)
	w.Push(3)

	require.True(t, w.HasNext())
	require.Equal(t, 1, w.Next())
	require.Equal(t, 2, w.Next())
	require.Equal(t, 3, w.Next())
	require.False(t, w.HasNext())

	// Test PushAll
	w.PushAll(4, 5, 6)

	require.True(t, w.HasNext())
	require.Equal(t, 4, w.Next())
	require.Equal(t, 5, w.Next())
	require.Equal(t, 6, w.Next())
	require.False(t, w.HasNext())

	// Test PushFront
	w.PushFront(7, 8)

	require.True(t, w.HasNext())
	require.Equal(t, 8, w.Next())
	require.Equal(t, 7, w.Next())
	require.False(t, w.HasNext())

	// Test Pushed
	require.True(t, w.Pushed(4))
	require.True(t, w.Pushed(5))
	require.True(t, w.Pushed(6))
	require.True(t, w.Pushed(7))
	require.True(t, w.Pushed(8))
	require.False(t, w.Pushed(9))
	require.False(t, w.Pushed(10))

	// Test StopWalk and WalkStopped
	w.StopWalk()

	require.True(t, w.WalkStopped())
	require.False(t, w.HasNext())

	// Test Reset
	w.Reset()

	require.False(t, w.WalkStopped())
	require.False(t, w.HasNext())
}
