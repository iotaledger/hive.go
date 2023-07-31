package ds

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// AuthenticatedMap is a sparse merkle tree based map.
type AuthenticatedMap[K, V any] interface {
	Root() types.Identifier
	Set(key K, value V) error
	Get(key K) (value V, exists bool, err error)
	Has(key K) (exists bool, err error)
	Delete(key K) (deleted bool, err error)
	Stream(consumerFunc func(key K, value V) error) error
	Commit() error
	Size() int
	WasRestoredFromStorage() bool
}

// NewAuthenticatedMap creates a new sparse merkle tree based map.
func NewAuthenticatedMap[K, V any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K], vToBytes kvstore.ObjectToBytes[V], bytesToV kvstore.BytesToObject[V]) AuthenticatedMap[K, V] {
	return newAuthenticatedMap(store, kToBytes, bytesToK, vToBytes, bytesToV)
}
