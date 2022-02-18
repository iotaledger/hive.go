package objectstorage

import "github.com/iotaledger/hive.go/objectstorage"

type StorableObject interface {
	FromBytes(bytes []byte) (StorableObject, error)
	FromObjectStorage(key, data []byte) (StorableObject, error)

	objectstorage.StorableObject
}
