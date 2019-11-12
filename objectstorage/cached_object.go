package objectstorage

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type CachedObject struct {
	key           []byte
	objectStorage *ObjectStorage
	value         StorableObject
	err           error
	consumers     int32
	published     int32
	store         int32
	stored        int32
	delete        int32
	wg            sync.WaitGroup
	valueMutex    sync.RWMutex
	releaseTimer  unsafe.Pointer
}

func newCachedObject(database *ObjectStorage, key []byte) (result *CachedObject) {
	result = &CachedObject{
		objectStorage: database,
		key:           key,
	}

	result.wg.Add(1)

	return
}

// Retrieves the StorableObject, that is cached in this container.
func (cachedObject *CachedObject) Get() (result StorableObject) {
	if !cachedObject.IsDeleted() {
		cachedObject.valueMutex.RLock()
		result = cachedObject.value
		cachedObject.valueMutex.RUnlock()
	}

	return
}

// Releases the object, to be picked up by the persistence layer (as soon as all consumers are done).
func (cachedObject *CachedObject) Release() {
	if consumers := atomic.AddInt32(&(cachedObject.consumers), -1); consumers == 0 {
		if cachedObject.objectStorage.options.cacheTime != 0 {
			atomic.StorePointer(&cachedObject.releaseTimer, unsafe.Pointer(time.AfterFunc(cachedObject.objectStorage.options.cacheTime, func() {
				atomic.StorePointer(&cachedObject.releaseTimer, nil)

				if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
					batchWrite(cachedObject)
				} else if consumers < 0 {
					panic("called Release() too often")
				}
			})))
		} else {
			batchWrite(cachedObject)
		}
	}
}

// Directly consumes the StorableObject. This method automatically Release()s the object when the callback is done.
func (cachedObject *CachedObject) Consume(consumer func(object StorableObject)) {
	if cachedObject.Exists() && !cachedObject.IsDeleted() {
		consumer(cachedObject.Get())
	}

	cachedObject.Release()
}

// Marks an object for deletion in the persistence layer.
func (cachedObject *CachedObject) Delete() *CachedObject {
	atomic.StoreInt32(&(cachedObject.store), 0)
	atomic.StoreInt32(&(cachedObject.stored), 0)
	atomic.StoreInt32(&(cachedObject.delete), 1)

	return cachedObject
}

// Returns true if this object is supposed to be deleted from the in the persistence layer (Delete() was called).
func (cachedObject *CachedObject) IsDeleted() bool {
	return atomic.LoadInt32(&(cachedObject.delete)) == 1
}

// Marks an object for being stored in the persistence layer.
func (cachedObject *CachedObject) Store() {
	atomic.StoreInt32(&(cachedObject.delete), 0)
	atomic.StoreInt32(&(cachedObject.stored), 0)
	atomic.StoreInt32(&(cachedObject.store), 1)
}

// Returns true if the object is either persisted already or is supposed to be persisted (Store() was called).
func (cachedObject *CachedObject) IsStored() bool {
	return atomic.LoadInt32(&(cachedObject.stored)) == 1 || atomic.LoadInt32(&(cachedObject.store)) == 1
}

// Registers a new consumer for this cached object.
func (cachedObject *CachedObject) RegisterConsumer() {
	atomic.AddInt32(&(cachedObject.consumers), 1)

	if timer := atomic.SwapPointer(&cachedObject.releaseTimer, nil); timer != nil {
		(*(*time.Timer)(timer)).Stop()
	}
}

func (cachedObject *CachedObject) Exists() bool {
	return cachedObject.Get() != nil
}

func (cachedObject *CachedObject) updateValue(value StorableObject) {
	cachedObject.valueMutex.Lock()
	cachedObject.value = value
	cachedObject.valueMutex.Unlock()
}

func (cachedObject *CachedObject) publishResult(result StorableObject, err error) bool {
	if atomic.AddInt32(&(cachedObject.published), 1) == 1 {
		cachedObject.value = result
		cachedObject.err = err
		cachedObject.wg.Done()

		return true
	}

	return false
}

func (cachedObject *CachedObject) waitForResult() (*CachedObject, error) {
	if atomic.LoadInt32(&(cachedObject.published)) != 1 {
		cachedObject.wg.Wait()
	}

	if err := cachedObject.err; err != nil {
		return nil, err
	} else {
		return cachedObject, nil
	}
}
