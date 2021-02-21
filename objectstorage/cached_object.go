package objectstorage

import (
	"math"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/timedexecutor"
	"github.com/iotaledger/hive.go/typeutils"
)

type CachedObject interface {
	Key() []byte
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
	scheduledTask       unsafe.Pointer
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

// Key returns the object storage key that is used to address the object.
func (cachedObject *CachedObjectImpl) Key() []byte {
	return cachedObject.key
}

// Retrieves the StorableObject, that is cached in this container.
func (cachedObject *CachedObjectImpl) Get() (result StorableObject) {
	cachedObject.valueMutex.RLock()
	defer cachedObject.valueMutex.RUnlock()

	return cachedObject.value
}

// Releases the object, to be picked up by the persistence layer (as soon as all consumers are done).
func (cachedObject *CachedObjectImpl) Release(force ...bool) {
	consumers := atomic.AddInt32(&(cachedObject.consumers), -1)
	if consumers > 1 {
		return
	}
	if consumers < 0 {
		panic("called Release() too often")
	}

	if cachedObject.objectStorage.options.cacheTime == 0 || (len(force) >= 1 && force[0]) {
		// only force release if there is no timer running, so that objects that landed in the cache through normal
		// loading stay available
		if atomic.LoadPointer(&cachedObject.scheduledTask) == nil {
			cachedObject.evict()
		}

		return
	}

	atomic.StorePointer(
		&cachedObject.scheduledTask,
		unsafe.Pointer(
			cachedObject.objectStorage.ReleaseExecutor().ExecuteAfter(
				cachedObject.delayedRelease,
				cachedObject.objectStorage.options.cacheTime,
			),
		),
	)
}

func (cachedObject *CachedObjectImpl) delayedRelease() {
	consumers := atomic.LoadInt32(&(cachedObject.consumers))
	if consumers > 1 {
		return
	}
	if consumers < 0 {
		panic("called Release() too often")
	}

	cachedObject.evict()
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
		cachedObject.evict()
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
	defer cachedObject.valueMutex.Unlock()

	if typeutils.IsInterfaceNil(cachedObject.value) || cachedObject.value.IsDeleted() {
		cachedObject.value = object
		cachedObject.blindDelete.UnSet()
		return
	}

	cachedObject.value.SetModified(object.IsModified())
	cachedObject.value.Persist(object.ShouldPersist())
	cachedObject.value.Delete(object.IsDeleted())
	cachedObject.value.Update(object)
	cachedObject.blindDelete.UnSet()
}

func (cachedObject *CachedObjectImpl) updateEmptyResult(update interface{}) (updated bool) {
	cachedObject.valueMutex.RLock()
	if !typeutils.IsInterfaceNil(cachedObject.value) && !cachedObject.value.IsDeleted() {
		cachedObject.valueMutex.RUnlock()
		return
	}

	cachedObject.valueMutex.RUnlock()
	cachedObject.valueMutex.Lock()
	defer cachedObject.valueMutex.Unlock()

	if !typeutils.IsInterfaceNil(cachedObject.value) && !cachedObject.value.IsDeleted() {
		return
	}

	switch typedUpdate := update.(type) {
	case StorableObject:
		cachedObject.value = typedUpdate
	case func() StorableObject:
		cachedObject.value = typedUpdate()
	default:
		panic("invalid argument in call to updateEmptyResult")
	}

	cachedObject.blindDelete.UnSet()
	updated = true

	return
}

func (cachedObject *CachedObjectImpl) waitForInitialResult() *CachedObjectImpl {
	cachedObject.wg.Wait()

	return cachedObject
}

func (cachedObject *CachedObjectImpl) cancelScheduledRelease() {
	if scheduledTask := atomic.SwapPointer(&cachedObject.scheduledTask, nil); scheduledTask != nil {
		(*(*timedexecutor.ScheduledTask)(scheduledTask)).Cancel()
	}
}

// evict either releases non-persistable objects or enqueues persistable objects into the batch writer.
func (cachedObject *CachedObjectImpl) evict() {
	atomic.StorePointer(&cachedObject.scheduledTask, nil)

	if !cachedObject.objectStorage.options.persistenceEnabled {
		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			storableObject.SetModified(false)
		}
		cachedObject.BatchWriteDone()
		return
	}

	cachedObject.objectStorage.options.batchedWriterInstance.Enqueue(cachedObject)
}

// BatchWriteObject interface methods

// BatchWrite checks if the cachedObject should be persisted.
// If all checks pass, the cachedObject is marshaled and added to the BatchedMutations.
// Do not call this method for objects that should not be persisted.
func (cachedObject *CachedObjectImpl) BatchWrite(batchedMuts kvstore.BatchedMutations) {
	consumers := atomic.LoadInt32(&(cachedObject.consumers))
	if consumers < 0 {
		panic("too many unregistered consumers of cached object")
	}

	storableObject := cachedObject.Get()

	if typeutils.IsInterfaceNil(storableObject) {
		// only blind delete if there are no consumers
		if consumers == 0 && cachedObject.blindDelete.IsSet() {
			if err := batchedMuts.Delete(cachedObject.key); err != nil {
				panic(err)
			}
		}

		return
	}

	if storableObject.IsDeleted() {
		// only delete if there are no consumers
		if consumers == 0 {
			storableObject.SetModified(false)

			if err := batchedMuts.Delete(cachedObject.key); err != nil {
				panic(err)
			}
		}

		return
	}

	// only store if there are no consumers anymore or the object should be stored on creation
	if consumers != 0 && !cachedObject.objectStorage.options.storeOnCreation {
		return
	}

	if !storableObject.IsModified() {
		return
	}

	storableObject.SetModified(false)

	if !storableObject.ShouldPersist() {
		return
	}

	var marshaledValue []byte
	if !cachedObject.objectStorage.options.keysOnly {
		marshaledValue = storableObject.ObjectStorageValue()
	}

	if err := batchedMuts.Set(cachedObject.key, marshaledValue); err != nil {
		panic(err)
	}
}

// BatchWriteDone is called after the cachedObject was persisted.
// It releases the cachedObject from the cache if no consumers are left and it was not modified in the meantime.
func (cachedObject *CachedObjectImpl) BatchWriteDone() {
	// acquire mutexes prior to cache modifications
	objectStorage := cachedObject.objectStorage
	objectStorage.flushMutex.RLock()
	defer objectStorage.flushMutex.RUnlock()
	objectStorage.cacheMutex.Lock()
	defer objectStorage.cacheMutex.Unlock()

	// abort if there are still consumers
	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers != 0 {
		return
	}

	// abort if the object was modified in the mean time
	if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) && storableObject.IsModified() {
		return
	}

	// abort if the object was evicted already
	if atomic.AddInt32(&cachedObject.evicted, 1) != 1 {
		return
	}

	// abort if the object could not be deleted from the cache
	if !objectStorage.deleteElementFromCache(cachedObject.key) {
		return
	}

	// abort if the storage is not empty
	if objectStorage.size != 0 {
		return
	}

	// mark storage as empty
	objectStorage.cachedObjectsEmpty.Done()
}

// BatchWriteScheduled returns true if the cachedObject is already scheduled for a BatchWrite operation.
func (cachedObject *CachedObjectImpl) BatchWriteScheduled() bool {
	return atomic.AddInt32(&(cachedObject.batchWriteScheduled), 1) != 1
}

// ResetBatchWriteScheduled resets the flag that the cachedObject is scheduled for a BatchWrite operation.
func (cachedObject *CachedObjectImpl) ResetBatchWriteScheduled() {
	atomic.StoreInt32(&(cachedObject.batchWriteScheduled), 0)
}
