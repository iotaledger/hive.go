package memstorage

import (
	"sync"

	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

// IndexedStorage is an evictable storage that stores storages for indexes.
type IndexedStorage[IndexType index.Type, K comparable, V any] struct {
	cache *shrinkingmap.ShrinkingMap[IndexType, *shrinkingmap.ShrinkingMap[K, V]]
	mutex sync.Mutex
}

// NewIndexedStorage creates a new indexed storage.
func NewIndexedStorage[IndexType index.Type, K comparable, V any]() *IndexedStorage[IndexType, K, V] {
	return &IndexedStorage[IndexType, K, V]{
		cache: shrinkingmap.New[IndexType, *shrinkingmap.ShrinkingMap[K, V]](),
	}
}

// Evict evicts the storage for the given index.
func (e *IndexedStorage[IndexType, K, V]) Evict(index IndexType) (evictedStorage *shrinkingmap.ShrinkingMap[K, V]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if storage, exists := e.cache.Get(index); exists {
		evictedStorage = storage

		e.cache.Delete(index)
	}

	return
}

// Get returns the storage for the given index.
func (e *IndexedStorage[IndexType, K, V]) Get(index IndexType, createIfMissing ...bool) (storage *shrinkingmap.ShrinkingMap[K, V]) {
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
func (e *IndexedStorage[IndexType, K, V]) ForEach(f func(index IndexType, storage *shrinkingmap.ShrinkingMap[K, V])) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.cache.ForEach(func(index IndexType, storage *shrinkingmap.ShrinkingMap[K, V]) bool {
		f(index, storage)

		return true
	})
}

// Clear clears the storage and returns the cleared elements.
func (e *IndexedStorage[IndexType, K, V]) Clear() (clearedKeys []IndexType, clearedStorages []*shrinkingmap.ShrinkingMap[K, V]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.cache.ForEach(func(index IndexType, storage *shrinkingmap.ShrinkingMap[K, V]) bool {
		clearedKeys = append(clearedKeys, index)
		clearedStorages = append(clearedStorages, storage)

		return true
	})
	e.cache = shrinkingmap.New[IndexType, *shrinkingmap.ShrinkingMap[K, V]]()

	return clearedKeys, clearedStorages
}
