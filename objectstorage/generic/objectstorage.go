package generic

import (
	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/objectstorage/typeutils"
	"github.com/izuc/zipp.foundation/runtime/event"
	"github.com/izuc/zipp.foundation/runtime/timed"
)

// ObjectStorage is a manual cache which keeps objects as long as consumers are using it.
type ObjectStorage[T StorableObject] struct {
	Events *Events

	*objectstorage.ObjectStorage
}

// NewStructStorage is the constructor for the ObjectStorage that stores struct types.
func NewStructStorage[U any, T PtrStorableObject[U]](store kvstore.KVStore, optionalOptions ...Option) (newObjectStorage *ObjectStorage[T]) {
	newObjectStorage = &ObjectStorage[T]{
		Events: &Events{
			ObjectEvicted: event.New2[[]byte, objectstorage.StorableObject](),
		},

		ObjectStorage: objectstorage.New(store, objectFactory[T, U], optionalOptions...),
	}

	newObjectStorage.ObjectStorage.Events.ObjectEvicted.Hook(func(key []byte, object objectstorage.StorableObject) {
		newObjectStorage.Events.ObjectEvicted.Trigger(key, object.(T))
	})

	return newObjectStorage
}

// NewInterfaceStorage is the constructor for the ObjectStorage that stores interface types.
func NewInterfaceStorage[T StorableObject](store kvstore.KVStore, objectFactory StorableObjectFactory, optionalOptions ...Option) (newObjectStorage *ObjectStorage[T]) {
	newObjectStorage = &ObjectStorage[T]{
		Events: &Events{
			ObjectEvicted: event.New2[[]byte, objectstorage.StorableObject](),
		},

		ObjectStorage: objectstorage.New(store, func(key, data []byte) (objectstorage.StorableObject, error) { return objectFactory(key, data) }, optionalOptions...),
	}

	newObjectStorage.ObjectStorage.Events.ObjectEvicted.Hook(func(key []byte, object objectstorage.StorableObject) {
		newObjectStorage.Events.ObjectEvicted.Trigger(key, object.(T))
	})

	return newObjectStorage
}

// Put adds the given object in the ObjectStorage cache.
func (o *ObjectStorage[T]) Put(object T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Put(object))
}

// Store stores the given object in the ObjectStorage.
func (o *ObjectStorage[T]) Store(object T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Store(object))
}

// GetSize returns the size of the ObjectStorage.
func (o *ObjectStorage[T]) GetSize() int {
	return o.ObjectStorage.GetSize()
}

// Get returns the object for the given key from cache. If object is not in cache, returns nil.
func (o *ObjectStorage[T]) Get(key []byte) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Get(key))
}

// Load returns the object for the given key. If object is not found in cache, tries to load the object from underlying store. Can only be used with persistence enabled.
func (o *ObjectStorage[T]) Load(key []byte) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Load(key))
}

// Contains returns true if the given key is in the ObjectStorage.
func (o *ObjectStorage[T]) Contains(key []byte, options ...ReadOption) (result bool) {
	return o.ObjectStorage.Contains(key, options...)
}

// ComputeIfAbsent computes and returns the default value if the given key is not in the ObjectStorage.
func (o *ObjectStorage[T]) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.ComputeIfAbsent(key, func(key []byte) objectstorage.StorableObject {
		remappedObj := remappingFunction(key)
		if !typeutils.IsInterfaceNil(remappedObj) {
			remappedObj.SetModified(true)
		}
		return remappedObj
	}))
}

// DeleteIfPresent deletes an element and return true if the element was deleted.
func (o *ObjectStorage[T]) DeleteIfPresent(key []byte) bool {
	return o.ObjectStorage.DeleteIfPresent(key)
}

// DeleteIfPresentAndReturn deletes an element and returns it. If the element does not exist then the return value is
// nil.
func (o *ObjectStorage[T]) DeleteIfPresentAndReturn(key []byte) T {
	return o.ObjectStorage.DeleteIfPresentAndReturn(key).(T)
}

// Delete performs a "blind delete", where we do not check the object's existence.
// blindDelete is used to delete without accessing the value log.
func (o *ObjectStorage[T]) Delete(key []byte) {
	o.ObjectStorage.Delete(key)
}

// StoreIfAbsent stores an object only if it was not stored before. In contrast to "ComputeIfAbsent", this method does not access the
// value log. If the object was not stored, then the returned CachedObject is nil and does not need to be Released.
func (o *ObjectStorage[T]) StoreIfAbsent(object T) (result *CachedObject[T], stored bool) {
	untypedObject, stored := o.ObjectStorage.StoreIfAbsent(object)
	if stored {
		return newCachedObject[T](untypedObject), stored
	}

	return nil, stored
}

// ForEach calls the consumer function on every object residing within the cache and the underlying persistence layer.
func (o *ObjectStorage[T]) ForEach(consumer func(key []byte, cachedObject *CachedObject[T]) bool, options ...IteratorOption) {
	o.ObjectStorage.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		return consumer(key, newCachedObject[T](cachedObject))
	}, options...)
}

// ForEachKeyOnly calls the consumer function on every storage key residing within the cache and the underlying persistence layer.
func (o *ObjectStorage[T]) ForEachKeyOnly(consumer func(key []byte) bool, options ...IteratorOption) {
	o.ObjectStorage.ForEachKeyOnly(func(key []byte) bool {
		return consumer(key)
	}, options...)
}

// Prune removes all values from the ObjectStorage.
func (o *ObjectStorage[T]) Prune() error {
	return o.ObjectStorage.Prune()
}

// Flush writes all objects from cache to the underlying store.
func (o *ObjectStorage[T]) Flush() {
	o.ObjectStorage.Flush()
}

// Shutdown shuts down the ObjectStorage.
func (o *ObjectStorage[T]) Shutdown() {
	o.ObjectStorage.Shutdown()
}

// FreeMemory copies the content of the internal maps to newly created maps.
// This is necessary, otherwise the GC is not able to free the memory used by the old maps.
// "delete" doesn't shrink the maximum memory used by the map, since it only marks the entry as deleted.
func (o *ObjectStorage[T]) FreeMemory() {
	o.ObjectStorage.FreeMemory()
}

// ReleaseExecutor returns the executor that schedules releases of CachedObjects after the configured CacheTime.
func (o *ObjectStorage[T]) ReleaseExecutor() (releaseExecutor *timed.Executor) {
	return o.ObjectStorage.ReleaseExecutor()
}

func objectFactory[T PtrStorableObject[U], U any](key, data []byte) (result objectstorage.StorableObject, err error) {
	var instance U
	typedResult := T(&instance)
	if err = typedResult.FromObjectStorage(key, data); err != nil {
		return nil, err
	}

	return typedResult, nil
}
