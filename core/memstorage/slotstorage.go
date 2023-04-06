package memstorage

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

type Index = int64

// IndexedStorage is an evictable storage that stores storages for indexes.
type IndexedStorage[K comparable, V any] struct {
	cache *shrinkingmap.ShrinkingMap[Index, *shrinkingmap.ShrinkingMap[K, V]]
	mutex sync.Mutex
}

// NewSlotStorage creates a new slot storage.
func NewSlotStorage[K comparable, V any]() *IndexedStorage[K, V] {
	return &IndexedStorage[K, V]{
		cache: shrinkingmap.New[Index, *shrinkingmap.ShrinkingMap[K, V]](),
	}
}

// Evict evicts the storage for the given index.
func (e *IndexedStorage[K, V]) Evict(index Index) (evictedStorage *shrinkingmap.ShrinkingMap[K, V]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if storage, exists := e.cache.Get(index); exists {
		evictedStorage = storage

		e.cache.Delete(index)
	}

	return
}

// Get returns the storage for the given index.
func (e *IndexedStorage[K, V]) Get(index Index, createIfMissing ...bool) (storage *shrinkingmap.ShrinkingMap[K, V]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	storage, exists := e.cache.Get(index)
	if exists {
		return storage
	}

	if len(createIfMissing) == 0 || !createIfMissing[0] {
		return nil
	}

	storage = shrinkingmap.New[K, V]()
	e.cache.Set(index, storage)

	return storage
}

// ForEach iterates over all storages.
func (e *IndexedStorage[K, V]) ForEach(f func(index Index, storage *shrinkingmap.ShrinkingMap[K, V])) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.cache.ForEach(func(index Index, storage *shrinkingmap.ShrinkingMap[K, V]) bool {
		f(index, storage)

		return true
	})
}
