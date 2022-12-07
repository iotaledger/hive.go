package timeutil

import (
	"context"
	"time"
)

// Sleep pauses the current goroutine for the duration d or until the context ctx is canceled.
// It returns whether Sleep paused for the entire duration or was canceled before that.
func Sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer CleanupTimer(t)

	select {
	case <-ctx.Done():
		return false

	case <-t.C:
		return true
	}
}
