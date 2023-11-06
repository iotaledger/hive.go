//nolint:staticcheck,golint,revive // we don't care about these linters in test cases
package contextutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergedContextCancel(t *testing.T) {

	mergedCtx, mergedCancel := MergeContexts(context.Background(), context.Background())
	mergedCancel()

	require.True(t, func() bool {
		select {
		case <-mergedCtx.Done():
			return true
		default:
			return false
		}
	}())

	require.ErrorIs(t, mergedCtx.Err(), ErrMergedContextCanceled)
	require.ErrorIs(t, mergedCtx.Err(), context.Canceled)
}

func TestMergedContextPrimaryCancel(t *testing.T) {

	ctx1, cancel1 := context.WithCancel(context.WithValue(context.Background(), "one", 1))
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.WithValue(context.Background(), "two", 2))
	defer cancel2()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)
	cancel1()

	require.True(t, func() bool {
		select {
		case <-mergedCtx.Done():
			return true
		case <-time.After(1 * time.Second):
			return false
		}
	}())

	require.Equal(t, ctx1.Err(), mergedCtx.Err())
}

func TestMergedContextSecondaryCancel(t *testing.T) {

	ctx1, cancel1 := context.WithCancel(context.WithValue(context.Background(), "one", 1))
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.WithValue(context.Background(), "two", 2))
	defer cancel2()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)
	cancel2()

	require.True(t, func() bool {
		select {
		case <-mergedCtx.Done():
			return true
		case <-time.After(1 * time.Second):
			return false
		}
	}())

	require.Equal(t, ctx2.Err(), mergedCtx.Err())
}

func TestMergedContextPrimaryCancelBefore(t *testing.T) {

	ctx1, cancel1 := context.WithCancel(context.WithValue(context.Background(), "one", 1))
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.WithValue(context.Background(), "two", 2))
	defer cancel2()

	cancel1()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)

	require.True(t, func() bool {
		select {
		case <-mergedCtx.Done():
			return true
		default:
			return false
		}
	}())

	require.Equal(t, ctx1.Err(), mergedCtx.Err())
}

func TestMergedContextSecondaryCancelBefore(t *testing.T) {

	ctx1, cancel1 := context.WithCancel(context.WithValue(context.Background(), "one", 1))
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.WithValue(context.Background(), "two", 2))
	defer cancel2()

	cancel2()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)

	require.True(t, func() bool {
		select {
		case <-mergedCtx.Done():
			return true
		default:
			return false
		}
	}())

	require.Equal(t, ctx2.Err(), mergedCtx.Err())
}

func TestMergedContextValues(t *testing.T) {

	ctx1, cancel1 := context.WithCancel(context.WithValue(context.Background(), "one", 1))
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.WithValue(context.Background(), "two", 2))
	defer cancel2()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)

	require.Equal(t, mergedCtx.Value("one"), 1)
	require.Equal(t, mergedCtx.Value("two"), 2)
	require.Nil(t, mergedCtx.Value("three"))
}

func TestMergedContextDeadline(t *testing.T) {

	deadline1 := time.Now().Add(10 * time.Second)
	deadline2 := time.Now().Add(1 * time.Second)

	ctx1, cancel1 := context.WithDeadline(context.Background(), deadline1)
	defer cancel1()

	ctx2, cancel2 := context.WithDeadline(context.Background(), deadline2)
	defer cancel2()

	mergedCtx, _ := MergeContexts(ctx1, ctx2)

	deadline, ok := mergedCtx.Deadline()
	require.False(t, deadline.IsZero())
	require.True(t, ok)

	require.Equal(t, deadline2, deadline)
}
