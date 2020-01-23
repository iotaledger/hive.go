package objectstorage

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
)

type CachedObject struct {
	key           []byte
	objectStorage *ObjectStorage
	value         StorableObject
	consumers     int32
	published     int32
	wg            sync.WaitGroup
	valueMutex    syncutils.RWMutex
	releaseTimer  unsafe.Pointer
	blindDelete   typeutils.AtomicBool
}

func newCachedObject(database *ObjectStorage, key []byte) (result *CachedObject) {
	result = &CachedObject{
		objectStorage: database,
		key:           key,
	}

	result.wg.Add(1)

	return
}

// Creates an "empty" CachedObject, that is not part of any ObjectStorage.
//
// Sometimes, we want to be able to offer a "filtered view" on the ObjectStorage and therefore be able to return an
// "empty" value on load operations even if the underlying object exists (i.e. the value tangle on top of the normal
// tangle only returns value transactions in its load operations).
func NewEmptyCachedObject(key []byte) (result *CachedObject) {
	result = &CachedObject{
		key:       key,
		published: 1,
		consumers: math.MinInt32,
	}

	return
}

// Retrieves the StorableObject, that is cached in this container.
func (cachedObject *CachedObject) Get() (result StorableObject) {
	cachedObject.valueMutex.RLock()
	result = cachedObject.value
	cachedObject.valueMutex.RUnlock()

	return
}

// Releases the object, to be picked up by the persistence layer (as soon as all consumers are done).
func (cachedObject *CachedObject) Release() {
	if consumers := atomic.AddInt32(&(cachedObject.consumers), -1); consumers == 0 {
		if cachedObject.objectStorage.options.cacheTime != 0 {
			atomic.StorePointer(&cachedObject.releaseTimer, unsafe.Pointer(time.AfterFunc(cachedObject.objectStorage.options.cacheTime, func() {
				atomic.StorePointer(&cachedObject.releaseTimer, nil)

				if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
					cachedObject.objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
				} else if consumers < 0 {
					panic("called Release() too often")
				}
			})))
		} else {
			cachedObject.objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
		}
	}
}

// Directly consumes the StorableObject. This method automatically Release()s the object when the callback is done.
func (cachedObject *CachedObject) Consume(consumer func(StorableObject)) {
	if storableObject := cachedObject.Get(); storableObject != nil && !storableObject.IsDeleted() {
		consumer(storableObject)
	}

	cachedObject.Release()
}

// Registers a new consumer for this cached object.
func (cachedObject *CachedObject) RegisterConsumer() {
	atomic.AddInt32(&(cachedObject.consumers), 1)

	cachedObject.cancelScheduledRelease()
}

func (cachedObject *CachedObject) Exists() bool {
	storableObject := cachedObject.Get()

	return storableObject != nil && !storableObject.IsDeleted()
}

func (cachedObject *CachedObject) publishResult(result StorableObject) bool {
	if atomic.AddInt32(&(cachedObject.published), 1) == 1 {
		cachedObject.value = result
		cachedObject.wg.Done()

		return true
	}

	return false
}

func (cachedObject *CachedObject) updateResult(object StorableObject) {
	cachedObject.valueMutex.Lock()
	if cachedObject.value == nil {
		cachedObject.value = object
	} else {
		cachedObject.value.Update(object)
	}
	cachedObject.valueMutex.Unlock()
}

func (cachedObject *CachedObject) updateEmptyResult(update interface{}) (updated bool) {
	cachedObject.valueMutex.RLock()
	if cachedObject.value == nil {
		cachedObject.valueMutex.RUnlock()

		cachedObject.valueMutex.Lock()
		if cachedObject.value == nil {
			if object, ok := update.(StorableObject); ok {
				cachedObject.value = object
			} else if updater, ok := update.(func() StorableObject); ok {
				cachedObject.value = updater()
			}

			updated = true
		}
		cachedObject.valueMutex.Unlock()
	} else {
		cachedObject.valueMutex.RUnlock()
	}

	return
}

func (cachedObject *CachedObject) waitForInitialResult() *CachedObject {
	cachedObject.wg.Wait()

	return cachedObject
}

func (cachedObject *CachedObject) cancelScheduledRelease() {
	if timer := atomic.SwapPointer(&cachedObject.releaseTimer, nil); timer != nil {
		(*(*time.Timer)(timer)).Stop()
	}
}
