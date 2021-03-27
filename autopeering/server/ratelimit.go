package server

import (
	leakyBucket "go.uber.org/ratelimit"
)

const (
	MaximumPacketPerSecond = 3 // per second
)

type LeakyBucketLimit struct {
	lB leakyBucket.Limiter
}

func newLeakyBucket() *LeakyBucketLimit {
	return &LeakyBucketLimit{
		lB: leakyBucket.New(MaximumPacketPerSecond),
	}
}