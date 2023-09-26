package generic

import "github.com/izuc/zipp.foundation/objectstorage"

// StorableObject is an interface to be implemented by an object that is stored in the ObjectStorage.
type StorableObject interface {
	FromObjectStorage(key, data []byte) error

	objectstorage.StorableObject
}

// PtrStorableObject is a wrapper type that holds a pointer to a StorableObject.
type PtrStorableObject[T any] interface {
	*T

	StorableObject
}

// StorableObjectFactory is a function that creates new StorageObject from a key and data.
type StorableObjectFactory func(key []byte, data []byte) (result StorableObject, err error)
