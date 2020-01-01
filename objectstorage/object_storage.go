package objectstorage

import (
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"

	"github.com/dgraph-io/badger/v2"
)

type ObjectStorage struct {
	storageId      []byte
	objectFactory  StorableObjectFactory
	cachedObjects  map[string]*CachedObject
	cacheMutex     syncutils.RWMutex
	options        *ObjectStorageOptions
}

func New(storageId []byte, objectFactory StorableObjectFactory, optionalOptions ...ObjectStorageOption) *ObjectStorage {
	return &ObjectStorage{
		storageId:      storageId,
		objectFactory:  objectFactory,
		cachedObjects:  map[string]*CachedObject{},
		options:        newObjectStorageOptions(optionalOptions),
	}
}

func (objectStorage *ObjectStorage) Prepare(object StorableObject) *CachedObject {
	return objectStorage.putObjectInCache(object)
}

func (objectStorage *ObjectStorage) Store(object StorableObject) *CachedObject {
	object.Persist()
	object.SetModified()

	return objectStorage.putObjectInCache(object)
}

func (objectStorage *ObjectStorage) GetSize() int {
	objectStorage.cacheMutex.RLock()
	size := len(objectStorage.cachedObjects)
	objectStorage.cacheMutex.RUnlock()
	return size
}

func (objectStorage *ObjectStorage) Load(key []byte) *CachedObject {
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

func (objectStorage *ObjectStorage) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) StorableObject) *CachedObject {
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

// Stores an object only if it was not stored before. In contrast to "ComputeIfAbsent", this method does not access the
// value log. If the object was not stored, then the returned CachedObject is nil and does not need to be Released.
func (objectStorage *ObjectStorage) StoreIfAbsent(key []byte, object StorableObject) (cachedObject *CachedObject, stored bool) {
	existingCachedObject, cacheHit := objectStorage.accessCache(key, false)
	if cacheHit {
		existingCachedObject.wg.Wait()

		object.Persist()
		object.SetModified()

		if stored = existingCachedObject.updateEmptyResult(object); stored {
			cachedObject = existingCachedObject
		}
	} else {
		if objectExists := objectStorage.objectExistsInBadger(key); !objectExists {
			object.Persist()
			object.SetModified()

			if newCachedObject, cacheHit := objectStorage.accessCache(key, true); cacheHit {
				newCachedObject.wg.Wait()

				if stored = newCachedObject.updateEmptyResult(object); stored {
					cachedObject = newCachedObject
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

func (objectStorage *ObjectStorage) Delete(key []byte) {
	cachedObject := objectStorage.Load(key)
	if existingObject := cachedObject.Get(); existingObject != nil {
		existingObject.Delete()
	}
	cachedObject.Release()
}

// Foreach can only iterate over persisted entries, so there might be a slight delay before you can find previously
// stored items in such an iteration.
func (objectStorage *ObjectStorage) ForEach(consumer func(key []byte, cachedObject *CachedObject) bool, optionalPrefixes ...[]byte) error {
	return objectStorage.options.badgerInstance.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = objectStorage.generatePrefix(optionalPrefixes)

		it := txn.NewIterator(iteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()[len(objectStorage.storageId):]

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

			if !consumer(key, cachedObject) {
				break
			}
		}
		it.Close()

		return nil
	})
}

func (objectStorage *ObjectStorage) Prune() error {
	objectStorage.cacheMutex.Lock()
	if err := objectStorage.options.badgerInstance.DropPrefix(objectStorage.storageId); err != nil {
		return err
	}
	objectStorage.cachedObjects = map[string]*CachedObject{}
	objectStorage.cacheMutex.Unlock()

	return nil
}

func (objectStorage *ObjectStorage) WaitForWritesToFlush() {
	objectStorage.options.batchedWriterInstance.WaitForWritesToFlush()
}

func (objectStorage *ObjectStorage) StopBatchWriter() {
	objectStorage.options.batchedWriterInstance.StopBatchWriter()
}

func (objectStorage *ObjectStorage) accessCache(key []byte, createMissingCachedObject bool) (cachedObject *CachedObject, cacheHit bool) {
	copiedKey := make([]byte, len(key))
	copy(copiedKey, key)
	stringKey := typeutils.BytesToString(copiedKey)

	objectStorage.cacheMutex.RLock()
	alreadyCachedObject, cachedObjectExists := objectStorage.cachedObjects[stringKey]
	if cachedObjectExists {
		alreadyCachedObject.RegisterConsumer()

		objectStorage.cacheMutex.RUnlock()

		cacheHit = true
	} else {
		objectStorage.cacheMutex.RUnlock()
		objectStorage.cacheMutex.Lock()
		if alreadyCachedObject, cachedObjectExists = objectStorage.cachedObjects[stringKey]; cachedObjectExists {
			alreadyCachedObject.RegisterConsumer()

			objectStorage.cacheMutex.Unlock()

			cacheHit = true
		} else {
			if createMissingCachedObject {
				alreadyCachedObject = newCachedObject(objectStorage, copiedKey)
				alreadyCachedObject.RegisterConsumer()

				objectStorage.cachedObjects[stringKey] = alreadyCachedObject
			}
			objectStorage.cacheMutex.Unlock()
		}
	}

	cachedObject = alreadyCachedObject

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
