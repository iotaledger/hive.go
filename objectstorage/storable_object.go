package objectstorage

import (
	"encoding"
)

type StorableObject interface {
	// import the default Marshaler interfaces
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

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
	PersistenceEnabled() bool

	// Updates the object with the values of another object "in place" (so it should use a pointer receiver)
	Update(other StorableObject)

	// Returns the key that identifies the object in the object storage.
	GetStorageKey() []byte
}
