package objectstorage

import "github.com/iotaledger/hive.go/objectstorage"

type StorableObject interface {
	FromBytes(bytes []byte) (StorableObject, error)

	objectstorage.StorableObject
}
