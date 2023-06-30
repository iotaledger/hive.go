// Package backoff implements backoff algorithms for retrying operations.
//
// Use Retry function for retrying operations that may fail.
// If Retry does not meet your needs, you can create an own backoff Policy.
//
// See Examples section below for usage examples.
package backoff

import (
	"time"

	"github.com/iotaledger/hive.go/ierrors"
)

// Stop indicates that no more retries should be made for use in NextBackOff().
const Stop time.Duration = -1

// Permanent wraps the given err in a permanent error signaling that the operation should not be retried.
func Permanent(err error) error {
	if err == nil {
		panic("no error specified")
	}

	return &permanentError{
		err: err,
	}
}

// Policy is a backoff policy for retrying an operation.
type Policy interface {
	// NextBackOff returns the duration to wait before retrying the operation,
	// or backoff.Stop to indicate that no more retries should be made.
	NextBackOff() time.Duration

	// New creates a new instance of the policy in its initial state.
	New() Policy
}

// NewBackOff wraps a backoff policy into a modifiable BackOff.
func NewBackOff(policy Policy) *BackOff {
	if b, ok := policy.(*BackOff); ok {
		return b
	}

	return &BackOff{Policy: policy}
}

// BackOff is a backoff policy that can be modified with options.
type BackOff struct {
	Policy
}

// With modifies the backoff policy by applying additional options.
func (b *BackOff) With(opts ...Option) *BackOff {
	p := b.Policy
	for _, opt := range opts {
		p = opt.apply(p)
	}

	return NewBackOff(p)
}

// Retry calls the function f until it does not return error or the backoff policy stops.
// The function is guaranteed to be run at least once.
// If the functions returns a permanent error, the operation is not retried, and the wrapped error is returned.
// Retry sleeps the goroutine for the duration returned by BackOff after a failed operation returns.
func Retry(p Policy, f func() error) error {
	p = p.New()

	var err error
	for {
		err = f()
		if err == nil {
			return nil
		}

		var permanent *permanentError
		if ierrors.As(err, &permanent) {
			return permanent.Unwrap()
		}

		duration := p.NextBackOff()
		if duration == Stop {
			break
		}

		time.Sleep(duration)
	}

	return err
}

// permanentError signals that the operation should not be retried.
type permanentError struct {
	err error
}

func (e *permanentError) Error() string {
	return e.err.Error()
}

func (e *permanentError) Unwrap() error {
	return e.err
}

type zeroPolicy struct{}

func (zeroPolicy) NextBackOff() time.Duration { return 0 }
func (zeroPolicy) New() Policy                { return zeroPolicy{} }

var zeroBackOffInstance = NewBackOff(zeroPolicy{})

// ZeroBackOff returns a fixed backoff policy whose backoff time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
func ZeroBackOff() *BackOff {
	return zeroBackOffInstance
}

type constantPolicy struct {
	Interval time.Duration
}

// ConstantBackOff returns a backoff policy that always returns the same backoff delay. This is in contrast to an
// exponential backoff policy, which returns a delay that grows longer as you call NextBackOff() over and over again.
func ConstantBackOff(d time.Duration) *BackOff {
	return NewBackOff(&constantPolicy{Interval: d})
}

func (b *constantPolicy) NextBackOff() time.Duration { return b.Interval }
func (b *constantPolicy) New() Policy                { return b }

// ExponentialBackOff returns a backoff policy that increases the backoff period for each retry attempt using a
// function that grows exponentially.
// After each call of NextBackOff() the interval is multiplied by the provided factor starting with initialInterval.
func ExponentialBackOff(initialInterval time.Duration, factor float64) *BackOff {
	p := &exponentialPolicy{
		initialInterval: initialInterval,
		factor:          factor,
		currentInterval: initialInterval,
	}

	return NewBackOff(p)
}

type exponentialPolicy struct {
	initialInterval time.Duration
	factor          float64
	currentInterval time.Duration
}

func (b *exponentialPolicy) NextBackOff() time.Duration {
	defer b.incrementCurrentInterval()

	return b.currentInterval
}

func (b *exponentialPolicy) New() Policy {
	return &exponentialPolicy{
		initialInterval: b.initialInterval,
		factor:          b.factor,
		currentInterval: b.currentInterval,
	}
}

// Increments the current interval by multiplying it with the multiplier.
func (b *exponentialPolicy) incrementCurrentInterval() {
	b.currentInterval = time.Duration(float64(b.currentInterval) * b.factor)
}
