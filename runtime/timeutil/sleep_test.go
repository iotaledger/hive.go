package timeutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/runtime/timeutil"
)

func TestSleepCompleted(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	completed := timeutil.Sleep(ctx, 500*time.Millisecond)
	elapsed := time.Since(start)

	require.Truef(t, completed, "Sleep function returned false, but it should have completed")
	require.Greaterf(t, elapsed, 500*time.Millisecond, "Sleep function returned before the specified duration")
}

func TestSleepCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	completed := timeutil.Sleep(ctx, time.Second)
	elapsed := time.Since(start)

	require.Falsef(t, completed, "Sleep function returned true, but it should have been canceled")
	require.Lessf(t, elapsed, 10*time.Millisecond, "Sleep function took longer than expected")
}
