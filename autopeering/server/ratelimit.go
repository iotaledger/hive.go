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

type throttling struct {
	leakyBucketLimit
	tokenBucketLimit
}

type leakyBucketLimit struct {
	ratelimit.Limiter
}

type tokenBucketLimit struct{}

func newLeakyBucket() *throttling {
	return &throttling{
		leakyBucketLimit{
			Limiter: ratelimit.New(MaximumPacketPerSecond),
		},
		tokenBucketLimit{},
	}
}

func newTokenBucket() *throttling {
	return &throttling{
		leakyBucketLimit{},
		tokenBucketLimit{},
	}
}

func newThrottling(s Strategy) *throttling {
	switch s {
	case LeakyBucket:
		return newLeakyBucket()
	case TokenBucket:
		return newTokenBucket()
	default:
		return newLeakyBucket()
	}
}

func (l *leakyBucketLimit) RateLimit() time.Time {
	return l.Take()
}

func (l *tokenBucketLimit) RateLimit() time.Time {
	// TODO()
	return time.Now()
}
