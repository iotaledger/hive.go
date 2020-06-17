package badger

import (
	"fmt"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgerStore_Clear(t *testing.T) {
	const itemCount = 5

	store := newBadgerStore(t)
	require.EqualValues(t, 0, countKeys(t, store))

	for i := 0; i < itemCount; i++ {
		err := store.Set([]byte(fmt.Sprint(i)), []byte("a"))
		require.NoError(t, err)
	}
	assert.EqualValues(t, itemCount, countKeys(t, store))

	// check that Clear removes all the keys
	err := store.Clear()
	assert.NoError(t, err)
	assert.EqualValues(t, 0, countKeys(t, store))
}

func newBadgerStore(t *testing.T) kvstore.KVStore {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	require.NoError(t, err)
	return New(db)
}

func countKeys(t *testing.T, store kvstore.KVStore) int {
	count := 0
	err := store.IterateKeys(kvstore.EmptyPrefix, func(k kvstore.Key) bool {
		count++
		return true
	})
	require.NoError(t, err)

	return count
}
