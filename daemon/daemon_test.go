package daemon_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	ordered_daemon "github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graceTime for go routines to start
const graceTime = 2 * time.Millisecond

func TestRun(t *testing.T) {
	daemon := ordered_daemon.New()

	var workerStarted typeutils.AtomicBool
	err := daemon.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		workerStarted.Set()
		<-shutdownSignal
	})
	require.NoError(t, err)

	assert.False(t, workerStarted.IsSet())

	var runFinished typeutils.AtomicBool
	go func() {
		daemon.Run()
		runFinished.Set()
	}()
	time.Sleep(graceTime)

	assert.False(t, runFinished.IsSet())
	daemon.ShutdownAndWait()
	time.Sleep(graceTime)
	assert.True(t, runFinished.IsSet())
}

func TestStartShutdown(t *testing.T) {
	daemon := ordered_daemon.New()

	var isShutdown, wasStarted typeutils.AtomicBool
	err := daemon.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		wasStarted.Set()
		<-shutdownSignal
		isShutdown.Set()
	})
	require.NoError(t, err)
	time.Sleep(graceTime)

	assert.False(t, wasStarted.IsSet())
	assert.False(t, isShutdown.IsSet())

	daemon.Start()
	time.Sleep(graceTime)
	assert.True(t, wasStarted.IsSet())
	assert.False(t, isShutdown.IsSet())

	daemon.ShutdownAndWait()
	assert.True(t, wasStarted.IsSet())
	assert.True(t, isShutdown.IsSet())
}

func TestShutdownOrder(t *testing.T) {
	daemon := ordered_daemon.New()

	orders := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	feedback := make(chan int, len(orders))
	for _, order := range orders {
		o := order
		err := daemon.BackgroundWorker(strconv.Itoa(o), func(shutdownSignal <-chan struct{}) {
			<-shutdownSignal
			feedback <- o
		}, o)
		require.NoError(t, err)
	}

	daemon.Start()
	daemon.ShutdownAndWait()
	close(feedback)

	for i := len(orders) - 1; i >= 0; i-- {
		assert.Equal(t, i, <-feedback)
	}
}

func TestReRun(t *testing.T) {
	daemon := ordered_daemon.New()

	terminate := make(chan struct{}, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	require.NoError(t, daemon.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		defer wg.Done()
		<-terminate
	}))

	daemon.Start()

	terminate <- struct{}{}
	wg.Wait()

	var wasStarted typeutils.AtomicBool
	require.NoError(t, daemon.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		wasStarted.Set()
		<-shutdownSignal
	}))

	daemon.ShutdownAndWait()
	assert.True(t, wasStarted.IsSet())
}
