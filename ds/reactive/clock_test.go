package reactive

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClock(t *testing.T) {
	var tickCount atomic.Int32

	clock := NewClock(100 * time.Millisecond)
	clock.OnTick(func(now time.Time) {
		tickCount.Add(1)
		require.Less(t, time.Now().Sub(now), 100*time.Millisecond)
	})

	time.Sleep(time.Second)

	require.GreaterOrEqual(t, int(tickCount.Load()), 10)
}

// TestClockInitialization tests the initialization of the clock struct
func TestClockInitialization(t *testing.T) {
	clk := newClock(1 * time.Second)

	if clk.variable == nil {
		t.Errorf("Expected variable to be initialized, but it was nil")
	}

	if clk.shutdown == nil {
		t.Errorf("Expected shutdown channel to be initialized, but it was nil")
	}
}

// TestClockDefaultTime tests the default time set in the clock struct
func TestClockDefaultTime(t *testing.T) {
	clk := newClock(1 * time.Second)
	initialTime := time.Now()

	require.Equal(t, initialTime.Truncate(1*time.Second), clk.variable.Get().Truncate(1*time.Second), "Expected initial time to be current time")
}

// TestClockShutdown tests the shutdown functionality of the clock
func TestClockShutdown(t *testing.T) {
	clk := newClock(100 * time.Millisecond)
	initialTime := clk.variable.Get()

	// Send a shutdown signal
	clk.Shutdown()

	// Wait for longer than the update interval to ensure no further updates
	time.Sleep(1 * time.Second)

	require.Equal(t, initialTime.Truncate(1*time.Second), clk.variable.Get().Truncate(1*time.Second), "Expected clock time to not update after shutdown")
}
