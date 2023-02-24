package valuenotifier

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	l := New[string]()

	aListener := l.Listener("a")
	aListener2 := l.Listener("a")
	bListener := l.Listener("b")
	dListener := l.Listener("d")
	eListener := l.Listener("e")
	fListener := l.Listener("f")
	gListener := l.Listener("g")

	// We expect that if we deregister this listener, that the other one will still work
	aListener2.Deregister()

	go func() {
		time.Sleep(200 * time.Millisecond)
		l.Notify("a")
		l.Notify("b")
		l.Notify("c")
		l.Notify("d")
		eListener.Deregister()
		l.Notify("e")
		l.Notify("unknown")
		// never notify "f"
	}()

	// We expect d times out before it is called
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, dListener.Wait(ctx), context.DeadlineExceeded)

	// We expect that an already deregistered listener cannot wait
	require.ErrorIs(t, aListener2.Wait(context.Background()), ErrListenerDeregistered)

	require.NoError(t, aListener.Wait(context.Background()))
	require.NoError(t, bListener.Wait(context.Background()))

	// We expect that a waiting listener fails if it is deregistered
	require.ErrorIs(t, eListener.Wait(context.Background()), ErrListenerDeregistered)

	// We expect f to time out because it was never called
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, fListener.Wait(ctx), context.DeadlineExceeded)

	// We expect g to return context.Canceled because the context used to wait is canceled
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	cancel()
	require.ErrorIs(t, gListener.Wait(ctx), context.Canceled)

	// There should be no listeners registered
	require.True(t, l.listeners.IsEmpty())
}
