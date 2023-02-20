package flushkv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
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

func TestFlushKVStore_Get(t *testing.T) {
	store := New(mapdb.NewMapDB())
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

func TestFlushKVStore_Iterate(t *testing.T) {
	store := New(mapdb.NewMapDB())
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

func TestFlushKVStore_IterateDirection(t *testing.T) {
	store := New(mapdb.NewMapDB())
	for _, entry := range testEntries {
		err := store.Set(entry.Key, entry.Value)
		require.NoError(t, err)
	}

	// forward iteration
	i := 0
	err := store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		entry := &struct {
			kvstore.Key
			kvstore.Value
		}{key, value}
		assert.Equal(t, testEntries[i], entry, "entries are not equal")
		i++

		return true
	}, kvstore.IterDirectionForward)
	assert.NoError(t, err)
	assert.Equal(t, len(testEntries), i)

	// backward iteration
	i = 0
	err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		entry := &struct {
			kvstore.Key
			kvstore.Value
		}{key, value}
		assert.Equal(t, testEntries[len(testEntries)-1-i], entry, "entries are not equal")
		i++

		return true
	}, kvstore.IterDirectionBackward)
	assert.NoError(t, err)
	assert.Equal(t, len(testEntries), i)
}

func TestFlushKVStore_Realm(t *testing.T) {
	store := New(mapdb.NewMapDB())
	realm := kvstore.Realm("realm")
	realmStore, err := store.WithRealm(realm)
	require.NoError(t, err)

	key := []byte("key")
	err = realmStore.Set(key, []byte("value"))
	require.NoError(t, err)

	tmpStore, err := store.WithRealm(kvstore.Realm("tmp"))
	require.NoError(t, err)

	key2 := []byte("key2")
	err = tmpStore.Set(key2, []byte("value"))
	require.NoError(t, err)

	realmStore2, err := store.WithRealm(realm)
	require.NoError(t, err)

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

func TestFlushKVStore_Clear(t *testing.T) {
	store := New(mapdb.NewMapDB())
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
