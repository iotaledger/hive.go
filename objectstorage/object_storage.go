package objectstorage

import (
	"errors"
	"sync"
	"unsafe"

	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/timedexecutor"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
)

// ObjectStorage is a manual cache which keeps objects as long as consumers are using it.
type ObjectStorage struct {
	cachedObjects      map[string]interface{}
	cacheMutex         syncutils.RWMutex
	options            *Options
	size               int
	flushMutex         syncutils.RWMutex
	cachedObjectsEmpty sync.WaitGroup
	shutdown           *atomic.Bool
	releaseExecutor    *atomic.UnsafePointer
	partitionsManager  *PartitionsManager

	Events Events
}

type ConsumerFunc = func(key []byte, cachedObject *CachedObjectImpl) bool

// New is the constructor for the ObjectStorage.
func New(store kvstore.KVStore, objectFactory StorableObjectFactory, optionalOptions ...Option) *ObjectStorage {

	storageOptions := newOptions(store, objectFactory, optionalOptions)

	return &ObjectStorage{
		cachedObjects:     make(map[string]interface{}),
		partitionsManager: NewPartitionsManager(),
		options:           storageOptions,
		shutdown:          atomic.NewBool(false),
		releaseExecutor:   atomic.NewUnsafePointer(unsafe.Pointer(timedexecutor.New(storageOptions.releaseExecutorWorkerCount))),

		Events: Events{
			ObjectEvicted: events.NewEvent(evictionEvent),
		},
	}
}

func (objectStorage *ObjectStorage) Put(object StorableObject) CachedObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	return objectStorage.putObjectInCache(object)
}

func (objectStorage *ObjectStorage) Store(object StorableObject) CachedObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	if !objectStorage.options.persistenceEnabled {
		panic("persistence is disabled - use Put(object StorableObject) instead of Store(object StorableObject)")
	}

	object.Persist(true)
	object.SetModified(true)

	return objectStorage.putObjectInCache(object)
}

func (objectStorage *ObjectStorage) GetSize() int {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	objectStorage.flushMutex.RLock()

	objectStorage.cacheMutex.RLock()
	size := objectStorage.size
	objectStorage.cacheMutex.RUnlock()

	objectStorage.flushMutex.RUnlock()

	return size
}

func (objectStorage *ObjectStorage) Get(key []byte) CachedObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if !cacheHit {
		cachedObject.publishResult(nil)
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

func (objectStorage *ObjectStorage) Load(key []byte) CachedObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	if !objectStorage.options.persistenceEnabled {
		panic("persistence is disabled - use Get(object StorableObject) instead of Load(object StorableObject)")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if !cacheHit {
		loadedObject := objectStorage.LoadObjectFromStore(key)
		if !typeutils.IsInterfaceNil(loadedObject) {
			loadedObject.Persist(true)
		}

		cachedObject.publishResult(loadedObject)
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

func (objectStorage *ObjectStorage) Contains(key []byte, options ...ReadOption) (result bool) {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	opts := &ReadOptions{}
	opts.apply(defaultReadOptions...)
	opts.apply(options...)

	if !opts.skipCache {
		if cachedObject, cacheHit := objectStorage.accessCache(key, false); cacheHit {
			defer cachedObject.Release()
			return cachedObject.waitForInitialResult().Exists()
		}
	}

	if !opts.skipStorage {
		return objectStorage.ObjectExistsInStore(key)
	}

	return false
}

func (objectStorage *ObjectStorage) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) StorableObject) CachedObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if cacheHit {
		cachedObject.wg.Wait()

		if cachedObject.updateEmptyResult(func() StorableObject {
			return remappingFunction(key)
		}) {
			cachedObject.storeOnCreation()
		}
	} else {
		loadedObject := objectStorage.LoadObjectFromStore(key)
		if !typeutils.IsInterfaceNil(loadedObject) {
			loadedObject.Persist(true)

			cachedObject.publishResult(loadedObject)
		} else {
			cachedObject.publishResult(remappingFunction(key))
			cachedObject.storeOnCreation()
		}
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

// This method deletes an element and return true if the element was deleted.
func (objectStorage *ObjectStorage) DeleteIfPresent(key []byte) bool {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObjectImpl) bool {
		cachedObject.wg.Wait()

		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			if !storableObject.IsDeleted() {
				storableObject.Delete(true)
				cachedObject.Release(true)

				return true
			}

		}
		cachedObject.Release(true)

		return false
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, false)
	if cacheHit {
		return deleteExistingEntry(cachedObject)
	}

	cachedObject, cacheHit = objectStorage.accessCache(key, true)
	if cacheHit {
		return deleteExistingEntry(cachedObject)
	}

	objectExistsInStore := objectStorage.ObjectExistsInStore(key)
	if objectExistsInStore {
		cachedObject.blindDelete.Store(true)
	}

	cachedObject.publishResult(nil)
	cachedObject.Release(true)

	return objectExistsInStore
}

// DeleteIfPresentAndReturn deletes an element and returns it. If the element does not exist then the return value is
// nil.
func (objectStorage *ObjectStorage) DeleteIfPresentAndReturn(key []byte) StorableObject {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObjectImpl) StorableObject {
		cachedObject.wg.Wait()

		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			if !storableObject.IsDeleted() {
				storableObject.Delete(true)
				cachedObject.Release(true)

				return storableObject
			}

		}
		cachedObject.Release(true)

		return nil
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, false)
	if cacheHit {
		return deleteExistingEntry(cachedObject)
	}

	cachedObject, cacheHit = objectStorage.accessCache(key, true)
	if cacheHit {
		return deleteExistingEntry(cachedObject)
	}

	storableObject := objectStorage.LoadObjectFromStore(key)
	if !typeutils.IsInterfaceNil(storableObject) {
		storableObject.Delete(true)
	}

	cachedObject.publishResult(nil)
	cachedObject.Release(true)

	return storableObject
}

// Performs a "blind delete", where we do not check the objects existence.
// blindDelete is used to delete without accessing the value log.
func (objectStorage *ObjectStorage) Delete(key []byte) {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObjectImpl) {
		cachedObject.wg.Wait()

		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			if !storableObject.IsDeleted() {
				storableObject.Delete(true)
				cachedObject.Release(true)

				return
			}

		}
		cachedObject.Release(true)
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, false)
	if cacheHit {
		deleteExistingEntry(cachedObject)

		return
	}

	cachedObject, cacheHit = objectStorage.accessCache(key, true)
	if cacheHit {
		deleteExistingEntry(cachedObject)

		return
	}

	cachedObject.blindDelete.Store(true)
	cachedObject.publishResult(nil)
	cachedObject.Release(true)
}

// Stores an object only if it was not stored before. In contrast to "ComputeIfAbsent", this method does not access the
// value log. If the object was not stored, then the returned CachedObject is nil and does not need to be Released.
func (objectStorage *ObjectStorage) StoreIfAbsent(object StorableObject) (result CachedObject, stored bool) {
	// abort if the object to store is nil
	if typeutils.IsInterfaceNil(object) {
		return
	}

	// prevent usage of shutdown storage
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	// retrieve object from the cache (without registering a cached object)
	key := object.ObjectStorageKey()
	existingCachedObject, cacheHit := objectStorage.accessCache(key, false)

	// try to update an existing cache entry if it is empty
	if cacheHit {
		cachedObject, updated := objectStorage.updateEmptyCachedObject(existingCachedObject, object)
		if !updated {
			cachedObject.Release()

			return nil, false
		}

		return cachedObject, updated
	}

	// abort if the object already exists in our database
	objectExists := objectStorage.ObjectExistsInStore(key)
	if objectExists {
		return
	}

	// retrieve object from the cache (with registering a cached object)
	existingCachedObject, cacheHit = objectStorage.accessCache(key, true)

	// try to update an existing cache entry if it is empty
	if cacheHit {
		cachedObject, updated := objectStorage.updateEmptyCachedObject(existingCachedObject, object)
		if !updated {
			cachedObject.Release()

			return nil, false
		}

		return cachedObject, updated
	}

	// Abort if the object exists in the database already - an object might have been written and evicted
	// since our last check so even though the object was not found in the cache, it might exists now anyway.
	//
	// Note: We need to fill our registered CachedObject with the actual value from the database instead of just
	//       returning without doing anything.
	if loadedObject := objectStorage.LoadObjectFromStore(key); !typeutils.IsInterfaceNil(loadedObject) {
		existingCachedObject.publishResult(loadedObject)
		existingCachedObject.Release()

		return
	}

	// put object into the prepared cached object
	object.Persist(true)
	object.SetModified(true)
	existingCachedObject.publishResult(object)
	existingCachedObject.storeOnCreation()

	// construct result
	stored = true
	result = wrapCachedObject(existingCachedObject, 0)

	return
}

// ForEach calls the consumer function on every object residing within the cache and the underlying persistence layer.
func (objectStorage *ObjectStorage) ForEach(consumer func(key []byte, cachedObject CachedObject) bool, options ...IteratorOption) {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	opts := &IteratorOptions{}
	opts.apply(defaultIteratorOptions...)
	opts.apply(options...)

	if objectStorage.options.keyPartitions == nil && len(opts.optionalPrefix) > 0 {
		panic("prefix iterations are only allowed when the option PartitionKey(....) is set")
	}

	iterations := 0
	var seenElements map[string]types.Empty
	if !opts.skipCache {
		if len(opts.optionalPrefix) == 0 {
			// iterate over all cached elements
			if seenElements = objectStorage.forEachCachedElement(func(key []byte, cachedObject *CachedObjectImpl) bool {
				iterations++
				if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
					// stop if maximum amount of iterations reached
					cachedObject.Release(true)
					return false
				}
				return consumer(key, wrapCachedObject(cachedObject, 0))
			}); seenElements == nil {
				// Iteration was aborted
				return
			}
		} else {
			// iterate over cached elements via their key partition
			if seenElements = objectStorage.forEachCachedElementWithPrefix(func(key []byte, cachedObject *CachedObjectImpl) bool {
				iterations++
				if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
					// stop if maximum amount of iterations reached
					cachedObject.Release(true)
					return false
				}
				return consumer(key, wrapCachedObject(cachedObject, 0))
			}, opts.optionalPrefix); seenElements == nil {
				// Iteration was aborted
				return
			}
		}
	}

	if opts.skipStorage {
		return
	}

	consumeFunc := func(key kvstore.Key, value kvstore.Value) bool {
		iterations++
		if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
			// stop if maximum amount of iterations reached
			return false
		}

		if _, elementSeen := seenElements[string(key)]; elementSeen {
			return true
		}

		cachedObject, cacheHit := objectStorage.accessCache(key, true)
		if !cacheHit {
			var storableObject StorableObject

			if objectStorage.options.keysOnly {
				var err error
				if storableObject, err = objectStorage.options.objectFactory(key, nil); err != nil {
					return true
				}
			} else {
				marshaledData := make([]byte, len(value))
				copy(marshaledData, value)
				storableObject = objectStorage.unmarshalObject(key, marshaledData)
			}

			if !typeutils.IsInterfaceNil(storableObject) {
				storableObject.Persist(true)
			}

			cachedObject.publishResult(storableObject)
		}

		cachedObject.waitForInitialResult()

		if cachedObject.Exists() && !consumer(key, wrapCachedObject(cachedObject, 0)) {
			// abort iteration
			return false
		}

		return true
	}

	if objectStorage.options.keysOnly {
		_ = objectStorage.options.store.IterateKeys(opts.optionalPrefix, func(key kvstore.Key) bool {
			return consumeFunc(key, []byte{})
		})
		return
	}

	_ = objectStorage.options.store.Iterate(opts.optionalPrefix, consumeFunc)
}

// ForEachKeyOnly calls the consumer function on every storage key residing within the cache and the underlying persistence layer.
func (objectStorage *ObjectStorage) ForEachKeyOnly(consumer func(key []byte) bool, options ...IteratorOption) {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	opts := &IteratorOptions{}
	opts.apply(defaultIteratorOptions...)
	opts.apply(options...)

	if objectStorage.options.keyPartitions == nil && len(opts.optionalPrefix) > 0 {
		panic("prefix iterations are only allowed when the option PartitionKey(....) is set")
	}

	iterations := 0
	var seenElements map[string]types.Empty
	if !opts.skipCache {
		if len(opts.optionalPrefix) == 0 {
			// iterate over all cached elements
			if seenElements = objectStorage.forEachCachedElement(func(key []byte, cachedObject *CachedObjectImpl) bool {
				iterations++
				cachedObject.Release(true)
				if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
					// stop if maximum amount of iterations reached
					return false
				}
				return consumer(key)
			}); seenElements == nil {
				// Iteration was aborted
				return
			}
		} else {
			// iterate over cached elements via their key partition
			if seenElements = objectStorage.forEachCachedElementWithPrefix(func(key []byte, cachedObject *CachedObjectImpl) bool {
				cachedObject.Release(true)
				iterations++
				if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
					// stop if maximum amount of iterations reached
					return false
				}
				return consumer(key)
			}, opts.optionalPrefix); seenElements == nil {
				// Iteration was aborted
				return
			}
		}
	}

	if opts.skipStorage {
		return
	}

	_ = objectStorage.options.store.IterateKeys(opts.optionalPrefix,
		func(key kvstore.Key) bool {
			iterations++
			if (opts.maxIterations != 0) && (iterations > opts.maxIterations) {
				// stop if maximum amount of iterations reached
				return false
			}

			if _, elementSeen := seenElements[string(key)]; elementSeen {
				return true
			}

			// the consumer tells the iterator to abort
			return consumer(key)
		},
	)
}

func (objectStorage *ObjectStorage) Prune() error {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}

	objectStorage.flushMutex.Lock()

	// mark cached objects as evicted
	objectStorage.deepIterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObjectImpl) bool {
		cachedObject.cancelScheduledRelease()
		cachedObject.evicted.Store(true)
		if storableObject := cachedObject.Get(); storableObject != nil {
			storableObject.SetModified(false)
		}

		return true
	})

	objectStorage.cacheMutex.Lock()
	if err := objectStorage.options.store.Clear(); err != nil {
		objectStorage.cacheMutex.Unlock()
		objectStorage.flushMutex.Unlock()
		return err
	}

	objectStorage.cachedObjects = make(map[string]interface{})
	if objectStorage.size != 0 {
		objectStorage.size = 0
		objectStorage.cachedObjectsEmpty.Done()
	}
	objectStorage.cacheMutex.Unlock()

	objectStorage.flushMutex.Unlock()

	return nil
}

func (objectStorage *ObjectStorage) Flush() {
	if objectStorage.shutdown.Load() {
		panic("trying to access shutdown object storage")
	}
	objectStorage.flush(false)
}

func (objectStorage *ObjectStorage) Shutdown() {
	objectStorage.shutdown.Store(true)

	objectStorage.flush(true)

	objectStorage.options.batchedWriterInstance.StopBatchWriter()
}

// FreeMemory copies the content of the internal maps to newly created maps.
// This is necessary, otherwise the GC is not able to free the memory used by the old maps.
// "delete" doesn't shrink the maximum memory used by the map, since it only marks the entry as deleted.
func (objectStorage *ObjectStorage) FreeMemory() {
	objectStorage.flushMutex.RLock()
	defer objectStorage.flushMutex.RUnlock()
	objectStorage.cacheMutex.Lock()
	defer objectStorage.cacheMutex.Unlock()

	// recursively free memory in partitions manager
	if objectStorage.partitionsManager != nil {
		objectStorage.partitionsManager.FreeMemory()
	}

	var deepIterateCacheAndFreeMemory func(sourceMap map[string]interface{}) map[string]interface{}
	deepIterateCacheAndFreeMemory = func(sourceMap map[string]interface{}) map[string]interface{} {
		objectsFound := make(map[string]interface{})
		partitionsFound := make(map[string]map[string]interface{})

		for key, value := range sourceMap {
			if _, cachedObjectReached := value.(*CachedObjectImpl); cachedObjectReached {
				// object level reached
				objectsFound[key] = value
			} else {
				// partition found
				partitionsFound[key] = value.(map[string]interface{})
			}
		}

		if len(partitionsFound) > 0 {
			// partitioned cache
			partitions := make(map[string]interface{})
			for key, partition := range partitionsFound {
				// recursively call for every partition
				partitions[key] = deepIterateCacheAndFreeMemory(partition)
			}
			return partitions
		}

		return objectsFound
	}

	objectStorage.cachedObjects = deepIterateCacheAndFreeMemory(objectStorage.cachedObjects)
}

// ReleaseExecutor returns the executor that schedules releases of CachedObjects after the configured CacheTime.
func (objectStorage *ObjectStorage) ReleaseExecutor() (releaseExecutor *timedexecutor.TimedExecutor) {
	return (*timedexecutor.TimedExecutor)(objectStorage.releaseExecutor.Load())
}

func (objectStorage *ObjectStorage) accessCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObjectImpl, cacheHit bool) {
	objectStorage.flushMutex.RLock()
	defer objectStorage.flushMutex.RUnlock()

	copiedKey := make([]byte, len(key))
	copy(copiedKey, key)

	if objectStorage.options.keyPartitions == nil {
		return objectStorage.accessNonPartitionedCache(copiedKey, createMissingCachedObject)
	}

	return objectStorage.accessPartitionedCache(copiedKey, createMissingCachedObject)
}

func (objectStorage *ObjectStorage) accessNonPartitionedCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObjectImpl, cacheHit bool) {
	objectKey := string(key)
	objectStorage.cacheMutex.RLock()

	currentMap := objectStorage.cachedObjects
	if alreadyCachedObject, cachedObjectExists := currentMap[objectKey]; cachedObjectExists {
		alreadyCachedObject.(*CachedObjectImpl).retain()
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl)
		objectStorage.cacheMutex.RUnlock()
		return
	}
	objectStorage.cacheMutex.RUnlock()

	if !createMissingCachedObject {
		return
	}

	objectStorage.cacheMutex.Lock()
	defer objectStorage.cacheMutex.Unlock()

	// check if the object was created in the meantime
	if alreadyCachedObject, cachedObjectExists := currentMap[objectKey]; cachedObjectExists {
		alreadyCachedObject.(*CachedObjectImpl).retain()
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl)
		return
	}

	// create a new cached object to hold the object
	newlyCachedObject := newCachedObject(objectStorage, key)
	newlyCachedObject.retain()

	if objectStorage.size == 0 {
		objectStorage.cachedObjectsEmpty.Add(1)
	}

	currentMap[objectKey] = newlyCachedObject
	objectStorage.size++
	return newlyCachedObject, false
}

func (objectStorage *ObjectStorage) accessPartitionedCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObjectImpl, cacheHit bool) {
	// acquire read lock so nobody can write to the cache
	objectStorage.cacheMutex.RLock()

	// ensure appropriate lock is unlocked
	var writeLocked bool
	defer func() {
		if writeLocked {
			objectStorage.cacheMutex.Unlock()
		} else {
			objectStorage.cacheMutex.RUnlock()
		}
	}()

	// initialize variables for the loop
	keyPartitionCount := len(objectStorage.options.keyPartitions)
	currentPartition := objectStorage.cachedObjects
	keyOffset := 0
	traversedPartitions := make([]string, 0)

	// loop through partitions up until the object layer
	for i := 0; i < keyPartitionCount-1; i++ {
		// determine the current key segment
		keyPartitionLength := objectStorage.options.keyPartitions[i]
		partitionKey := string(key[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		// if the target partition is found: advance to the next level
		subPartition, subPartitionExists := currentPartition[partitionKey]
		if subPartitionExists {
			currentPartition = subPartition.(map[string]interface{})

			traversedPartitions = append(traversedPartitions, partitionKey)

			continue
		}

		// abort if we are not supposed to create new entries
		if !createMissingCachedObject {
			return
		}

		// switch to write locks and check for existence again
		if !writeLocked {
			objectStorage.partitionsManager.Retain(traversedPartitions)
			// defer in a loop is usually bad, but this only gets called once because we switch to a write locks once
			defer objectStorage.partitionsManager.Release(traversedPartitions)

			objectStorage.cacheMutex.RUnlock()
			objectStorage.cacheMutex.Lock()
			writeLocked = true

			// if the target partition was created while switching locks: advance to the next level
			subPartition, subPartitionExists = currentPartition[partitionKey]
			if subPartitionExists {
				currentPartition = subPartition.(map[string]interface{})

				continue
			}
		}

		// create and advance partition
		subPartition = make(map[string]interface{})
		currentPartition[partitionKey] = subPartition
		currentPartition = subPartition.(map[string]interface{})
	}

	// determine the object key
	keyPartitionLength := objectStorage.options.keyPartitions[keyPartitionCount-1]
	partitionKey := string(key[keyOffset : keyOffset+keyPartitionLength])

	// return if object exists
	if alreadyCachedObject, cachedObjectExists := currentPartition[partitionKey]; cachedObjectExists {
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl).retain().(*CachedObjectImpl)

		return
	}

	// abort if we are not supposed to create new entries
	if !createMissingCachedObject {
		return
	}

	// switch to write locks and check for existence again
	if !writeLocked {
		objectStorage.partitionsManager.Retain(traversedPartitions)
		defer objectStorage.partitionsManager.Release(traversedPartitions)

		objectStorage.cacheMutex.RUnlock()
		objectStorage.cacheMutex.Lock()
		writeLocked = true

		if alreadyCachedObject, cachedObjectExists := currentPartition[partitionKey]; cachedObjectExists {
			cacheHit = true
			cachedObject = alreadyCachedObject.(*CachedObjectImpl).retain().(*CachedObjectImpl)

			return
		}
	}

	// mark objectStorage as non-empty
	if objectStorage.size == 0 {
		objectStorage.cachedObjectsEmpty.Add(1)
	}

	// create a new cached object ...
	cachedObject = newCachedObject(objectStorage, key)
	cachedObject.retain()

	// ... and store it
	currentPartition[partitionKey] = cachedObject
	objectStorage.size++

	return
}

// updateEmptyCachedObject updates the value of the given CachedObject with the given object if and only if the
// CachedObject was empty before. It returns the CachedObject (or nil if it wasn't updated) and a boolean flag indicating
// if the object was updated.
func (objectStorage *ObjectStorage) updateEmptyCachedObject(cachedObject *CachedObjectImpl, object StorableObject) (result CachedObject, updated bool) {
	// wait for the cached object to be available
	cachedObject.waitForInitialResult()

	// prepare object to be stored
	object.Persist(true)
	object.SetModified(true)

	// try to update the object if it is empty or abort otherwise
	updated = cachedObject.updateEmptyResult(object)
	if !updated {
		return cachedObject, updated
	}

	cachedObject.storeOnCreation()

	// construct result
	result = wrapCachedObject(cachedObject, 0)

	return
}

func (objectStorage *ObjectStorage) deleteElementFromCache(key []byte) bool {
	if objectStorage.options.keyPartitions == nil {
		return objectStorage.deleteElementFromUnpartitionedCache(key)
	}

	return objectStorage.deleteElementFromPartitionedCache(key)
}

func (objectStorage *ObjectStorage) deleteElementFromUnpartitionedCache(key []byte) bool {
	_cachedObject, cachedObjectExists := objectStorage.cachedObjects[string(key)]
	if cachedObjectExists {
		delete(objectStorage.cachedObjects, string(key))

		objectStorage.size--

		cachedObject := _cachedObject.(*CachedObjectImpl)
		storableObject := cachedObject.Get()
		if !typeutils.IsInterfaceNil(storableObject) && !storableObject.IsDeleted() {
			objectStorage.Events.ObjectEvicted.Trigger(key, cachedObject.Get())
		}
	}

	return cachedObjectExists
}

func (objectStorage *ObjectStorage) deleteElementFromPartitionedCache(key []byte) (elementExists bool) {
	keyPartitionCount := len(objectStorage.options.keyPartitions)
	keyOffset := 0
	mapStack := make([]map[string]interface{}, 1)
	mapStack[0] = objectStorage.cachedObjects
	traversedPartitions := make([]string, 0)

	// iterate through partitions towards the value
	for keyPartitionId, keyPartitionLength := range objectStorage.options.keyPartitions {
		// retrieve current partition
		currentMap := mapStack[len(mapStack)-1]

		// determine current key segment
		stringKey := string(key[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		// if we didn't arrive at the values, yet
		if keyPartitionId != keyPartitionCount-1 {
			// retrieve next partition
			subMap, subMapExists := currentMap[stringKey]

			// abort if the partition does not exist
			if !subMapExists {
				return
			}

			// advance to the next "level" of partitions
			mapStack = append(mapStack, subMap.(map[string]interface{}))
			traversedPartitions = append(traversedPartitions, stringKey)

			continue
		}

		// check if value exists
		if _, elementExists = currentMap[stringKey]; elementExists {
			// remove value
			delete(currentMap, stringKey)
			objectStorage.size--

			// clean up empty parent partitions (recursively)
			parentKeyPartitionId := keyPartitionId
			keyOffset -= keyPartitionLength
			for parentKeyPartitionId >= 1 && len(mapStack[parentKeyPartitionId]) == 0 {
				if objectStorage.partitionsManager.IsRetained(traversedPartitions[:parentKeyPartitionId]) {
					return
				}

				parentKeyPartitionId--
				parentMap := mapStack[parentKeyPartitionId]
				parentKeyLength := objectStorage.options.keyPartitions[parentKeyPartitionId]

				delete(parentMap, string(key[keyOffset-parentKeyLength:keyOffset]))
				keyOffset -= parentKeyLength
			}
		}
	}

	return
}

func (objectStorage *ObjectStorage) putObjectInCache(object StorableObject) CachedObject {
	// retrieve the cache entry
	cachedObject, cacheHit := objectStorage.accessCache(object.ObjectStorageKey(), true)

	// update and return the object if we have a cache hit
	if cacheHit {
		// try to replace the object if its is empty
		result, updated := objectStorage.updateEmptyCachedObject(cachedObject, object)
		if !updated {
			panic("tried to replace non-empty object in cache")
		}

		return result
	}

	// publish the result to the cached object and return
	cachedObject.publishResult(object)
	cachedObject.storeOnCreation()
	return wrapCachedObject(cachedObject, 0)
}

// LoadObjectFromStore loads a storable object from the persistence layer.
func (objectStorage *ObjectStorage) LoadObjectFromStore(key []byte) StorableObject {
	if !objectStorage.options.persistenceEnabled {
		return nil
	}

	if objectStorage.options.keysOnly {
		contains, err := objectStorage.options.store.Has(key)
		if err != nil {
			// No need to check for kvstore.ErrKeyNotFound here
			panic(err)
		}

		if !contains {
			return nil
		}

		object, err := objectStorage.options.objectFactory(key, nil)
		if err != nil {
			panic(err)
		}

		return object
	}

	var marshaledData []byte
	value, err := objectStorage.options.store.Get(key)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil
		}

		panic(err)
	}

	marshaledData = make([]byte, len(value))
	copy(marshaledData, value)

	return objectStorage.unmarshalObject(key, marshaledData)
}

// DeleteEntryFromStore deletes an entry from the persistence layer.
func (objectStorage *ObjectStorage) DeleteEntryFromStore(key []byte) {
	if !objectStorage.options.persistenceEnabled {
		return
	}

	if err := objectStorage.options.store.Delete(key); err != nil {
		if !errors.Is(err, kvstore.ErrKeyNotFound) {
			panic(err)
		}
	}
}

// DeleteEntriesFromStore deletes entries from the persistence layer.
func (objectStorage *ObjectStorage) DeleteEntriesFromStore(keys [][]byte) {
	if !objectStorage.options.persistenceEnabled {
		return
	}

	batchedMuts := objectStorage.options.store.Batched()
	for i := 0; i < len(keys); i++ {
		if err := batchedMuts.Delete(keys[i]); err != nil {
			batchedMuts.Cancel()
			panic(err)
		}
	}

	if err := batchedMuts.Commit(); err != nil {
		panic(err)
	}
}

func (objectStorage *ObjectStorage) ObjectExistsInStore(key []byte) bool {
	if !objectStorage.options.persistenceEnabled {
		return false
	}

	has, err := objectStorage.options.store.Has(key)
	if err != nil {
		if !errors.Is(err, kvstore.ErrKeyNotFound) {
			panic(err)
		}
	}
	return has
}

func (objectStorage *ObjectStorage) unmarshalObject(key []byte, data []byte) StorableObject {
	object, err := objectStorage.options.objectFactory(key, data)
	if err != nil {
		panic(err)
	}

	return object
}

func (objectStorage *ObjectStorage) flush(shutdown bool) {
	objectStorage.flushMutex.Lock()

	// cancel all pending release tasks (we flush manually) and create a new executor if we didn't shut down
	objectStorage.ReleaseExecutor().Shutdown(timedexecutor.CancelPendingTasks, timedexecutor.DontWaitForShutdown)
	if !shutdown {
		objectStorage.releaseExecutor.Store(unsafe.Pointer(timedexecutor.New(objectStorage.options.releaseExecutorWorkerCount)))
	}

	// create a list of objects that shall be flushed (so the BatchWriter can access the cachedObjects mutex and delete)
	cachedObjects := make([]*CachedObjectImpl, objectStorage.size)
	var i int
	objectStorage.deepIterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObjectImpl) bool {
		cachedObject.scheduledTask.Store(nil)

		cachedObjects[i] = cachedObject
		i++

		return true
	})

	objectStorage.flushMutex.Unlock()

	// force release the collected objects
	for j := 0; j < i; j++ {
		if consumers := cachedObjects[j].consumers.Dec(); consumers == 0 {
			cachedObjects[j].evict()
		}
	}

	objectStorage.options.batchedWriterInstance.Flush()
	objectStorage.cachedObjectsEmpty.Wait()
}

// iterates over all cached objects and calls the consumer function on them.
func (objectStorage *ObjectStorage) deepIterateThroughCachedElements(sourceMap map[string]interface{}, consumer ConsumerFunc) bool {
	// We first iterate through the target objects and collect them in a temporary list, so we can release the ReadLock
	// as fast as possible. This allows us to call the consumers without the ReadLock being set (which avoids
	// deadlocks by consumers that i.e. try to issue a force release).
	objectStorage.cacheMutex.RLock()
	foundObjects := make([]*CachedObjectImpl, len(sourceMap))
	foundObjectsCounter := 0
	foundPartitions := make([]map[string]interface{}, len(sourceMap))
	foundPartitionsCounter := 0
	for _, value := range sourceMap {
		if cachedObject, cachedObjectReached := value.(*CachedObjectImpl); cachedObjectReached {
			cachedObject.retain()

			foundObjects[foundObjectsCounter] = cachedObject
			foundObjectsCounter++
		} else {
			foundPartitions[foundPartitionsCounter] = value.(map[string]interface{})
			foundPartitionsCounter++
		}
	}
	objectStorage.cacheMutex.RUnlock()

	// The founds objects can not be removed in the mean time since we have retained them already. This means, that we
	// can safely iterate through them and call the corresponding consumer.
	aborted := false
	for i := 0; i < foundObjectsCounter; i++ {
		// release the previously retained objects after we have detected an abort
		if aborted {
			foundObjects[i].Release()

			continue
		}

		// Call consumer with the cached object and check if we should abort the iteration.
		cachedObject := foundObjects[i]
		cachedObject.waitForInitialResult()
		if !consumer(cachedObject.key, cachedObject) {
			aborted = true
		}
	}
	if aborted {
		return false
	}

	// It could in theory happen, that the found partition got cleaned up in between storing it in our list and
	// iterating through it, but this is not a problem, because then the map will be empty in the consecutive call to
	// deepIterateThroughCachedElements and we will ignore the previously deleted elements.
	for i := 0; i < foundPartitionsCounter; i++ {
		if !objectStorage.deepIterateThroughCachedElements(foundPartitions[i], consumer) {
			// Iteration was aborted
			return false
		}
	}

	return true
}

// calls the consumer function on every object within the cache and returns a set of seen keys.
func (objectStorage *ObjectStorage) forEachCachedElement(consumer ConsumerFunc) map[string]types.Empty {
	seenElements := make(map[string]types.Empty)
	if !objectStorage.deepIterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObjectImpl) bool {
		seenElements[string(cachedObject.key)] = types.Void

		if !cachedObject.Exists() {
			cachedObject.Release()

			return true
		}

		return consumer(key, cachedObject)
	}) {
		// Iteration was aborted
		return nil
	}

	return seenElements
}

func (objectStorage *ObjectStorage) forEachCachedElementWithPrefix(consumer ConsumerFunc, prefix []byte) map[string]types.Empty {
	seenElements := make(map[string]types.Empty)

	prefixLength := len(prefix)
	keyPartitions := objectStorage.options.keyPartitions
	partitionCount := len(keyPartitions)
	currentPartition := objectStorage.cachedObjects
	keyOffset := 0

	for i, partitionKeyLength := range keyPartitions {
		// if the keyOffset equals the prefixLength, we're at the wanted layer
		if keyOffset == prefixLength {
			break
		}

		if keyOffset+partitionKeyLength > prefixLength {
			panic("the prefix length does not align with the set KeyPartitions")
		}

		partitionKey := prefix[keyOffset : keyOffset+partitionKeyLength]
		keyOffset += partitionKeyLength

		// advance partitions as long as we don't hit the object layer
		if i != partitionCount-1 {
			objectStorage.cacheMutex.RLock()
			subPartition, subPartitionExists := currentPartition[string(partitionKey)]
			objectStorage.cacheMutex.RUnlock()
			if !subPartitionExists {
				// no partition exists for the given prefix
				return seenElements
			}
			currentPartition = subPartition.(map[string]interface{})
			continue
		}

		if keyOffset < prefixLength {
			panic("the prefix is too long for the set KeyPartition")
		}

		objectStorage.cacheMutex.RLock()
		cachedObject := currentPartition[string(partitionKey)].(*CachedObjectImpl)
		objectStorage.cacheMutex.RUnlock()
		cachedObject.waitForInitialResult()
		seenElements[string(cachedObject.key)] = types.Void

		// the given prefix references a partition
		if !cachedObject.Exists() {
			continue
		}

		cachedObject.retain()
		if !consumer(cachedObject.key, cachedObject) {
			// Iteration was aborted
			return nil
		}

		return seenElements
	}

	// deep-iterate over the current partition and call the consumer function on each object
	if !objectStorage.deepIterateThroughCachedElements(currentPartition, func(key []byte, cachedObject *CachedObjectImpl) bool {
		seenElements[string(cachedObject.key)] = types.Void

		if !cachedObject.Exists() {
			cachedObject.Release()

			return true
		}

		return consumer(key, cachedObject)
	}) {
		// Iteration was aborted
		return nil
	}

	if keyOffset > prefixLength {
		panic("the prefix length does not align with the set KeyPartitions")
	}

	return seenElements
}

// StorableObjectFactory is used to address the factory method that generically creates StorableObjects. It receives the
// key and the serialized data of the object and returns an "empty" StorableObject that just has its key set. The object
// is then fully unmarshaled by the ObjectStorage which calls the UnmarshalObjectStorageValue with the data. The data
// is anyway provided in this method already to allow the dynamic creation of different object types depending on the
// stored data.
type StorableObjectFactory func(key []byte, data []byte) (result StorableObject, err error)
