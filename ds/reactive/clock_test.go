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
		require.Less(t, time.Now().Sub(now), 30*time.Millisecond)
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

// TestClockInitialTime tests the default time set in the clock struct
func TestClockInitialTime(t *testing.T) {
	clk := newClock(1 * time.Millisecond)
	initialTime := time.Now()

	diff := initialTime.Sub(clk.variable.Get())
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff <= 10*time.Millisecond, "Expected initial time to be within 10ms of current time")
}

// TestClockShutdown tests the shutdown functionality of the clock
func TestClockShutdown(t *testing.T) {
	clk := newClock(100 * time.Millisecond)
	initialTime := clk.variable.Get()

	// Send a shutdown signal
	clk.Shutdown()

	// Wait for a short time to ensure the clock is not advancing after shutdown
	time.Sleep(1 * time.Second)

	// Wait for longer than the update interval to ensure no further updates
	diff := initialTime.Sub(clk.variable.Get())
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff < 100*time.Millisecond, "Expected clock time to not update after shutdown")
}
