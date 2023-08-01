package aset

import (
	"github.com/iotaledger/hive.go/ads/amap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Set is a sparse merkle tree based set.
type AuthenticatedSet[K any] struct {
	*amap.AuthenticatedMap[K, types.Empty]
}

// NewAuthenticatedSet creates a new sparse merkle tree based set.
func NewAuthenticatedSet[K any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K]) *AuthenticatedSet[K] {
	return &AuthenticatedSet[K]{
		AuthenticatedMap: amap.NewAuthenticatedMap(store, kToBytes, bytesToK, types.Empty.Bytes, types.EmptyFromBytes),
	}
}

// Add adds the key to the set.
func (s *AuthenticatedSet[K]) Add(key K) error {
	return s.Set(key, types.Void)
}

// Stream iterates over the set and calls the callback for each element.
func (s *AuthenticatedSet[K]) Stream(callback func(key K) error) error {
	return s.AuthenticatedMap.Stream(func(key K, _ types.Empty) error {
		return callback(key)
	})
}
