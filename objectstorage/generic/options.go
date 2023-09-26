package generic

import (
	"time"

	"github.com/izuc/zipp.foundation/kvstore/debug"
	"github.com/izuc/zipp.foundation/objectstorage"
)

type (
	Option          = objectstorage.Option
	Options         = objectstorage.Options
	ReadOption      = objectstorage.ReadOption
	ReadOptions     = objectstorage.ReadOptions
	IteratorOption  = objectstorage.IteratorOption
	IteratorOptions = objectstorage.IteratorOptions
)

// CacheTime sets the time after which the object is evicted from the cache.
func CacheTime(duration time.Duration) Option {
	return objectstorage.CacheTime(duration)
}

// LogAccess sets up a logger that logs all calls to the underlying store in the given file. It is possible to filter
// the logged commands by providing an optional filter flag.
func LogAccess(fileName string, commandsFilter ...debug.Command) Option {
	return objectstorage.LogAccess(fileName, commandsFilter...)
}

// PersistenceEnabled enables the persistence of the object storage.
func PersistenceEnabled(persistenceEnabled bool) Option {
	return objectstorage.PersistenceEnabled(persistenceEnabled)
}

// KeysOnly is used to store only the keys of the elements.
func KeysOnly(keysOnly bool) Option {
	return objectstorage.KeysOnly(keysOnly)
}

// StoreOnCreation writes an object directly to the persistence layer on creation.
func StoreOnCreation(store bool) Option {
	return objectstorage.StoreOnCreation(store)
}

// ReleaseExecutorWorkerCount sets the number of workers that execute the
// scheduled eviction of the objects in parallel (whenever they become due).
func ReleaseExecutorWorkerCount(releaseExecutorWorkerCount int) Option {
	return objectstorage.ReleaseExecutorWorkerCount(releaseExecutorWorkerCount)
}

// LeakDetectionEnabled enables the leak detection of the object storage.
func LeakDetectionEnabled(leakDetectionEnabled bool, options ...LeakDetectionOptions) Option {
	return objectstorage.LeakDetectionEnabled(leakDetectionEnabled, options...)
}

// OverrideLeakDetectionWrapper is used to override the default leak detection wrapper.
func OverrideLeakDetectionWrapper[T StorableObject](wrapperFunc func(cachedObject *CachedObject[T]) LeakDetectionWrapper) Option {
	return objectstorage.OverrideLeakDetectionWrapper(func(cachedObject *objectstorage.CachedObjectImpl) objectstorage.LeakDetectionWrapper {
		return wrapperFunc(newCachedObject[T](cachedObject))
	})
}

// PartitionKey sets the partition sizes of the key.
func PartitionKey(keyPartitions ...int) Option {
	return objectstorage.PartitionKey(keyPartitions...)
}

// OnEvictionCallback sets a function that is called on eviction of the object.
func OnEvictionCallback[T StorableObject](cb func(cachedObject *CachedObject[T])) Option {
	return objectstorage.OnEvictionCallback(func(cachedObject objectstorage.CachedObject) {
		cb(newCachedObject[T](cachedObject))
	})
}

// WithReadSkipCache is used to skip the elements in the cache.
func WithReadSkipCache(skipCache bool) ReadOption {
	return objectstorage.WithReadSkipCache(skipCache)
}

// WithReadSkipStorage is used to skip the elements in the storage.
func WithReadSkipStorage(skipStorage bool) ReadOption {
	return objectstorage.WithReadSkipStorage(skipStorage)
}

// WithIteratorSkipCache is used to skip the elements in the cache.
func WithIteratorSkipCache(skipCache bool) IteratorOption {
	return objectstorage.WithIteratorSkipCache(skipCache)
}

// WithIteratorSkipStorage is used to skip the elements in the storage.
func WithIteratorSkipStorage(skipStorage bool) IteratorOption {
	return objectstorage.WithIteratorSkipStorage(skipStorage)
}

// WithIteratorPrefix is used to iterate a subset of elements with a defined prefix.
func WithIteratorPrefix(prefix []byte) IteratorOption {
	return objectstorage.WithIteratorPrefix(prefix)
}

// WithIteratorMaxIterations is used to stop the iteration after a certain amount of iterations.
// 0 disables the limit.
func WithIteratorMaxIterations(maxIterations int) IteratorOption {
	return objectstorage.WithIteratorMaxIterations(maxIterations)
}
