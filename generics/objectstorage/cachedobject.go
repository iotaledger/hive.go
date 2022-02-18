package objectstorage

import (
	"strconv"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/stringify"
)

// region CachedObject //////////////////////////////////////////////////////////////////////////////////////////

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
	return c.cachedObject.Exists()
}

func (c *CachedObject[T]) Get() (result T) {
	return c.cachedObject.Get().(T)
}

func (c *CachedObject[T]) Unwrap() (result T, exists bool) {
	if !c.Exists() {
		return
	}
	r := c.Get()
	result = r
	exists = true
	return
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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region CachedObjects //////////////////////////////////////////////////////////////////////////////////////////

// CachedObjects represents a collection of CachedObject objects.
type CachedObjects[T StorableObject] []*CachedObject[T]

// Unwrap is the type-casted equivalent of Get. It returns a slice of unwrapped objects and optionally skips any objects
// that do not exist or are deleted, sets default type value for missing elements.
func (c CachedObjects[T]) Unwrap(skip ...bool) (unwrappedChildBranches []T) {
	skipMissing := false
	if len(skip) > 0 && skip[0] == true {
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

// String returns a human readable version of the CachedObjects.
func (c CachedObjects[T]) String() string {
	structBuilder := stringify.StructBuilder("CachedObjects")
	for i, cachedObject := range c {
		structBuilder.AddField(stringify.StructField(strconv.Itoa(i), cachedObject))
	}

	return structBuilder.String()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
