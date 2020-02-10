package objectstorage

import (
	"time"

	"github.com/dgraph-io/badger/v2"
)

type ObjectStorageOptions struct {
	badgerInstance        *badger.DB
	batchedWriterInstance *BatchedWriter
	cacheTime             time.Duration
	keyPartitions         []int
	persistenceEnabled    bool
	leakDetectionOptions  *LeakDetectionOptions
	leakDetectionWrapper  func(cachedObject *CachedObjectImpl) LeakDetectionWrapper
}

func newObjectStorageOptions(optionalOptions []ObjectStorageOption) *ObjectStorageOptions {
	result := &ObjectStorageOptions{
		cacheTime:          0,
		persistenceEnabled: true,
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	if result.leakDetectionOptions != nil && result.leakDetectionWrapper == nil {
		result.leakDetectionWrapper = newLeakDetectionWrapperImpl
	}

	if result.badgerInstance == nil && result.persistenceEnabled {
		result.badgerInstance = GetBadgerInstance()
	}

	if result.batchedWriterInstance == nil {
		result.batchedWriterInstance = NewBatchedWriter(result.badgerInstance)
	}

	return result
}

type ObjectStorageOption func(*ObjectStorageOptions)

func CacheTime(duration time.Duration) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.cacheTime = duration
	}
}

func BadgerInstance(badgerInstance *badger.DB) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.badgerInstance = badgerInstance
	}
}

func BatchedWriterInstance(batchedWriterInstance *BatchedWriter) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.batchedWriterInstance = batchedWriterInstance
	}
}

func PersistenceEnabled(persistenceEnabled bool) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.persistenceEnabled = persistenceEnabled
	}
}

func LeakDetectionEnabled(leakDetectionEnabled bool, options ...LeakDetectionOptions) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
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
				panic("too many arguments in call to EnableLeakDetection (only 0 or 1 allowed")
			}
		}
	}
}

func OverrideLeakDetectionWrapper(wrapperFunc func(cachedObject *CachedObjectImpl) LeakDetectionWrapper) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.leakDetectionWrapper = wrapperFunc
	}
}

func PartitionKey(keyPartitions ...int) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.keyPartitions = keyPartitions
	}
}
