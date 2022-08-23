package lrucache

import (
	"time"
)

type Options struct {
	EvictionCallback  func(keyOrBatchedKeys interface{}, valueOrBatchedValues interface{})
	EvictionBatchSize uint64
	IdleTimeout       time.Duration
}

var defaultOptions = &Options{
	EvictionCallback:  nil,
	EvictionBatchSize: 1,
	IdleTimeout:       30 * time.Second,
}
