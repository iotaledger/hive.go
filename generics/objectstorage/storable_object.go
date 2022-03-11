package objectstorage

import "github.com/iotaledger/hive.go/objectstorage"

type StorableObject interface {
	FromObjectStorage(key, data []byte) (StorableObject, error)

	objectstorage.StorableObject
}

type StorableObjectFactory func(key []byte, data []byte) (result StorableObject, err error)
