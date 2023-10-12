package reactive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClock(t *testing.T) {
	tickCount := 0

	clock := NewClock(100 * time.Millisecond)
	clock.OnTick(func(now time.Time) {
		tickCount++
		require.Less(t, time.Now().Sub(now), 100*time.Millisecond)
	})

	time.Sleep(time.Second)

	require.Greater(t, tickCount, 10)
}
