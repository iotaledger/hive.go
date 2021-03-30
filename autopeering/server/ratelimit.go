package server

import (
	"go.uber.org/ratelimit"
	"time"
)

const (
	MaximumPacketPerSecond = 100 // per second
)

const (
	LeakyBucket = iota
	TokenBucket
)

type Strategy uint8

type TrafficShaping interface {
	RateLimit() time.Time
}

type throttlingService struct {
	strategy Strategy
	limiter interface{}
}

// RateLimit take new leaky bucket
func (t *throttlingService) RateLimit() time.Time {
	return t.limiter.(ratelimit.Limiter).Take()
}

func newLeakyBucketLimiter() ratelimit.Limiter {
	limiter := ratelimit.New(MaximumPacketPerSecond)
	return limiter
}

func newTokenBucketLimiter() interface{} {
	// TODO()
	return nil
}

// NewService inits new throttling service
func NewService(strategy Strategy) TrafficShaping {
	ts := new(throttlingService)
	ts.strategy = strategy
	if strategy == LeakyBucket {
		ts.limiter = newLeakyBucketLimiter()
	} else if strategy == TokenBucket {
		ts.limiter = newTokenBucketLimiter()
	}
	return ts
}
