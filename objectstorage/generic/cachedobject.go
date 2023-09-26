package generic

import (
	"strconv"

	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/stringify"
)

// region CachedObject //////////////////////////////////////////////////////////////////////////////////////////

// CachedObject is a wrapper around a value that is stored in the object storage.
// It provides necessary function that object storage needs to correctly handle the object.
type CachedObject[T StorableObject] struct {
	cachedObject objectstorage.CachedObject
}

func newCachedObject[T StorableObject](cachedObject objectstorage.CachedObject) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: cachedObject,
	}
}

// NewEmptyCachedObject creates an "empty" CachedObject, that is not part of any ObjectStorage.
//
// Sometimes, we want to be able to offer a "filtered view" on the ObjectStorage and therefore be able to return an
// "empty" value on load operations even if the underlying object exists (i.e. the value tangle on top of the normal
// tangle only returns value transactions in its load operations).
func NewEmptyCachedObject[T StorableObject](key []byte) (result *CachedObject[T]) {
	return &CachedObject[T]{
		cachedObject: objectstorage.NewEmptyCachedObject(key),
	}
}

// Key returns the object storage key that is used to address the object.
func (c *CachedObject[T]) Key() []byte {
	return c.cachedObject.Key()
}

// Exists returns true if the StorableObject in this container does exist (could be found in the database and was not
// marked as deleted).
func (c *CachedObject[T]) Exists() bool {
	return c.cachedObject.Exists()
}

// Get retrieves the StorableObject, that is cached in this container.
func (c *CachedObject[T]) Get() (result T) {
	return c.cachedObject.Get().(T)
}

// Unwrap returns the underlying object with correct type.
func (c *CachedObject[T]) Unwrap() (result T, exists bool) {
	if !c.Exists() {
		return
	}
	r := c.Get()
	result = r
	exists = true

	return
}

// Consume directly consumes the StorableObject. This method automatically Release()s the object when the callback is done.
// Returns true if the callback was called.
func (c *CachedObject[T]) Consume(consumer func(T), forceRelease ...bool) bool {
	return c.cachedObject.Consume(func(object objectstorage.StorableObject) {
		consumer(object.(T))
	})
}

// Retain registers a new consumer for this cached object.
func (c *CachedObject[T]) Retain() *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.Retain(),
	}
}

// Release the object, to be picked up by the persistence layer (as soon as all consumers are done).
func (c *CachedObject[T]) Release(force ...bool) {
	c.cachedObject.Release(force...)
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
func (c *CachedObject[T]) Transaction(callback func(object T), identifiers ...interface{}) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.Transaction(func(object objectstorage.StorableObject) {
			callback(object.(T))
		}),
	}
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
func (c *CachedObject[T]) RTransaction(callback func(object T), identifiers ...interface{}) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.RTransaction(func(object objectstorage.StorableObject) {
			callback(object.(T))
		}),
	}
}

// BatchWrite checks if the cachedObject should be persisted.
// If all checks pass, the cachedObject is marshaled and added to the BatchedMutations.
// Do not call this method for objects that should not be persisted.
func (c *CachedObject[T]) BatchWrite(batchedMuts kvstore.BatchedMutations) {
	c.cachedObject.BatchWrite(batchedMuts)
}

// BatchWriteDone is called after the cachedObject was persisted.
// It releases the cachedObject from the cache if no consumers are left and it was not modified in the meantime.
func (c *CachedObject[T]) BatchWriteDone() {
	c.cachedObject.BatchWriteDone()
}

// BatchWriteScheduled returns true if the cachedObject is already scheduled for a BatchWrite operation.
func (c *CachedObject[T]) BatchWriteScheduled() bool {
	return c.cachedObject.BatchWriteScheduled()
}

// ResetBatchWriteScheduled resets the flag that the cachedObject is scheduled for a BatchWrite operation.
func (c *CachedObject[T]) ResetBatchWriteScheduled() {
	c.cachedObject.ResetBatchWriteScheduled()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region CachedObjects //////////////////////////////////////////////////////////////////////////////////////////

// CachedObjects represents a collection of CachedObject objects.
type CachedObjects[T StorableObject] []*CachedObject[T]

// Unwrap is the type-casted equivalent of Get. It returns a slice of unwrapped objects and optionally skips any objects
// that do not exist or are deleted, sets default type value for missing elements.
func (c CachedObjects[T]) Unwrap(skip ...bool) (unwrappedChildBranches []T) {
	skipMissing := false
	if len(skip) > 0 && skip[0] {
		skipMissing = true
	}
	unwrappedChildBranches = make([]T, 0, len(c))
	for _, cachedChildBranch := range c {
		val, exists := cachedChildBranch.Unwrap()
		if exists || !skipMissing {
			unwrappedChildBranches = append(unwrappedChildBranches, val)
		}
	}

	return
}

// Exists returns a slice of boolean values to indicate whether element at a given index exists.
func (c CachedObjects[T]) Exists() (exists []bool) {
	exists = make([]bool, len(c))
	for i, cachedChildBranch := range c {
		exists[i] = cachedChildBranch.Exists()
	}

	return
}

// Consume iterates over the CachedObjects, unwraps them and passes a type-casted version to the consumer (if the object
// is not empty - it exists). It automatically releases the object when the consumer finishes. It returns true, if at
// least one object was consumed.
func (c CachedObjects[T]) Consume(consumer func(T), forceRelease ...bool) (consumed bool) {
	for _, cachedObject := range c {
		consumed = cachedObject.Consume(consumer, forceRelease...) || consumed
	}

	return
}

// Release is a utility function that allows us to release all CachedObjects in the collection.
func (c CachedObjects[T]) Release(force ...bool) {
	for _, cachedObject := range c {
		cachedObject.Release(force...)
	}
}

// String returns a human-readable version of the CachedObjects.
func (c CachedObjects[T]) String() string {
	structBuilder := stringify.NewStructBuilder("CachedObjects")
	for i, cachedObject := range c {
		structBuilder.AddField(stringify.NewStructField(strconv.Itoa(i), cachedObject))
	}

	return structBuilder.String()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
