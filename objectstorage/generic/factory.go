package generic

import (
	"github.com/izuc/zipp.foundation/kvstore"
)

// NewStoreWithRealm creates a new kvstore.KVStore with the given the store and prefixes.
func NewStoreWithRealm(store kvstore.KVStore, packagePrefix byte, storagePrefix byte) kvstore.KVStore {
	storeWithRealm, err := store.WithExtendedRealm([]byte{packagePrefix, storagePrefix})
	if err != nil {
		panic(err)
	}

	return storeWithRealm
}
