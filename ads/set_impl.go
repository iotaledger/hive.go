package ads

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Set is a sparse merkle tree based set.
type authenticatedSet[IdentifierType types.IdentifierType, K any] struct {
	*authenticatedMap[IdentifierType, K, types.Empty]
}

// NewAuthenticatedSet creates a new sparse merkle tree based set.
func newAuthenticatedSet[IdentifierType types.IdentifierType, K any](
	store kvstore.KVStore,
	identifierToBytes kvstore.ObjectToBytes[IdentifierType],
	bytesToIdentifier kvstore.BytesToObject[IdentifierType],
	keyToBytes kvstore.ObjectToBytes[K],
	bytesToKey kvstore.BytesToObject[K],
) Set[IdentifierType, K] {
	return &authenticatedSet[IdentifierType, K]{
		authenticatedMap: newAuthenticatedMap[IdentifierType](store, identifierToBytes, bytesToIdentifier, keyToBytes, bytesToKey, types.Empty.Bytes, types.EmptyFromBytes),
	}
}

// Add adds the key to the set.
func (s *authenticatedSet[IdentifierType, K]) Add(key K) error {
	return s.Set(key, types.Void)
}

// Stream iterates over the set and calls the callback for each element.
func (s *authenticatedSet[IdentifierType, K]) Stream(callback func(key K) error) error {
	return s.authenticatedMap.Stream(func(key K, _ types.Empty) error {
		return callback(key)
	})
}
