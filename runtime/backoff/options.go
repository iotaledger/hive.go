package backoff

import (
	"context"
	"math/rand"
	"time"
)

// An Option configures a BackOff.
type Option interface {
	apply(b Policy) Policy
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(Policy) Policy

func (f optionFunc) apply(p Policy) Policy {
	return f(p)
}

// statelessOptionFunc wraps a function modifying the duration so it satisfies the Option interface.
type statelessOptionFunc func(time.Duration) time.Duration

func (f statelessOptionFunc) apply(p Policy) Policy {
	return &statelessOption{
		delegate: p,
		f:        f,
	}
}

// MaxRetries configures a backoff policy to return Stop if NextBackOff() has been called too many times.
func MaxRetries(max int) Option {
	return optionFunc(func(p Policy) Policy {
		return &maxRetriesOption{
			delegate: p,
			maxTries: max,
			numTries: 0,
		}
	})
}

// MaxInterval configures a backoff policy to not return longer intervals when NextBackOff() is called.
func MaxInterval(maxInterval time.Duration) Option {
	return statelessOptionFunc(func(duration time.Duration) time.Duration {
		if duration > maxInterval {
			return maxInterval
		}

		return duration
	})
}

// Timeout configures a backoff policy to stop when the current time passes the time given with timeout.
func Timeout(timeout time.Time) Option {
	return statelessOptionFunc(func(duration time.Duration) time.Duration {
		if duration == Stop {
			return Stop
		}
		now := time.Now()
		if now.After(timeout) {
			return Stop
		}
		if now.Add(duration).After(timeout) {
			return timeout.Sub(now)
		}

		return duration
	})
}

// Cancel configures a backoff policy to stop if the given context is done.
func Cancel(ctx context.Context) Option {
	return statelessOptionFunc(func(duration time.Duration) time.Duration {
		select {
		case <-ctx.Done():
			return Stop
		default:
		}

		return duration
	})
}

// Jitter configures a backoff policy to randomly modify the duration by the given factor.
// The modified duration is a random value in the interval [randomFactor * duration, duration).
func Jitter(randomFactor float64) Option {
	return statelessOptionFunc(func(duration time.Duration) time.Duration {
		if duration == Stop {
			return Stop
		}
		if randomFactor <= 0 {
			return duration
		}
		delta := randomFactor * float64(duration)

		//nolint:gosec // we do not care about weak random numbers here
		return time.Duration(float64(duration) - rand.Float64()*delta)
	})
}

type statelessOption struct {
	delegate Policy
	f        func(time.Duration) time.Duration
}

func (o *statelessOption) apply(p Policy) Policy {
	return &statelessOption{
		delegate: p,
		f:        o.f,
	}
}

func (o *statelessOption) NextBackOff() time.Duration {
	return o.f(o.delegate.NextBackOff())
}

func (o *statelessOption) New() Policy {
	return &statelessOption{
		delegate: o.delegate.New(),
		f:        o.f,
	}
}

type maxRetriesOption struct {
	delegate Policy
	maxTries int
	numTries int
}

func (b *maxRetriesOption) NextBackOff() time.Duration {
	if b.maxTries == 0 {
		return Stop
	}
	if b.maxTries > 0 {
		if b.maxTries <= b.numTries {
			return Stop
		}
		b.numTries++
	}

	return b.delegate.NextBackOff()
}

func (b *maxRetriesOption) New() Policy {
	return &maxRetriesOption{
		delegate: b.delegate.New(),
		maxTries: b.maxTries,
		numTries: 0,
	}
}
