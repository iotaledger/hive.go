package objectstorage

type StorableObject interface {
	// Marks the object as modified, which causes it to be written to the disk (if persistence is enabled).
	// Default value when omitting the parameter: true
	SetModified(modified ...bool)

	// Returns true if the object was marked as modified.
	IsModified() bool

	// Marks the object to be deleted from the persistence layer.
	// Default value when omitting the parameter: true
	Delete(delete ...bool)

	// Returns true if the object was marked as deleted.
	IsDeleted() bool

	// Enables or disables persistence for this object. Objects that have persistence disabled get discarded once they
	// are evicted from the cache.
	// Default value when omitting the parameter: true
	Persist(enabled ...bool)

	// Returns "true" if this object is going to be persisted.
	ShouldPersist() bool

	// Updates the object with the values of another object "in place" (so it should use a pointer receiver)
	Update(other StorableObject)

	// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
	ObjectStorageKey() []byte

	// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
	ObjectStorageValue() []byte

	// UnmarshalStorageValue parses the value part of the k/v store.
	UnmarshalObjectStorageValue(valueBytes []byte) (consumedBytes int, err error)
}
