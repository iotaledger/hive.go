package backoff

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ierrors"
)

func TestMaxRetries(t *testing.T) {
	const retries = 10

	p := ZeroBackOff().With(MaxRetries(retries))

	var count uint
	err := Retry(p, func() error {
		count++

		return errTest
	})
	assert.True(t, ierrors.Is(err, errTest))
	assert.EqualValues(t, retries+1, count)
}

func TestMaxRetriesNew(t *testing.T) {
	const retries = 10

	p := ZeroBackOff().With(MaxRetries(retries))

	for i := 0; i < 3; i++ {
		var count uint
		err := Retry(p, func() error {
			count++

			return errTest
		})
		assert.True(t, ierrors.Is(err, errTest))
		assert.EqualValues(t, retries+1, count)
	}
}

func TestMaxRetriesParallel(t *testing.T) {
	const (
		retries     = 10
		parallelism = 3
	)

	var wg sync.WaitGroup
	p := ZeroBackOff().With(MaxRetries(retries))

	test := func() {
		defer wg.Done()
		var count uint
		err := Retry(p, func() error {
			count++

			return errTest
		})
		assert.True(t, ierrors.Is(err, errTest))
		assert.EqualValues(t, retries+1, count)
	}

	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go test()
	}
	wg.Wait()
}

func TestMaxInterval(t *testing.T) {
	const interval = 20 * time.Millisecond

	p := ConstantBackOff(10 * interval).With(MaxInterval(interval))

	var (
		count uint
		last  time.Time
	)
	err := Retry(p, func() error {
		now := time.Now()
		if count > 0 {
			assertInterval(t, interval, now.Sub(last))
		}
		last = now
		if count >= 10 {
			return nil
		}
		count++

		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 10, count)
}

func TestTimeout(t *testing.T) {
	timeout := time.Now().Add(20 * time.Millisecond)
	max := timeout.Add(intervalDelta)

	p := ZeroBackOff().With(Timeout(timeout))

	err := Retry(p, func() error {
		assert.LessOrEqual(t, time.Now().UnixNano(), max.UnixNano())

		return errTest
	})
	assert.True(t, ierrors.Is(err, errTest))
	assert.GreaterOrEqual(t, time.Now().UnixNano(), timeout.UnixNano())
}

func TestCancel(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	p := ZeroBackOff().With(Cancel(ctx))

	stopped := make(chan struct{}, 1)
	go func() {
		err := Retry(p, func() error {
			return errTest
		})
		assert.True(t, ierrors.Is(err, errTest))
		stopped <- struct{}{}
	}()
	time.Sleep(10 * time.Millisecond)
	select {
	case <-stopped:
		assert.FailNow(t, "retry stopped prematurely")
	default:
	}
	ctxCancel()
	<-stopped
}

func TestJitter(t *testing.T) {
	const (
		interval    = 20 * time.Millisecond
		factor      = 0.5
		minInterval = time.Duration(float64(interval)*factor) - intervalDelta
		maxInterval = interval + intervalDelta
	)

	p := ConstantBackOff(interval).With(Jitter(factor))

	var (
		count uint
		last  time.Time
	)
	err := Retry(p, func() error {
		now := time.Now()
		if count > 0 {
			assert.GreaterOrEqual(t, now.Sub(last).Microseconds(), minInterval.Microseconds())
			assert.Less(t, now.Sub(last).Microseconds(), maxInterval.Microseconds())
		}
		last = now
		if count >= 10 {
			return nil
		}
		count++

		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 10, count)
}
