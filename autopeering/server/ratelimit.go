package server

import (
	leakyBucket "go.uber.org/ratelimit"
	"sync"
	"time"
)

const (
	MaximumPacketPerSecond = 3 // per second
)

type RateLimit interface {
	NewLeakyBucket() leakyBucket.Limiter
	TakeLeakyBucket(lB leakyBucket.Limiter) time.Time
	TokenBucket()
}

type rateLimit struct{
	RateLimit
	sync.RWMutex
}

func (r *rateLimit) NewLeakyBucket() leakyBucket.Limiter {
	return leakyBucket.New(MaximumPacketPerSecond)
}

func (r *rateLimit) TakeLeakyBucket(lB leakyBucket.Limiter) time.Time {
	return lB.Take()
}

func (r *rateLimit) TokenBucket() {
	panic("implement me")
}