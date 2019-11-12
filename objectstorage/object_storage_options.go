package objectstorage

import (
	"time"
)

type ObjectStorageOptions struct {
	cacheTime          time.Duration
	persistenceEnabled bool
}

func newTransportOutputStorageFilters(optionalOptions []ObjectStorageOption) *ObjectStorageOptions {
	result := &ObjectStorageOptions{
		cacheTime: 0,
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	return result
}

type ObjectStorageOption func(*ObjectStorageOptions)

func CacheTime(duration time.Duration) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.cacheTime = duration
	}
}

func PersistenceEnabled(persistenceEnabled bool) ObjectStorageOption {
	return func(args *ObjectStorageOptions) {
		args.persistenceEnabled = persistenceEnabled
	}
}
