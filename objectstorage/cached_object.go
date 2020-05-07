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

type CachedObject interface {
	Exists() bool
	Get() (result StorableObject)
	Consume(consumer func(StorableObject), forceRelease ...bool) bool
	Retain() CachedObject
	retain() CachedObject
	Release(force ...bool)
}

type CachedObjectImpl struct {
	key                 []byte
	objectStorage       *ObjectStorage
	value               StorableObject
	consumers           int32
	published           int32
	evicted             int32
	batchWriteScheduled int32
	wg                  sync.WaitGroup
	valueMutex          syncutils.RWMutex
	releaseTimer        unsafe.Pointer
	blindDelete         typeutils.AtomicBool
}

func newCachedObject(database *ObjectStorage, key []byte) (result *CachedObjectImpl) {
	result = &CachedObjectImpl{
		objectStorage: database,
		key:           key,
	}

	result.wg.Add(1)

	return
}

// Creates an "empty" CachedObjectImpl, that is not part of any ObjectStorage.
//
// Sometimes, we want to be able to offer a "filtered view" on the ObjectStorage and therefore be able to return an
// "empty" value on load operations even if the underlying object exists (i.e. the value tangle on top of the normal
// tangle only returns value transactions in its load operations).
func NewEmptyCachedObject(key []byte) (result *CachedObjectImpl) {
	result = &CachedObjectImpl{
		key:       key,
		published: 1,
		consumers: math.MinInt32,
	}

	return
}

// Retrieves the StorableObject, that is cached in this container.
func (cachedObject *CachedObjectImpl) Get() (result StorableObject) {
	cachedObject.valueMutex.RLock()
	result = cachedObject.value
	cachedObject.valueMutex.RUnlock()

	return
}

// Releases the object, to be picked up by the persistence layer (as soon as all consumers are done).
func (cachedObject *CachedObjectImpl) Release(force ...bool) {
	var forceRelease bool
	if len(force) >= 1 {
		forceRelease = force[0]
	}

	if consumers := atomic.AddInt32(&(cachedObject.consumers), -1); consumers == 0 {
		if !forceRelease && cachedObject.objectStorage.options.cacheTime != 0 {
			atomic.StorePointer(&cachedObject.releaseTimer, unsafe.Pointer(time.AfterFunc(cachedObject.objectStorage.options.cacheTime, func() {
				atomic.StorePointer(&cachedObject.releaseTimer, nil)

				if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
					cachedObject.objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
				} else if consumers < 0 {
					panic("called Release() too often")
				}
			})))
		} else {
			// only force release if there is no timer running, so that objects that landed in the cache through normal
			// loading stay available
			if atomic.LoadPointer(&cachedObject.releaseTimer) == nil {
				cachedObject.objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
			}
		}
	} else if consumers < 0 {
		panic("called Release() too often")
	}
}

// Directly consumes the StorableObject. This method automatically Release()s the object when the callback is done.
// Returns true if the callback was called.
func (cachedObject *CachedObjectImpl) Consume(consumer func(StorableObject), forceRelease ...bool) bool {
	defer cachedObject.Release(forceRelease...)

	if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) && !storableObject.IsDeleted() {
		consumer(storableObject)

		return true
	}

	return false
}

// Registers a new consumer for this cached object.
func (cachedObject *CachedObjectImpl) Retain() CachedObject {
	if atomic.AddInt32(&(cachedObject.consumers), 1) == 1 {
		panic("called Retain() on an already released CachedObject")
	}

	cachedObject.cancelScheduledRelease()

	return cachedObject
}

// Registers a new consumer for this cached object.
func (cachedObject *CachedObjectImpl) retain() CachedObject {
	atomic.AddInt32(&(cachedObject.consumers), 1)

	cachedObject.cancelScheduledRelease()

	return cachedObject
}

func (cachedObject *CachedObjectImpl) Exists() bool {
	storableObject := cachedObject.Get()

	return !typeutils.IsInterfaceNil(storableObject) && !storableObject.IsDeleted()
}

func (cachedObject *CachedObjectImpl) publishResult(result StorableObject) bool {
	if atomic.AddInt32(&(cachedObject.published), 1) == 1 {
		cachedObject.value = result
		cachedObject.wg.Done()

		return true
	}

	return false
}

func (cachedObject *CachedObjectImpl) updateResult(object StorableObject) {
	cachedObject.valueMutex.Lock()
	if typeutils.IsInterfaceNil(cachedObject.value) {
		cachedObject.value = object
		cachedObject.blindDelete.UnSet()
	} else {
		cachedObject.value.SetModified(object.IsModified())
		cachedObject.value.Persist(object.PersistenceEnabled())
		cachedObject.value.Delete(object.IsDeleted())
		cachedObject.value.Update(object)
		cachedObject.blindDelete.UnSet()
	}
	cachedObject.valueMutex.Unlock()
}

func (cachedObject *CachedObjectImpl) updateEmptyResult(update interface{}) (updated bool) {
	cachedObject.valueMutex.RLock()
	if typeutils.IsInterfaceNil(cachedObject.value) || cachedObject.value.IsDeleted() {
		cachedObject.valueMutex.RUnlock()

		cachedObject.valueMutex.Lock()
		if typeutils.IsInterfaceNil(cachedObject.value) || cachedObject.value.IsDeleted() {
			if object, ok := update.(StorableObject); ok {
				cachedObject.value = object
				cachedObject.blindDelete.UnSet()
			} else if updater, ok := update.(func() StorableObject); ok {
				cachedObject.value = updater()
				cachedObject.blindDelete.UnSet()
			}

			updated = true
		}
		cachedObject.valueMutex.Unlock()
	} else {
		cachedObject.valueMutex.RUnlock()
	}

	return
}

func (cachedObject *CachedObjectImpl) waitForInitialResult() *CachedObjectImpl {
	cachedObject.wg.Wait()

	return cachedObject
}

func (cachedObject *CachedObjectImpl) cancelScheduledRelease() {
	if timer := atomic.SwapPointer(&cachedObject.releaseTimer, nil); timer != nil {
		(*(*time.Timer)(timer)).Stop()
	}
}
