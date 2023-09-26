package objectstorage

type StorableObject interface {
	// Marks the object as modified, which causes it to be written to the disk (if persistence is enabled).
	// Returns the former state of the boolean.
	SetModified(modified ...bool) (wasSet bool)

	// Returns true if the object was marked as modified.
	IsModified() bool

	// Marks the object to be deleted from the persistence layer.
	// Returns the former state of the boolean.
	//nolint:predeclared // lets keep this for now
	Delete(delete ...bool) (wasSet bool)

	// Returns true if the object was marked as deleted.
	IsDeleted() bool

	// Enables or disables persistence for this object. Objects that have persistence disabled get discarded once they
	// are evicted from the cache.
	// Returns the former state of the boolean.
	Persist(enabled ...bool) (wasSet bool)

	// Returns "true" if this object is going to be persisted.
	ShouldPersist() bool

	// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
	ObjectStorageKey() []byte

	// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
	ObjectStorageValue() []byte
}
