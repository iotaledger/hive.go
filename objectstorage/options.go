package objectstorage

import (
	"time"
)

type Options struct {
	batchedWriterInstance *BatchedWriter
	cacheTime             time.Duration
	keyPartitions         []int
	persistenceEnabled    bool
	keysOnly              bool
	leakDetectionOptions  *LeakDetectionOptions
	leakDetectionWrapper  func(cachedObject *CachedObjectImpl) LeakDetectionWrapper
}

func newOptions(objectStorage *ObjectStorage, optionalOptions []Option) *Options {
	result := &Options{
		cacheTime:          0,
		persistenceEnabled: true,
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	if result.leakDetectionOptions != nil && result.leakDetectionWrapper == nil {
		result.leakDetectionWrapper = newLeakDetectionWrapperImpl
	}

	if result.batchedWriterInstance == nil {
		result.batchedWriterInstance = NewBatchedWriter(objectStorage.store)
	}

	return result
}

type Option func(*Options)

func CacheTime(duration time.Duration) Option {
	return func(args *Options) {
		args.cacheTime = duration
	}
}

func BatchedWriterInstance(batchedWriterInstance *BatchedWriter) Option {
	return func(args *Options) {
		args.batchedWriterInstance = batchedWriterInstance
	}
}

func PersistenceEnabled(persistenceEnabled bool) Option {
	return func(args *Options) {
		args.persistenceEnabled = persistenceEnabled
	}
}

func KeysOnly(keysOnly bool) Option {
	return func(args *Options) {
		args.keysOnly = keysOnly
	}
}

func LeakDetectionEnabled(leakDetectionEnabled bool, options ...LeakDetectionOptions) Option {
	return func(args *Options) {
		if leakDetectionEnabled {
			switch len(options) {
			case 0:
				args.leakDetectionOptions = &LeakDetectionOptions{
					MaxConsumersPerObject: 20,
					MaxConsumerHoldTime:   240 * time.Second,
				}
			case 1:
				args.leakDetectionOptions = &options[0]
			default:
				panic("too many additional arguments in call to LeakDetectionEnabled (only 0 or 1 allowed")
			}
		}
	}
}

func OverrideLeakDetectionWrapper(wrapperFunc func(cachedObject *CachedObjectImpl) LeakDetectionWrapper) Option {
	return func(args *Options) {
		args.leakDetectionWrapper = wrapperFunc
	}
}

func PartitionKey(keyPartitions ...int) Option {
	return func(args *Options) {
		args.keyPartitions = keyPartitions
	}
}
