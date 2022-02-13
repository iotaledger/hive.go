package objectstorage

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedObject[T StorableObject] struct {
	cachedObject objectstorage.CachedObject
}

func newCachedObject[T StorableObject](cachedObject objectstorage.CachedObject) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: cachedObject,
	}
}

func NewEmptyCachedObject[T StorableObject](key []byte) (result *CachedObject[T]) {
	return &CachedObject[T]{
		cachedObject: objectstorage.NewEmptyCachedObject(key),
	}
}

func (c *CachedObject[T]) Key() []byte {
	return c.cachedObject.Key()
}

func (c *CachedObject[T]) Exists() bool {
	return c.Exists()
}

func (c *CachedObject[T]) Get() (result T) {
	return c.cachedObject.Get().(T)
}

func (c *CachedObject[T]) Consume(consumer func(T), forceRelease ...bool) bool {
	return c.cachedObject.Consume(func(object objectstorage.StorableObject) {
		consumer(object.(T))
	})
}

func (c *CachedObject[T]) Retain() *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.Retain(),
	}
}

func (c *CachedObject[T]) Release(force ...bool) {
	c.cachedObject.Release(force...)
}

func (c *CachedObject[T]) Transaction(callback func(object T), identifiers ...interface{}) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.Transaction(func(object objectstorage.StorableObject) {
			callback(object.(T))
		}),
	}
}

func (c *CachedObject[T]) RTransaction(callback func(object T), identifiers ...interface{}) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: c.cachedObject.RTransaction(func(object objectstorage.StorableObject) {
			callback(object.(T))
		}),
	}
}

func (c *CachedObject[T]) BatchWrite(batchedMuts kvstore.BatchedMutations) {
	c.cachedObject.BatchWrite(batchedMuts)
}

func (c *CachedObject[T]) BatchWriteDone() {
	c.cachedObject.BatchWriteDone()
}

func (c *CachedObject[T]) BatchWriteScheduled() bool {
	return c.cachedObject.BatchWriteScheduled()
}

func (c *CachedObject[T]) ResetBatchWriteScheduled() {
	c.cachedObject.ResetBatchWriteScheduled()
}
