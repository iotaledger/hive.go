package badger

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testEntries = []*struct {
	kvstore.Key
	kvstore.Value
}{
	{Key: []byte("a"), Value: []byte("valueA")},
	{Key: []byte("b"), Value: []byte("valueB")},
	{Key: []byte("c"), Value: []byte("valueC")},
	{Key: []byte("d"), Value: []byte("valueD")},
}

func TestBadgerStore_Get(t *testing.T) {
	store := newBadgerStore(t)
	for _, entry := range testEntries {
		err := store.Set(entry.Key, entry.Value)
		require.NoError(t, err)
	}

	for _, entry := range testEntries {
		value, err := store.Get(entry.Key)
		assert.Equal(t, entry.Value, value)
		assert.NoError(t, err)
	}

	value, err := store.Get([]byte("invalid"))
	assert.Nil(t, value)
	assert.Equal(t, kvstore.ErrKeyNotFound, err)
}

func TestBadgerStore_Iterate(t *testing.T) {
	store := newBadgerStore(t)
	for _, entry := range testEntries {
		err := store.Set(entry.Key, entry.Value)
		require.NoError(t, err)
	}

	i := 0
	err := store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		entry := &struct {
			kvstore.Key
			kvstore.Value
		}{key, value}
		assert.Contains(t, testEntries, entry)
		i++
		return true
	})
	assert.NoError(t, err)
	assert.Equal(t, len(testEntries), i)
}

func TestBadgerStore_Realm(t *testing.T) {
	store := newBadgerStore(t)
	realm := kvstore.Realm("realm")
	realmStore := store.WithRealm(realm)

	key := []byte("key")
	err := realmStore.Set(key, []byte("value"))
	require.NoError(t, err)

	tmpStore := store.WithRealm(kvstore.Realm("tmp"))
	key2 := []byte("key2")
	err = tmpStore.Set(key2, []byte("value"))
	require.NoError(t, err)

	realmStore2 := store.WithRealm(realm)

	has, err := realmStore2.Has(key)
	assert.NoError(t, err)
	assert.True(t, has)
	has, err = realmStore2.Has(key2)
	assert.NoError(t, err)
	assert.False(t, has)

	// when clearing "realm" the key in "tmp" should still be there
	assert.NoError(t, realmStore.Clear())
	has, err = tmpStore.Has(key2)
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestBadgerStore_Clear(t *testing.T) {
	store := newBadgerStore(t)
	require.EqualValues(t, 0, countKeys(t, store))

	for _, entry := range testEntries {
		err := store.Set(entry.Key, entry.Value)
		require.NoError(t, err)
	}
	assert.Equal(t, len(testEntries), countKeys(t, store))

	// check that Clear removes all the keys
	err := store.Clear()
	assert.NoError(t, err)
	assert.EqualValues(t, 0, countKeys(t, store))
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

func newBadgerStore(t *testing.T) kvstore.KVStore {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	require.NoError(t, err)
	return New(db)
}
