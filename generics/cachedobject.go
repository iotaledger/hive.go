package generics

import "github.com/iotaledger/hive.go/objectstorage"

type CachedObject[T objectstorage.StorableObject] struct {
	cachedObject objectstorage.CachedObject
}

func NewCachedObject[T objectstorage.StorableObject](cachedObject objectstorage.CachedObject) *CachedObject[T] {
	return &CachedObject[T]{
		cachedObject: cachedObject,
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
