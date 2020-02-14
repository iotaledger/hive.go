package objectstorage

import (
	"sync"

	"github.com/dgraph-io/badger/v2"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
)

// ObjectStorage is a manual cache which keeps objects as long as consumers are using it.
type ObjectStorage struct {
	storageId          []byte
	objectFactory      StorableObjectFactory
	cachedObjects      map[string]interface{}
	cacheMutex         syncutils.RWMutex
	options            *ObjectStorageOptions
	size               int
	flushMutex         syncutils.RWMutex
	cachedObjectsEmpty sync.WaitGroup
	shutdown           typeutils.AtomicBool

	Events Events
}

type ConsumerFunc = func(key []byte, cachedObject *CachedObjectImpl) bool

func New(storageId []byte, objectFactory StorableObjectFactory, optionalOptions ...ObjectStorageOption) *ObjectStorage {
	return &ObjectStorage{
		storageId:     storageId,
		objectFactory: objectFactory,
		cachedObjects: make(map[string]interface{}),
		options:       newObjectStorageOptions(optionalOptions),

		Events: Events{
			ObjectEvicted: events.NewEvent(evictionEvent),
		},
	}
}

func (objectStorage *ObjectStorage) Put(object StorableObject) CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	return wrapCachedObject(objectStorage.putObjectInCache(object), 0)
}

func (objectStorage *ObjectStorage) Store(object StorableObject) CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if !objectStorage.options.persistenceEnabled {
		panic("persistence is disabled - use Put(object StorableObject) instead of Store(object StorableObject)")
	}

	object.Persist()
	object.SetModified()

	return wrapCachedObject(objectStorage.putObjectInCache(object), 0)
}

func (objectStorage *ObjectStorage) GetSize() int {
	if objectStorage.shutdown.IsSet() {
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
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if !cacheHit {
		cachedObject.publishResult(nil)
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

func (objectStorage *ObjectStorage) Load(key []byte) CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if !objectStorage.options.persistenceEnabled {
		panic("persistence is disabled - use Get(object StorableObject) instead of Load(object StorableObject)")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if !cacheHit {
		loadedObject := objectStorage.loadObjectFromBadger(key)
		if loadedObject != nil {
			loadedObject.Persist()
		}

		cachedObject.publishResult(loadedObject)
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

func (objectStorage *ObjectStorage) Contains(key []byte) (result bool) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if cachedObject, cacheHit := objectStorage.accessCache(key, false); cacheHit {
		result = cachedObject.waitForInitialResult().Exists()

		cachedObject.Release()
	} else {
		result = objectStorage.objectExistsInBadger(key)
	}

	return
}

func (objectStorage *ObjectStorage) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) StorableObject) CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if cacheHit {
		cachedObject.wg.Wait()

		cachedObject.updateEmptyResult(func() StorableObject {
			return remappingFunction(key)
		})
	} else {
		loadedObject := objectStorage.loadObjectFromBadger(key)
		if loadedObject != nil {
			loadedObject.Persist()

			cachedObject.publishResult(loadedObject)
		} else {
			cachedObject.publishResult(remappingFunction(key))
		}
	}

	return wrapCachedObject(cachedObject.waitForInitialResult(), 0)
}

// This method deletes an element and return true if the element was deleted.
func (objectStorage *ObjectStorage) DeleteIfPresent(key []byte) bool {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObjectImpl) bool {
		cachedObject.wg.Wait()

		if storableObject := cachedObject.Get(); storableObject != nil {
			if !storableObject.IsDeleted() {
				storableObject.Delete()
				cachedObject.Release()

				return true
			}

			cachedObject.Release()
		}

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

	if objectStorage.objectExistsInBadger(key) {
		cachedObject.blindDelete.Set()
		cachedObject.publishResult(nil)
		cachedObject.Release()

		return true
	}

	cachedObject.Release()

	return false
}

// Performs a "blind delete", where we do not check the objects existence.
func (objectStorage *ObjectStorage) Delete(key []byte) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObjectImpl) {
		cachedObject.wg.Wait()

		if storableObject := cachedObject.Get(); storableObject != nil {
			if !storableObject.IsDeleted() {
				storableObject.Delete()
				cachedObject.Release()

				return
			}

			cachedObject.Release()
		}
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

	cachedObject.blindDelete.Set()
	cachedObject.publishResult(nil)
	cachedObject.Release()
}

// Stores an object only if it was not stored before. In contrast to "ComputeIfAbsent", this method does not access the
// value log. If the object was not stored, then the returned CachedObject is nil and does not need to be Released.
func (objectStorage *ObjectStorage) StoreIfAbsent(object StorableObject) (result CachedObject, stored bool) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	var cachedObject *CachedObjectImpl
	key := object.GetStorageKey()

	existingCachedObject, cacheHit := objectStorage.accessCache(key, false)
	if cacheHit {
		existingCachedObject.wg.Wait()

		object.Persist()
		object.SetModified()

		if stored = existingCachedObject.updateEmptyResult(object); stored {
			cachedObject = existingCachedObject
		} else {
			existingCachedObject.Release()
		}
	} else {
		if objectExists := objectStorage.objectExistsInBadger(key); !objectExists {
			object.Persist()
			object.SetModified()

			if newCachedObject, cacheHit := objectStorage.accessCache(key, true); cacheHit {
				newCachedObject.wg.Wait()

				if stored = newCachedObject.updateEmptyResult(object); stored {
					cachedObject = newCachedObject
				} else {
					newCachedObject.Release()
				}
			} else {
				newCachedObject.publishResult(object)

				stored = true
				cachedObject = newCachedObject
			}
		}
	}

	if cachedObject != nil {
		result = wrapCachedObject(cachedObject.waitForInitialResult(), 0)
	}

	return
}

// ForEach calls the consumer function on every object residing within the cache and the underlying persistence layer.
func (objectStorage *ObjectStorage) ForEach(consumer func(key []byte, cachedObject CachedObject) bool, optionalPrefix ...[]byte) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if objectStorage.options.keyPartitions == nil && len(optionalPrefix) >= 1 {
		panic("prefix iterations are only allowed when the option PartitionKey(....) is set")
	}

	var seenElements map[string]types.Empty
	if len(optionalPrefix) == 0 || len(optionalPrefix[0]) == 0 {
		// iterate over all cached elements
		if seenElements = objectStorage.forEachCachedElement(func(key []byte, cachedObject *CachedObjectImpl) bool {
			return consumer(key, wrapCachedObject(cachedObject, 0))
		}); seenElements == nil {
			return
		}
	} else {
		// iterate over cached elements via their key partition
		if seenElements = objectStorage.forEachCachedElementWithPrefix(func(key []byte, cachedObject *CachedObjectImpl) bool {
			return consumer(key, wrapCachedObject(cachedObject, 0))
		}, optionalPrefix[0]); seenElements == nil {
			return
		}
	}

	if err := objectStorage.options.badgerInstance.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = objectStorage.generatePrefix(optionalPrefix)

		it := txn.NewIterator(iteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()[len(objectStorage.storageId):]

			if _, elementSeen := seenElements[typeutils.BytesToString(key)]; elementSeen {
				continue
			}

			cachedObject, cacheHit := objectStorage.accessCache(key, true)
			if !cacheHit {
				if err := item.Value(func(val []byte) error {
					marshaledData := make([]byte, len(val))
					copy(marshaledData, val)

					storableObject := objectStorage.unmarshalObject(key, marshaledData)
					if storableObject != nil {
						storableObject.Persist()
					}

					cachedObject.publishResult(storableObject)

					return nil
				}); err != nil {
					panic(err)
				}
			}

			cachedObject.waitForInitialResult()

			if cachedObject.Exists() && !consumer(key, wrapCachedObject(cachedObject, 0)) {
				break
			}
		}
		it.Close()

		return nil
	}); err != nil {
		panic(err)
	}
}

func (objectStorage *ObjectStorage) Prune() error {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	objectStorage.flushMutex.Lock()

	objectStorage.cacheMutex.Lock()
	if err := objectStorage.options.badgerInstance.DropPrefix(objectStorage.storageId); err != nil {
		objectStorage.cacheMutex.Unlock()
		objectStorage.flushMutex.Unlock()

		return err
	}
	objectStorage.cachedObjects = make(map[string]interface{})
	objectStorage.cacheMutex.Unlock()

	objectStorage.flushMutex.Unlock()

	return nil
}

func (objectStorage *ObjectStorage) Flush() {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	objectStorage.flushMutex.Lock()
	objectStorage.flush()
	objectStorage.flushMutex.Unlock()
}

func (objectStorage *ObjectStorage) Shutdown() {
	objectStorage.shutdown.Set()

	objectStorage.flush()
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
	objectKey := typeutils.BytesToString(key)
	currentMap := objectStorage.cachedObjects

	objectStorage.cacheMutex.RLock()
	if alreadyCachedObject, cachedObjectExists := currentMap[objectKey]; cachedObjectExists {
		alreadyCachedObject.(*CachedObjectImpl).Retain()
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
		alreadyCachedObject.(*CachedObjectImpl).Retain()
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl)
		return
	}

	// create a new cached object to hold the object
	newlyCachedObject := newCachedObject(objectStorage, key)
	newlyCachedObject.Retain()

	if objectStorage.size == 0 {
		objectStorage.cachedObjectsEmpty.Add(1)
	}

	currentMap[objectKey] = newlyCachedObject
	objectStorage.size++
	return newlyCachedObject, false
}

func (objectStorage *ObjectStorage) accessPartitionedCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObjectImpl, cacheHit bool) {
	keyPartitions := objectStorage.options.keyPartitions
	currentPartition := objectStorage.cachedObjects
	keyOffset := 0

	objectStorage.cacheMutex.RLock()
	var writeLocked bool

	// advance and create partitions up until the object layer
	var partitionKey string
	for i := 0; i < len(keyPartitions)-1; i++ {
		keyPartitionLength := keyPartitions[i]
		partitionKey = typeutils.BytesToString(key[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		subPartition, subPartitionExists := currentPartition[partitionKey]
		if subPartitionExists {
			currentPartition = subPartition.(map[string]interface{})
			continue
		}

		if !createMissingCachedObject {
			objectStorage.cacheMutex.RUnlock()
			return
		}

		if !writeLocked {
			objectStorage.cacheMutex.RUnlock()
			objectStorage.cacheMutex.Lock()
			writeLocked = true
		}

		subPartition, subPartitionExists = currentPartition[partitionKey]
		if subPartitionExists {
			currentPartition = subPartition.(map[string]interface{})
			continue
		}

		// create and advance partition
		subPartition = make(map[string]interface{})
		currentPartition[partitionKey] = subPartition
		currentPartition = subPartition.(map[string]interface{})
	}

	// ensure appropriate lock is unlocked
	defer func() {
		if writeLocked {
			objectStorage.cacheMutex.Unlock()
		} else {
			objectStorage.cacheMutex.RUnlock()
		}
	}()

	// grab the object key
	objectLayer := currentPartition
	objectKey := typeutils.BytesToString(key[keyOffset:])
	alreadyCachedObject, cachedObjectExists := objectLayer[objectKey]
	if cachedObjectExists {
		alreadyCachedObject.(*CachedObjectImpl).Retain()
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl)
		return
	}

	if !createMissingCachedObject {
		return
	}

	if !writeLocked {
		objectStorage.cacheMutex.RUnlock()
		objectStorage.cacheMutex.Lock()
		writeLocked = true
	}

	if alreadyCachedObject, cachedObjectExists = objectLayer[objectKey]; cachedObjectExists {
		alreadyCachedObject.(*CachedObjectImpl).Retain()
		cacheHit = true
		cachedObject = alreadyCachedObject.(*CachedObjectImpl)
		return
	}

	// create a new cached object to hold the object
	newlyCachedObj := newCachedObject(objectStorage, key)
	newlyCachedObj.Retain()

	if objectStorage.size == 0 {
		objectStorage.cachedObjectsEmpty.Add(1)
	}

	objectLayer[objectKey] = newlyCachedObj
	objectStorage.size++
	return newlyCachedObj, false
}

func (objectStorage *ObjectStorage) deleteElementFromCache(key []byte) bool {
	if objectStorage.options.keyPartitions == nil {
		return objectStorage.deleteElementFromUnpartitionedCache(key)
	}
	return objectStorage.deleteElementFromPartitionedCache(key)
}

func (objectStorage *ObjectStorage) deleteElementFromUnpartitionedCache(key []byte) bool {
	_cachedObject, cachedObjectExists := objectStorage.cachedObjects[typeutils.BytesToString(key)]
	if cachedObjectExists {
		delete(objectStorage.cachedObjects, typeutils.BytesToString(key))

		objectStorage.size--

		cachedObject := _cachedObject.(*CachedObjectImpl)
		storableObject := cachedObject.Get()
		if storableObject != nil && !storableObject.IsDeleted() {
			objectStorage.Events.ObjectEvicted.Trigger(key, cachedObject.Get())
		}
	}

	return cachedObjectExists
}

func (objectStorage *ObjectStorage) deleteElementFromPartitionedCache(key []byte) (elementExisted bool) {
	keyPartitions := objectStorage.options.keyPartitions
	partitionCount := len(keyPartitions)
	keyOffset := 0

	// holds a stack of partitions
	partitionStack := []map[string]interface{}{objectStorage.cachedObjects}

	// stack up partitions as long as we aren't at the object layer
	var partitionId int
	var currentPartition map[string]interface{}
	for ; partitionId < partitionCount-1; partitionId++ {
		partitionKeyLength := keyPartitions[partitionId]
		currentPartition = partitionStack[len(partitionStack)-1]

		partitionKey := typeutils.BytesToString(key[keyOffset : keyOffset+partitionKeyLength])
		keyOffset += partitionKeyLength

		subPartition, subMapExists := currentPartition[partitionKey]
		if !subMapExists {
			// partition doesn't exist, so we can't delete anything
			return
		}
		partitionStack = append(partitionStack, subPartition.(map[string]interface{}))
	}

	// forward current partition to object layer
	objectLayer := partitionStack[len(partitionStack)-1]

	// grab the object key
	objectKey := typeutils.BytesToString(key[keyOffset:])
	if _, elementExisted = objectLayer[objectKey]; !elementExisted {
		return
	}

	// delete the object from the partition
	delete(objectLayer, objectKey)
	objectStorage.size--

	// if there are other elements in this partition or this object storage contains
	// only two partitions, we can stop the deletion op.
	if len(objectLayer) != 0 || len(partitionStack) <= 1 {
		return
	}

	// if the current partition is empty and we have parent partitions, we start to delete
	// partitions bottom-up, if the corresponding partitions became empty
	parentPartitionId := partitionId
	for ; len(currentPartition) == 0 && len(partitionStack) >= 2; parentPartitionId-- {

		// get parent partition
		parentPartition := partitionStack[len(partitionStack)-2]
		parentKeyLength := objectStorage.options.keyPartitions[parentPartitionId]

		// since the previous partition became empty, we delete the key leading to it from the parent partition
		delete(parentPartition, typeutils.BytesToString(key[keyOffset-parentKeyLength:keyOffset]))
		keyOffset -= parentKeyLength

		// now recursively go up a partition and check whether it became empty by the deletion
		currentPartition = parentPartition
		partitionStack = partitionStack[:len(partitionStack)-1]
	}
	return
}

func (objectStorage *ObjectStorage) putObjectInCache(object StorableObject) *CachedObjectImpl {
	cachedObject, _ := objectStorage.accessCache(object.GetStorageKey(), true)
	if !cachedObject.publishResult(object) {
		cachedObject.updateResult(object)
	}

	return cachedObject
}

func (objectStorage *ObjectStorage) loadObjectFromBadger(key []byte) StorableObject {
	if !objectStorage.options.persistenceEnabled {
		return nil
	}
	var marshaledData []byte
	if err := objectStorage.options.badgerInstance.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(objectStorage.generatePrefix([][]byte{key})); err != nil {
			return err
		} else {
			return item.Value(func(val []byte) error {
				marshaledData = make([]byte, len(val))
				copy(marshaledData, val)

				return nil
			})
		}
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return nil
		} else {
			panic(err)
		}
	} else {
		return objectStorage.unmarshalObject(key, marshaledData)
	}
}

func (objectStorage *ObjectStorage) objectExistsInBadger(key []byte) bool {
	if !objectStorage.options.persistenceEnabled {
		return false
	}

	if err := objectStorage.options.badgerInstance.View(func(txn *badger.Txn) (err error) {
		_, err = txn.Get(append(objectStorage.storageId, key...))

		return
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return false
		} else {
			panic(err)
		}
	} else {
		return true
	}
}

func (objectStorage *ObjectStorage) unmarshalObject(key []byte, data []byte) StorableObject {
	object := objectStorage.objectFactory(key)
	if err := object.UnmarshalBinary(data); err != nil {
		panic(err)
	} else {
		return object
	}
}

func (objectStorage *ObjectStorage) generatePrefix(optionalPrefixes [][]byte) (prefix []byte) {
	prefix = objectStorage.storageId
	for _, optionalPrefix := range optionalPrefixes {
		prefix = append(prefix, optionalPrefix...)
	}

	return
}

func (objectStorage *ObjectStorage) flush() {
	// create a list of objects that shall be flushed (so the BatchWriter can access the cachedObjects mutex and delete)
	objectStorage.cacheMutex.RLock()
	cachedObjects := make([]*CachedObjectImpl, objectStorage.size)
	var i int
	objectStorage.deepIterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObjectImpl) bool {
		cachedObject.cancelScheduledRelease()

		cachedObjects[i] = cachedObject
		i++

		return true
	})
	objectStorage.cacheMutex.RUnlock()

	// manually push the objects to the BatchWriter
	for _, cachedObject := range cachedObjects {
		objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
	}

	objectStorage.cachedObjectsEmpty.Wait()
}

// iterates over all cached objects and calls the consumer function on them.
func (objectStorage *ObjectStorage) deepIterateThroughCachedElements(sourceMap map[string]interface{}, consumer ConsumerFunc) bool {
	for _, value := range sourceMap {
		if cachedObject, cachedObjectReached := value.(*CachedObjectImpl); cachedObjectReached {
			if !consumer(cachedObject.key, cachedObject) {
				return false
			}
			continue
		}
		if !objectStorage.deepIterateThroughCachedElements(value.(map[string]interface{}), consumer) {
			return false
		}
	}

	return true
}

// calls the consumer function on every object within the cache and returns a set of seen keys.
func (objectStorage *ObjectStorage) forEachCachedElement(consumer ConsumerFunc) map[string]types.Empty {
	objectStorage.cacheMutex.RLock()
	defer objectStorage.cacheMutex.RUnlock()

	seenElements := make(map[string]types.Empty)
	if !objectStorage.deepIterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObjectImpl) bool {
		seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

		if !cachedObject.Exists() {
			return true
		}

		cachedObject.Retain()
		return consumer(key, cachedObject)
	}) {
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

	objectStorage.cacheMutex.RLock()
	defer objectStorage.cacheMutex.RUnlock()

	for i, partitionKeyLength := range keyPartitions {
		// if the keyOffset equals the prefixLength, we're at the wanted layer
		if keyOffset == prefixLength {
			break
		}

		if keyOffset+partitionKeyLength > prefixLength {
			panic("the prefix length does not align with the set KeyPartitions")
		}

		partitionKey := typeutils.BytesToString(prefix[keyOffset : keyOffset+partitionKeyLength])
		keyOffset += partitionKeyLength

		// advance partitions as long as we don't hit the object layer
		if i != partitionCount-1 {
			subPartition, subPartitionExists := currentPartition[partitionKey]
			if !subPartitionExists {
				// no partition exists for the given prefix
				return nil
			}
			currentPartition = subPartition.(map[string]interface{})
			continue
		}

		if keyOffset < prefixLength {
			panic("the prefix is too long for the set KeyPartition")
		}

		cachedObject := currentPartition[partitionKey].(*CachedObjectImpl)
		seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

		// the given prefix references a partition
		if !cachedObject.Exists() {
			continue
		}

		cachedObject.Retain()
		if !consumer(cachedObject.key, cachedObject) {
			return nil
		}

		return seenElements
	}

	// deep-iterate over the current partition and call the consumer function on each object
	if !objectStorage.deepIterateThroughCachedElements(currentPartition, func(key []byte, cachedObject *CachedObjectImpl) bool {
		seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

		if !cachedObject.Exists() {
			return true
		}

		cachedObject.Retain()
		return consumer(key, cachedObject)
	}) {
		return nil
	}

	if keyOffset > prefixLength {
		panic("the prefix length does not align with the set KeyPartitions")
	}

	return seenElements
}
