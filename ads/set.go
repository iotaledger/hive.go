package ads

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Set is a set that can produce proofs for its elements which can be verified against a known merkle root
// that is formed using a sparse merkle tree.
type Set[K any] interface {
	// Root returns the root of the sparse merkle tree.
	Root() types.Identifier

	// Add adds the key to the set.
	Add(key K) error

	// Has returns true if the given key exists.
	Has(key K) (exists bool, err error)

	// Delete deletes the given key.
	Delete(key K) (deleted bool, err error)

	// Stream streams all the set elements to the given consumer function.
	Stream(consumerFunc func(key K) error) error

	// Commit persists the changes to the underlying store.
	Commit() error

	// Size returns the number of elements in the set.
	Size() int

	// WasRestoredFromStorage returns true if the set was restored from an existing storage.
	WasRestoredFromStorage() bool
}

// NewSet creates a new sparse merkle tree based set.
func NewSet[K any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K]) Set[K] {
	return newAuthenticatedSet(store, kToBytes, bytesToK)
}
