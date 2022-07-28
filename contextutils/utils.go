package contextutils

import (
	"context"
)

// ReturnErrIfCtxDone returns the given error if the provided context is done.
func ReturnErrIfCtxDone(ctx context.Context, err error) error {
	select {
	case <-ctx.Done():
		return err
	default:
		return nil
	}
}

// ReturnErrIfChannelClosed returns the given error if the provided channel was closed.
func ReturnErrIfChannelClosed(channel <-chan struct{}, err error) error {
	select {
	case <-channel:
		return err
	default:
		return nil
	}
}
