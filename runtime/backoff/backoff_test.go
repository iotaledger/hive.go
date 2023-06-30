//nolint:unparam // we don't care about these linters in test cases
package backoff

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ierrors"
)

var errTest = ierrors.New("test")

const (
	intervalDelta = 100 * time.Millisecond // allowed deviation to pass the test
	retryCount    = 5
)

func assertInterval(t assert.TestingT, expected time.Duration, actual time.Duration) bool {
	return assert.GreaterOrEqual(t, actual.Microseconds(), expected.Microseconds()) &&
		assert.LessOrEqual(t, actual.Microseconds(), (expected+intervalDelta).Microseconds())
}

func TestNoError(t *testing.T) {
	err := Retry(ZeroBackOff(), func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestPermanentError(t *testing.T) {
	err := Retry(ZeroBackOff(), func() error {
		return Permanent(errTest)
	})
	assert.EqualError(t, err, errTest.Error())
}

func TestZeroBackOff(t *testing.T) {
	var count uint
	last := time.Now()
	err := Retry(ZeroBackOff(), func() error {
		now := time.Now()
		assertInterval(t, 0, now.Sub(last))
		last = now
		if count >= retryCount {
			return nil
		}
		count++

		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, retryCount, count)
}

func TestConstantBackOff(t *testing.T) {
	const interval = 2 * intervalDelta

	p := ConstantBackOff(interval)

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
		if count >= retryCount {
			return nil
		}
		count++

		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, retryCount, count)
}

func TestExponentialBackOff(t *testing.T) {
	const (
		interval = 2 * intervalDelta
		factor   = 1.5
	)

	p := ExponentialBackOff(interval, factor)

	var (
		count uint
		last  time.Time
	)
	err := Retry(p, func() error {
		now := time.Now()
		if count > 0 {
			expected := time.Duration(float64(interval) * math.Pow(factor, float64(count-1)))
			assertInterval(t, expected, now.Sub(last))
		}
		last = now
		if count >= retryCount {
			return nil
		}
		count++

		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, retryCount, count)
}

func TestExponentialBackOffParallel(t *testing.T) {
	const (
		interval    = 20 * time.Millisecond
		factor      = 1.5
		parallelism = 3
	)

	var wg sync.WaitGroup
	p := ExponentialBackOff(interval, factor)

	test := func() {
		defer wg.Done()
		var (
			count uint
			last  time.Time
		)
		err := Retry(p, func() error {
			now := time.Now()
			if count > 0 {
				expected := time.Duration(float64(interval) * math.Pow(factor, float64(count-1)))
				assertInterval(t, expected, now.Sub(last))
			}
			last = now
			if count >= retryCount {
				return nil
			}
			count++

			return errTest
		})
		assert.NoError(t, err)
		assert.EqualValues(t, retryCount, count)
	}

	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go test()
	}
	wg.Wait()
}

func BenchmarkBackOff(b *testing.B) {
	p := ZeroBackOff().With(
		MaxRetries(b.N),
		MaxInterval(time.Microsecond),
		Timeout(time.Now().Add(time.Minute)),
		Cancel(context.Background()),
		Jitter(0.5),
	)

	b.ResetTimer()
	_ = Retry(p, func() error {
		return errTest
	})
}
