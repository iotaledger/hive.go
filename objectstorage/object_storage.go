package objectstorage

import (
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
)

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
}

func New(storageId []byte, objectFactory StorableObjectFactory, optionalOptions ...ObjectStorageOption) *ObjectStorage {
	return &ObjectStorage{
		storageId:     storageId,
		objectFactory: objectFactory,
		cachedObjects: make(map[string]interface{}),
		options:       newObjectStorageOptions(optionalOptions),
	}
}

func (objectStorage *ObjectStorage) Put(object StorableObject) *CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	return objectStorage.putObjectInCache(object)
}

func (objectStorage *ObjectStorage) Store(object StorableObject) *CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if !objectStorage.options.persistenceEnabled {
		panic("persistence is disabled - use Put(object StorableObject) instead of Store(object StorableObject)")
	}

	object.Persist()
	object.SetModified()

	return objectStorage.putObjectInCache(object)
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

func (objectStorage *ObjectStorage) Get(key []byte) *CachedObject {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	cachedObject, cacheHit := objectStorage.accessCache(key, true)
	if !cacheHit {
		cachedObject.publishResult(nil)
	}

	return cachedObject.waitForInitialResult()
}

func (objectStorage *ObjectStorage) Load(key []byte) *CachedObject {
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

	return cachedObject.waitForInitialResult()
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

func (objectStorage *ObjectStorage) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) StorableObject) *CachedObject {
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

	return cachedObject.waitForInitialResult()
}

// This method deletes an element and return true if the element was deleted.
func (objectStorage *ObjectStorage) DeleteIfPresent(key []byte) bool {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	deleteExistingEntry := func(cachedObject *CachedObject) bool {
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

	deleteExistingEntry := func(cachedObject *CachedObject) {
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
func (objectStorage *ObjectStorage) StoreIfAbsent(key []byte, object StorableObject) (cachedObject *CachedObject, stored bool) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

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
		cachedObject.waitForInitialResult()
	}

	return
}

// Foreach can only iterate over persisted entries, so there might be a slight delay before you can find previously
// stored items in such an iteration.
func (objectStorage *ObjectStorage) ForEach(consumer func(key []byte, cachedObject *CachedObject) bool, optionalPrefix ...[]byte) {
	if objectStorage.shutdown.IsSet() {
		panic("trying to access shutdown object storage")
	}

	if objectStorage.options.keyPartitions == nil && len(optionalPrefix) >= 1 {
		panic("prefix iterations are only allowed when the option PartitionKey(....) is set")
	}

	var seenElements map[string]types.Empty
	if len(optionalPrefix) == 0 || len(optionalPrefix[0]) == 0 {
		if seenElements = objectStorage.forEachCachedElement(consumer); seenElements == nil {
			return
		}
	} else {
		if seenElements = objectStorage.forEachCachedElementWithPrefix(consumer, optionalPrefix[0]); seenElements == nil {
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

			if _, elementSeen := seenElements[typeutils.BytesToString(key)]; !elementSeen {
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

				if cachedObject.Exists() && !consumer(key, cachedObject) {
					break
				}
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

func (objectStorage *ObjectStorage) accessCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObject, cacheHit bool) {
	objectStorage.flushMutex.RLock()

	copiedKey := make([]byte, len(key))
	copy(copiedKey, key)

	keyPartitions := objectStorage.options.keyPartitions
	if keyPartitions == nil {
		keyPartitions = []int{len(key)}
	}
	keyPartitionCount := len(keyPartitions)

	currentMap := objectStorage.cachedObjects
	keyOffset := 0

	objectStorage.cacheMutex.RLock()
	var writeLocked bool
	for i, keyPartitionLength := range keyPartitions {
		partitionStringKey := typeutils.BytesToString(copiedKey[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		if i == keyPartitionCount-1 {
			alreadyCachedObject, cachedObjectExists := currentMap[partitionStringKey]
			if cachedObjectExists {
				alreadyCachedObject.(*CachedObject).RegisterConsumer()

				cacheHit = true
			} else {
				if !createMissingCachedObject {
					objectStorage.cacheMutex.RUnlock()

					return
				}

				if !writeLocked {
					objectStorage.cacheMutex.RUnlock()
					objectStorage.cacheMutex.Lock()
					writeLocked = true
				}

				if alreadyCachedObject, cachedObjectExists = currentMap[partitionStringKey]; cachedObjectExists {
					alreadyCachedObject.(*CachedObject).RegisterConsumer()

					cacheHit = true
				} else {
					alreadyCachedObject = newCachedObject(objectStorage, copiedKey)
					alreadyCachedObject.(*CachedObject).RegisterConsumer()

					if objectStorage.size == 0 {
						objectStorage.cachedObjectsEmpty.Add(1)
					}

					currentMap[partitionStringKey] = alreadyCachedObject
					objectStorage.size++
				}
			}

			cachedObject = alreadyCachedObject.(*CachedObject)
		} else {
			subMap, subMapExists := currentMap[partitionStringKey]
			if subMapExists {
				currentMap = subMap.(map[string]interface{})
			} else {
				if !createMissingCachedObject {
					objectStorage.cacheMutex.RUnlock()

					return
				}

				if !writeLocked {
					objectStorage.cacheMutex.RUnlock()
					objectStorage.cacheMutex.Lock()
					writeLocked = true
				}

				subMap, subMapExists = currentMap[partitionStringKey]
				if subMapExists {
					currentMap = subMap.(map[string]interface{})
				} else {
					subMap = make(map[string]interface{})

					currentMap[partitionStringKey] = subMap

					currentMap = subMap.(map[string]interface{})
				}
			}
		}
	}

	if writeLocked {
		objectStorage.cacheMutex.Unlock()
	} else {
		objectStorage.cacheMutex.RUnlock()
	}

	return
}

func (objectStorage *ObjectStorage) deleteElementFromCache(key []byte) bool {
	if objectStorage.options.keyPartitions == nil {
		return objectStorage.deleteElementFromUnpartitionedCache(key)
	} else {
		return objectStorage.deleteElementFromPartitionedCache(key)
	}
}

func (objectStorage *ObjectStorage) deleteElementFromUnpartitionedCache(key []byte) bool {
	_, cachedObjectExists := objectStorage.cachedObjects[typeutils.BytesToString(key)]
	if cachedObjectExists {
		delete(objectStorage.cachedObjects, typeutils.BytesToString(key))

		objectStorage.size--
	}

	return cachedObjectExists
}

func (objectStorage *ObjectStorage) deleteElementFromPartitionedCache(key []byte) (elementExisted bool) {
	keyPartitionCount := len(objectStorage.options.keyPartitions)
	keyOffset := 0
	mapStack := make([]map[string]interface{}, 1)
	mapStack[0] = objectStorage.cachedObjects

	for keyPartitionId, keyPartitionLength := range objectStorage.options.keyPartitions {
		currentMap := mapStack[len(mapStack)-1]

		partitionStringKey := typeutils.BytesToString(key[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		if keyPartitionId == keyPartitionCount-1 {
			lastPartitionStringKey := typeutils.BytesToString(key[keyOffset-keyPartitionLength:])

			_, elementExisted = currentMap[lastPartitionStringKey]
			if elementExisted {
				delete(currentMap, lastPartitionStringKey)

				if len(currentMap) == 0 && len(mapStack) > 1 {
					parentKeyPartitionId := keyPartitionId
					keyOffset -= keyPartitionLength
					for len(currentMap) == 0 && len(mapStack) > 1 {
						parentMap := mapStack[len(mapStack)-2]
						parentKeyPartitionId = parentKeyPartitionId - 1
						parentKeyLength := objectStorage.options.keyPartitions[parentKeyPartitionId]

						delete(parentMap, typeutils.BytesToString(key[keyOffset-parentKeyLength:keyOffset]))
						keyOffset -= parentKeyLength

						currentMap = parentMap
						mapStack = mapStack[:len(mapStack)-1]
					}
				}

				objectStorage.size--
			}
		} else {
			if subMap, subMapExists := currentMap[partitionStringKey]; subMapExists {
				mapStack = append(mapStack, subMap.(map[string]interface{}))
			} else {
				return
			}
		}
	}

	return
}

func (objectStorage *ObjectStorage) putObjectInCache(object StorableObject) *CachedObject {
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
	cachedObjects := make([]*CachedObject, objectStorage.size)
	var i int
	objectStorage.iterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObject) bool {
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

func (objectStorage *ObjectStorage) iterateThroughCachedElements(sourceMap map[string]interface{}, consumer func(key []byte, cachedObject *CachedObject) bool) bool {
	for _, value := range sourceMap {
		if cachedObject, cachedObjectReached := value.(*CachedObject); cachedObjectReached {
			if !consumer(cachedObject.key, cachedObject) {
				return false
			}
		} else {
			if !objectStorage.iterateThroughCachedElements(value.(map[string]interface{}), consumer) {
				return false
			}
		}
	}

	return true
}

func (objectStorage *ObjectStorage) forEachCachedElement(consumer func(key []byte, cachedObject *CachedObject) bool) map[string]types.Empty {
	seenElements := make(map[string]types.Empty)
	objectStorage.cacheMutex.RLock()
	if !objectStorage.iterateThroughCachedElements(objectStorage.cachedObjects, func(key []byte, cachedObject *CachedObject) bool {
		seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

		if !cachedObject.Exists() {
			return true
		}

		cachedObject.RegisterConsumer()
		return consumer(key, cachedObject)
	}) {
		objectStorage.cacheMutex.RUnlock()

		return nil
	}
	objectStorage.cacheMutex.RUnlock()

	return seenElements
}

func (objectStorage *ObjectStorage) forEachCachedElementWithPrefix(consumer func(key []byte, cachedObject *CachedObject) bool, prefix []byte) map[string]types.Empty {
	seenElements := make(map[string]types.Empty)

	optionalPrefixLength := len(prefix)
	keyPartitions := objectStorage.options.keyPartitions
	keyPartitionCount := len(keyPartitions)
	currentMap := objectStorage.cachedObjects
	keyOffset := 0

	objectStorage.cacheMutex.RLock()
	for i, keyPartitionLength := range keyPartitions {
		if keyOffset == optionalPrefixLength {
			if !objectStorage.iterateThroughCachedElements(currentMap, func(key []byte, cachedObject *CachedObject) bool {
				seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

				if !cachedObject.Exists() {
					return true
				}

				cachedObject.RegisterConsumer()
				return consumer(key, cachedObject)
			}) {
				objectStorage.cacheMutex.RUnlock()

				return nil
			}

			break
		}

		if keyOffset+keyPartitionLength > optionalPrefixLength {
			objectStorage.cacheMutex.RUnlock()

			panic("the prefix length does not align with the set KeyPartition")
		}

		partitionStringKey := typeutils.BytesToString(prefix[keyOffset : keyOffset+keyPartitionLength])
		keyOffset += keyPartitionLength

		if i == keyPartitionCount-1 {
			if keyOffset < optionalPrefixLength {
				objectStorage.cacheMutex.RUnlock()

				panic("the prefix is too long for the set KeyPartition")
			}

			cachedObject := currentMap[partitionStringKey].(*CachedObject)
			seenElements[typeutils.BytesToString(cachedObject.key)] = types.Void

			if cachedObject.Exists() {
				cachedObject.RegisterConsumer()
				if !consumer(cachedObject.key, cachedObject) {
					objectStorage.cacheMutex.RUnlock()

					return nil
				}
			}
		} else {
			if subMap, subMapExists := currentMap[partitionStringKey]; subMapExists {
				currentMap = subMap.(map[string]interface{})
			} else {
				break
			}
		}
	}
	objectStorage.cacheMutex.RUnlock()

	if keyOffset > optionalPrefixLength {
		panic("the prefix length does not align with the set KeyPartition")
	}

	return seenElements
}
