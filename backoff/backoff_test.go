package backoff

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test")

const (
	intervalDelta = 10 * time.Millisecond // allow 10ms deviation to pass the test
)

func assertInterval(t assert.TestingT, expected time.Duration, actual time.Duration) bool {
	return assert.Greater(t, actual.Microseconds(), expected.Microseconds()) &&
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
		assert.LessOrEqual(t, now.Sub(last).Microseconds(), intervalDelta.Microseconds())
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

func TestConstantBackOff(t *testing.T) {
	const interval = 20 * time.Millisecond

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
		if count >= 10 {
			return nil
		}
		count++
		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 10, count)
}

func TestExponentialBackOff(t *testing.T) {
	const (
		interval = 20 * time.Millisecond
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
		if count >= 10 {
			return nil
		}
		count++
		return errTest
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 10, count)
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
			if count >= 10 {
				return nil
			}
			count++
			return errTest
		})
		assert.NoError(t, err)
		assert.EqualValues(t, 10, count)
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
		Cancel(nil),
		Jitter(0.5),
	)

	b.ResetTimer()
	_ = Retry(p, func() error {
		return errTest
	})
}
