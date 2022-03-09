package objectstorage

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/timedexecutor"
)

type ObjectStorage[T StorableObject] struct {
	Events *Events

	*objectstorage.ObjectStorage
}

func New[T StorableObject](store kvstore.KVStore, optionalOptions ...Option) (newObjectStorage *ObjectStorage[T]) {
	newObjectStorage = &ObjectStorage[T]{
		Events: &Events{
			ObjectEvicted: events.NewEvent(evictionEvent[T]),
		},

		ObjectStorage: objectstorage.New(store, objectFactory[T], optionalOptions...),
	}

	newObjectStorage.ObjectStorage.Events.ObjectEvicted.Attach(events.NewClosure(func(key []byte, object objectstorage.StorableObject) {
		newObjectStorage.Events.ObjectEvicted.Trigger(key, object.(T))
	}))

	return newObjectStorage
}

func (o *ObjectStorage[T]) Put(object T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Put(object))
}

func (o *ObjectStorage[T]) Store(object T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Store(object))
}

func (o *ObjectStorage[T]) GetSize() int {
	return o.ObjectStorage.GetSize()
}

func (o *ObjectStorage[T]) Get(key []byte) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Get(key))
}

func (o *ObjectStorage[T]) Load(key []byte) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.Load(key))
}

func (o *ObjectStorage[T]) Contains(key []byte, options ...ReadOption) (result bool) {
	return o.ObjectStorage.Contains(key, options...)
}

func (o *ObjectStorage[T]) ComputeIfAbsent(key []byte, remappingFunction func(key []byte) T) *CachedObject[T] {
	return newCachedObject[T](o.ObjectStorage.ComputeIfAbsent(key, func(key []byte) objectstorage.StorableObject {
		return remappingFunction(key)
	}))
}

func (o *ObjectStorage[T]) DeleteIfPresent(key []byte) bool {
	return o.ObjectStorage.DeleteIfPresent(key)
}

func (o *ObjectStorage[T]) DeleteIfPresentAndReturn(key []byte) T {
	return o.ObjectStorage.DeleteIfPresentAndReturn(key).(T)
}

func (o *ObjectStorage[T]) Delete(key []byte) {
	o.ObjectStorage.Delete(key)
}

func (o *ObjectStorage[T]) StoreIfAbsent(object T) (result *CachedObject[T], stored bool) {
	untypedObject, stored := o.ObjectStorage.StoreIfAbsent(object)
	if stored {
		return newCachedObject[T](untypedObject), stored
	}
	return nil, stored
}

func (o *ObjectStorage[T]) ForEach(consumer func(key []byte, cachedObject *CachedObject[T]) bool, options ...IteratorOption) {
	o.ObjectStorage.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		return consumer(key, newCachedObject[T](cachedObject))
	}, options...)
}

func (o *ObjectStorage[T]) ForEachKeyOnly(consumer func(key []byte) bool, options ...IteratorOption) {
	o.ObjectStorage.ForEachKeyOnly(func(key []byte) bool {
		return consumer(key)
	}, options...)
}

func (o *ObjectStorage[T]) Prune() error {
	return o.ObjectStorage.Prune()
}

func (o *ObjectStorage[T]) Flush() {
	o.ObjectStorage.Flush()
}

func (o *ObjectStorage[T]) Shutdown() {
	o.ObjectStorage.Shutdown()
}

func (o *ObjectStorage[T]) FreeMemory() {
	o.ObjectStorage.FreeMemory()
}

func (o *ObjectStorage[T]) ReleaseExecutor() (releaseExecutor *timedexecutor.TimedExecutor) {
	return o.ObjectStorage.ReleaseExecutor()
}

func objectFactory[T StorableObject](key, data []byte) (result objectstorage.StorableObject, err error) {
	var obj T

	return obj.FromObjectStorage(key, data)
}
