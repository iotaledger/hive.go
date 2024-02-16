package ads

import (
	"github.com/pokt-network/smt/kvstore"

	"github.com/iotaledger/hive.go/ierrors"
	hivekvstore "github.com/iotaledger/hive.go/kvstore"
)

var _ kvstore.MapStore = &mapStoreAdapter{}

// mapStoreAdapter is a wrapper around a hive KVStore that implements the MapStore interface
// from pokt-network/smt/kvstore.
type mapStoreAdapter struct {
	underlying hivekvstore.KVStore
}

func newMapStoreAdapter(store hivekvstore.KVStore) *mapStoreAdapter {
	return &mapStoreAdapter{
		underlying: store,
	}
}

// Get returns the value for a given key.
func (k *mapStoreAdapter) Get(key []byte) ([]byte, error) {
	return k.underlying.Get(key)
}

// Set sets/updates the value for a given key.
func (k *mapStoreAdapter) Set(key, value []byte) error {
	return k.underlying.Set(key, value)
}

// Delete removes a key.
func (k *mapStoreAdapter) Delete(key []byte) error {
	return k.underlying.Delete(key)
}

// Len returns the number of key-value pairs in the store.
func (k *mapStoreAdapter) Len() int {
	count := 0

	//nolint:revive // better be explicit here
	if err := k.underlying.IterateKeys(hivekvstore.EmptyPrefix, func(key []byte) bool {
		count++
		return true
	}); err != nil {
		panic(ierrors.Errorf("failed to iterate over keys: %w", err))
	}

	return count
}

// --- Debug ---

// ClearAll deletes all key-value pairs in the store.
func (k *mapStoreAdapter) ClearAll() error {
	return k.underlying.Clear()
}
