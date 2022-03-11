package objectstorage

import (
	"time"

	"github.com/iotaledger/hive.go/kvstore/debug"
	"github.com/iotaledger/hive.go/objectstorage"
)

type (
	Option          = objectstorage.Option
	Options         = objectstorage.Options
	ReadOption      = objectstorage.ReadOption
	ReadOptions     = objectstorage.ReadOptions
	IteratorOption  = objectstorage.IteratorOption
	IteratorOptions = objectstorage.IteratorOptions
)

func WithObjectFactory(objectFactory StorableObjectFactory) Option {
	return objectstorage.WithObjectFactory(func(key []byte, data []byte) (result objectstorage.StorableObject, err error) {
		return objectFactory(key, data)
	})
}

func CacheTime(duration time.Duration) Option {
	return objectstorage.CacheTime(duration)
}

func LogAccess(fileName string, commandsFilter ...debug.Command) Option {
	return objectstorage.LogAccess(fileName, commandsFilter...)
}

func PersistenceEnabled(persistenceEnabled bool) Option {
	return objectstorage.PersistenceEnabled(persistenceEnabled)
}

func KeysOnly(keysOnly bool) Option {
	return objectstorage.KeysOnly(keysOnly)
}

func StoreOnCreation(store bool) Option {
	return objectstorage.StoreOnCreation(store)
}

func ReleaseExecutorWorkerCount(releaseExecutorWorkerCount int) Option {
	return objectstorage.ReleaseExecutorWorkerCount(releaseExecutorWorkerCount)
}

func LeakDetectionEnabled(leakDetectionEnabled bool, options ...LeakDetectionOptions) Option {
	return objectstorage.LeakDetectionEnabled(leakDetectionEnabled, options...)
}

func OverrideLeakDetectionWrapper[T StorableObject](wrapperFunc func(cachedObject *CachedObject[T]) LeakDetectionWrapper) Option {
	return objectstorage.OverrideLeakDetectionWrapper(func(cachedObject *objectstorage.CachedObjectImpl) objectstorage.LeakDetectionWrapper {
		return wrapperFunc(newCachedObject[T](cachedObject))
	})
}

func PartitionKey(keyPartitions ...int) Option {
	return objectstorage.PartitionKey(keyPartitions...)
}

func OnEvictionCallback[T StorableObject](cb func(cachedObject *CachedObject[T])) Option {
	return objectstorage.OnEvictionCallback(func(cachedObject objectstorage.CachedObject) {
		cb(newCachedObject[T](cachedObject))
	})
}

func WithReadSkipCache(skipCache bool) ReadOption {
	return objectstorage.WithReadSkipCache(skipCache)
}

func WithReadSkipStorage(skipStorage bool) ReadOption {
	return objectstorage.WithReadSkipStorage(skipStorage)
}

func WithIteratorSkipCache(skipCache bool) IteratorOption {
	return objectstorage.WithIteratorSkipCache(skipCache)
}

func WithIteratorSkipStorage(skipStorage bool) IteratorOption {
	return objectstorage.WithIteratorSkipStorage(skipStorage)
}

func WithIteratorPrefix(prefix []byte) IteratorOption {
	return objectstorage.WithIteratorPrefix(prefix)
}

func WithIteratorMaxIterations(maxIterations int) IteratorOption {
	return objectstorage.WithIteratorMaxIterations(maxIterations)
}
