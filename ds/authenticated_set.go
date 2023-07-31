package ds

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// AuthenticatedSet is a sparse merkle tree based set.
type AuthenticatedSet[K any] interface {
	Root() types.Identifier
	Add(key K) error
	Has(key K) (exists bool, err error)
	Delete(key K) (deleted bool, err error)
	Stream(consumerFunc func(key K) error) error
	Commit() error
	Size() int
	WasRestoredFromStorage() bool
}

// NewAuthenticatedSet creates a new sparse merkle tree based map.
func NewAuthenticatedSet[K any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K]) AuthenticatedSet[K] {
	return newAuthenticatedSet(store, kToBytes, bytesToK)
}
