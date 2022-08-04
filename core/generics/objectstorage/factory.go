package objectstorage

import "github.com/iotaledger/hive.go/core/kvstore"

// NewStoreWithRealm creates a new kvstore.KVStore with the given the store and prefixes.
func NewStoreWithRealm(store kvstore.KVStore, packagePrefix byte, storagePrefix byte) kvstore.KVStore {
	storeWithRealm, err := store.WithRealm([]byte{packagePrefix, storagePrefix})
	if err != nil {
		panic(err)
	}
	return storeWithRealm
}
