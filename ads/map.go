package ads

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Map is a map that can produce proofs for its values which can be verified against a known merkle root
// that is formed using a sparse merkle tree.
type Map[K, V any] interface {
	// Set sets the given key to the given value.
	Set(key K, value V) error

	// Get returns the value for the given key.
	Get(key K) (value V, exists bool, err error)

	// Has returns true if the given key exists.
	Has(key K) (exists bool, err error)

	// Delete deletes the given key.
	Delete(key K) (deleted bool, err error)

	// Stream streams all key-value pairs to the given consumer function.
	Stream(consumerFunc func(key K, value V) error) error

	// Commit commits the changes to the underlying store.
	Commit() error

	// Root returns the root of the sparse merkle tree.
	Root() types.Identifier

	// Size returns the number of elements in the map.
	Size() int

	// WasRestoredFromStorage returns true if the map was restored from an existing storage.
	WasRestoredFromStorage() bool
}

// NewMap creates a new AuthenticatedMap.
func NewMap[K, V any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K], vToBytes kvstore.ObjectToBytes[V], bytesToV kvstore.BytesToObject[V]) Map[K, V] {
	return newAuthenticatedMap(store, kToBytes, bytesToK, vToBytes, bytesToV)
}
