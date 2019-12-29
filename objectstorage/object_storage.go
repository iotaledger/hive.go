package objectstorage

import (
	"github.com/dgraph-io/badger"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
)

type ObjectStorage struct {
	badgerInstance *badger.DB
	storageId      []byte
	objectFactory  StorableObjectFactory
	cachedObjects  map[string]*CachedObject
	cacheMutex     syncutils.RWMutex
	options        *ObjectStorageOptions
}

func New(storageId string, objectFactory StorableObjectFactory, optionalOptions ...ObjectStorageOption) *ObjectStorage {
	return &ObjectStorage{
		badgerInstance: GetBadgerInstance(),
		storageId:      []byte(storageId),
		objectFactory:  objectFactory,
		cachedObjects:  map[string]*CachedObject{},
		options:        newTransportOutputStorageFilters(optionalOptions),
	}
}

func (objectStorage *ObjectStorage) Prepare(object StorableObject) *CachedObject {
	return objectStorage.storeObjectInCache(object, false)
}

func (objectStorage *ObjectStorage) Store(object StorableObject) *CachedObject {
	object.Persist()
	object.SetModified()

	return objectStorage.storeObjectInCache(object, true)
}

func (objectStorage *ObjectStorage) Load(key []byte) (*CachedObject, error) {
	return objectStorage.accessCacheWithCallbacks(key, nil, func(cachedObject *CachedObject) {
		loadedObject, err := objectStorage.loadObjectFromBadger(key)
		if loadedObject != nil {
			loadedObject.Persist()
		}

		cachedObject.publishResult(loadedObject, err)
	}, true).waitForResult()
}

func (objectStorage *ObjectStorage) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) (StorableObject, error)) (*CachedObject, error) {
	return objectStorage.accessCacheWithCallbacks(key, func(cachedObject *CachedObject) {
		// wait for it to be published
		cachedObject.wg.Wait()

		// if currentValue is still nil => update result
		cachedObject.valueMutex.RLock()
		if cachedObject.value == nil {
			cachedObject.valueMutex.RUnlock()

			cachedObject.valueMutex.Lock()
			if cachedObject.value == nil {
				object, err := remappingFunction(key)

				cachedObject.value = object
				cachedObject.errMutex.Lock()
				cachedObject.err = err
				cachedObject.errMutex.Unlock()
			}
			cachedObject.valueMutex.Unlock()
		} else {
			cachedObject.valueMutex.RUnlock()
		}
	}, func(cachedObject *CachedObject) {
		loadedObject, err := objectStorage.loadObjectFromBadger(key)
		if loadedObject != nil {
			loadedObject.Persist()

			cachedObject.publishResult(loadedObject, err)
		} else {
			cachedObject.publishResult(remappingFunction(key))
		}
	}, true).waitForResult()
}

// Stores an object only if it was not stored before. In contrast to "ComputeIfAbsent", this method does not access the
// value log. If the object was not stored, then the returned CachedObject is nil and does not need to be Released.
func (objectStorage *ObjectStorage) StoreIfAbsent(key []byte, object StorableObject) (stored bool, cachedObject *CachedObject, err error) {
	objectStorage.accessCacheWithCallbacks(key, func(existingCachedObject *CachedObject) {
		existingCachedObject.wg.Wait()

		if stored = existingCachedObject.updateEmptyResult(object, nil); stored {
			cachedObject = existingCachedObject
		}
	}, func(*CachedObject) {
		if objectExists, err := objectStorage.objectExistsInBadger(key); err != nil {
			panic(err)
		} else {
			if !objectExists {
				if newCachedObject, cacheHit := objectStorage.accessCache(key, true); cacheHit {
					newCachedObject.wg.Wait()

					if stored = newCachedObject.updateEmptyResult(object, nil); stored {
						cachedObject = newCachedObject
					}
				} else {
					object.Persist()
					object.SetModified()

					newCachedObject.publishResult(object, nil)

					stored = true
					cachedObject = newCachedObject
				}
			}
		}
	}, false)

	_, err = cachedObject.waitForResult()

	return
}

func (objectStorage *ObjectStorage) Delete(key []byte) {
	objectStorage.accessCacheWithCallbacks(key, func(cachedObject *CachedObject) {
		if storableObject := cachedObject.Get(); storableObject != nil {
			storableObject.Persist()
			storableObject.Delete()
		}

		cachedObject.Release()
	}, func(cachedObject *CachedObject) {
		if storableObject := cachedObject.Get(); storableObject != nil {
			storableObject.Delete()
		}

		cachedObject.publishResult(nil, nil)
		cachedObject.Release()
	}, true)
}

// Foreach can only iterate over persisted entries, so there might be a slight delay before you can find previously
// stored items in such an iteration.
func (objectStorage *ObjectStorage) ForEach(consumer func(key []byte, cachedObject *CachedObject) bool, optionalPrefixes ...[]byte) error {
	return objectStorage.badgerInstance.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = objectStorage.generatePrefix(optionalPrefixes)

		it := txn.NewIterator(iteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()[len(objectStorage.storageId):]

			if cachedObject, err := objectStorage.accessCacheWithCallbacks(key, nil, func(cachedObject *CachedObject) {
				_ = item.Value(func(val []byte) error {
					marshaledData := make([]byte, len(val))
					copy(marshaledData, val)

					storableObject, err := objectStorage.unmarshalObject(key, marshaledData)
					if storableObject != nil {
						storableObject.Persist()
					}

					cachedObject.publishResult(storableObject, err)

					return nil
				})
			}, true).waitForResult(); err != nil {
				it.Close()

				return err
			} else {
				if !consumer(key, cachedObject) {
					break
				}
			}
		}
		it.Close()

		return nil
	})
}

func (objectStorage *ObjectStorage) Prune() error {
	objectStorage.cacheMutex.Lock()
	if err := objectStorage.badgerInstance.DropPrefix(objectStorage.storageId); err != nil {
		return err
	}
	objectStorage.cachedObjects = map[string]*CachedObject{}
	objectStorage.cacheMutex.Unlock()

	return nil
}

func (objectStorage *ObjectStorage) accessCacheWithCallbacks(key []byte, onCacheHit func(*CachedObject), onCacheMiss func(*CachedObject), createMissingCachedObject bool) *CachedObject {
	cachedObject, cacheHit := objectStorage.accessCache(key, createMissingCachedObject)
	if cacheHit {
		if onCacheHit != nil {
			onCacheHit(cachedObject)
		}
	} else {
		if onCacheMiss != nil {
			onCacheMiss(cachedObject)
		}
	}

	return cachedObject
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

func (objectStorage *ObjectStorage) storeObjectInCache(object StorableObject, persist bool) *CachedObject {
	return objectStorage.accessCacheWithCallbacks(object.GetStorageKey(), func(cachedObject *CachedObject) {
		if !cachedObject.publishResult(object, nil) {
			if currentValue := cachedObject.Get(); currentValue != nil {
				currentValue.Update(object)
			} else {
				cachedObject.updateValue(object)
			}
		}
	}, func(cachedObject *CachedObject) {
		if persist {
			object.Persist()
		}

		cachedObject.publishResult(object, nil)
	}, true)
}

func (objectStorage *ObjectStorage) loadObjectFromBadger(key []byte) (StorableObject, error) {
	var marshaledData []byte
	if err := objectStorage.badgerInstance.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(append(objectStorage.storageId, key...)); err != nil {
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
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return objectStorage.unmarshalObject(key, marshaledData)
	}
}

func (objectStorage *ObjectStorage) objectExistsInBadger(key []byte) (bool, error) {
	if err := objectStorage.badgerInstance.View(func(txn *badger.Txn) (err error) {
		_, err = txn.Get(append(objectStorage.storageId, key...))

		return
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func (objectStorage *ObjectStorage) unmarshalObject(key []byte, data []byte) (StorableObject, error) {
	object := objectStorage.objectFactory(key)
	if err := object.UnmarshalBinary(data); err != nil {
		return nil, err
	} else {
		return object, nil
	}
}

func (objectStorage *ObjectStorage) generatePrefix(optionalPrefixes [][]byte) (prefix []byte) {
	prefix = objectStorage.storageId
	for _, optionalPrefix := range optionalPrefixes {
		prefix = append(prefix, optionalPrefix...)
	}

	return
}
