package ads

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Set is a sparse merkle tree based set.
type Set[K any] interface {
	Root() types.Identifier
	Add(key K) error
	Has(key K) (exists bool, err error)
	Delete(key K) (deleted bool, err error)
	Stream(consumerFunc func(key K) error) error
	Commit() error
	Size() int
	WasRestoredFromStorage() bool
}

// NewSet creates a new sparse merkle tree based map.
func NewSet[K any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K]) Set[K] {
	return newAuthenticatedSet(store, kToBytes, bytesToK)
}