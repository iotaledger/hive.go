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
	Transaction(callback func(object StorableObject), identifiers ...interface{}) CachedObject
	RTransaction(callback func(object StorableObject), identifiers ...interface{}) CachedObject
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
	transactionMutex    syncutils.RWMultiMutex
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

// Exists returns true if the StorableObject in this container does exist (could be found in the database and was not
// marked as deleted).
func (cachedObject *CachedObjectImpl) Exists() bool {
	storableObject := cachedObject.Get()

	return !typeutils.IsInterfaceNil(storableObject) && !storableObject.IsDeleted()
}

// Transaction is a synchronization primitive that executes the callback atomically which means that if multiple
// Transactions are being started from different goroutines, then only one of them can run at the same time.
//
// The identifiers allow to define the scope of the Transaction. Transactions with different scopes can run at the same
// time and act as if they are secured by different mutexes.
//
// It is also possible to provide multiple identifiers and the callback waits until all of them can be acquired at the
// same time. In contrast to normal mutexes where acquiring multiple locks can lead to deadlocks, this method is
// deadlock safe.
//
// Note: It is the equivalent of a mutex.Lock/Unlock.
func (cachedObject *CachedObjectImpl) Transaction(callback func(object StorableObject), identifiers ...interface{}) CachedObject {
	if len(identifiers) == 0 {
		panic("Transaction requires at least one identifier for the scope")
	}

	cachedObject.transactionMutex.Lock(identifiers...)
	defer cachedObject.transactionMutex.Unlock(identifiers...)

	callback(cachedObject.Get())

	return cachedObject
}

// RTransaction is a synchronization primitive that executes the callback together with other RTransactions but never
// together with a normal Transaction.
//
// The identifiers allow to define the scope of the RTransaction. RTransactions with different scopes can run at the
// same time independently of other RTransactions and act as if they are secured by different mutexes.
//
// It is also possible to provide multiple identifiers and the callback waits until all of them can be acquired at the
// same time. In contrast to normal mutexes where acquiring multiple locks can lead to deadlocks, this method is
// deadlock safe.
//
// Note: It is the equivalent of a mutex.RLock/RUnlock.
func (cachedObject *CachedObjectImpl) RTransaction(callback func(object StorableObject), identifiers ...interface{}) CachedObject {
	if len(identifiers) == 0 {
		panic("RTransaction requires at least one identifier for the scope")
	}

	cachedObject.transactionMutex.RLock(identifiers...)
	defer cachedObject.transactionMutex.RUnlock(identifiers...)

	callback(cachedObject.Get())

	return cachedObject
}

// Registers a new consumer for this cached object.
func (cachedObject *CachedObjectImpl) retain() CachedObject {
	atomic.AddInt32(&(cachedObject.consumers), 1)

	cachedObject.cancelScheduledRelease()

	return cachedObject
}

func (cachedObject *CachedObjectImpl) storeOnCreation() {
	if cachedObject.objectStorage.options.persistenceEnabled && cachedObject.objectStorage.options.storeOnCreation && !typeutils.IsInterfaceNil(cachedObject.value) && cachedObject.value.IsModified() && cachedObject.value.ShouldPersist() {
		// store the object immediately
		cachedObject.objectStorage.options.batchedWriterInstance.batchWrite(cachedObject)
	}
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
	if typeutils.IsInterfaceNil(cachedObject.value) || cachedObject.value.IsDeleted() {
		cachedObject.value = object
		cachedObject.blindDelete.UnSet()
	} else {
		cachedObject.value.SetModified(object.IsModified())
		cachedObject.value.Persist(object.ShouldPersist())
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
