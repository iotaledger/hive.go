package timeutil

import (
	"context"
	"time"
)

func Sleep(ctx context.Context, interval time.Duration) bool {
	select {
	case <-ctx.Done():
		return false

	case <-time.After(interval):
		return true
	}
}
