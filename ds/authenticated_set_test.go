package ds_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func TestSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := ads.NewSet(store,
		testKey.Bytes,
		testKeyFromBytes,
	)

	key := testKey([]byte{'a'})
	newSet.Add(key)
	exist := newSet.Has(key)
	require.True(t, exist)

	// add the same key again
	newSet.Add(key)
	exist = newSet.Has(key)
	require.True(t, exist)
	require.Equal(t, 1, newSet.Size())
	root := newSet.Root()

	// Test deleting a key
	require.True(t, newSet.Delete(key))
	exist = newSet.Has(key)
	require.False(t, exist)

	// Test deleting a non-existent key
	require.False(t, newSet.Delete(key))
	require.Equal(t, 0, newSet.Size())

	// make sure the root has changed
	root1 := newSet.Root()
	require.NotEqualValues(t, root, root1)

	// new set from old store, make sure the root is correct
	newSet1 := ads.NewSet(store,
		testKey.Bytes,
		testKeyFromBytes,
	)
	require.EqualValues(t, newSet.Root(), newSet1.Root())
}

func TestStreamSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := ads.NewSet(store,
		testKey.Bytes,
		testKeyFromBytes,
	)

	key1 := testKey([]byte{'b'})
	key2 := testKey([]byte{'c'})
	newSet.Add(key1)
	newSet.Add(key2)
	require.Equal(t, 2, newSet.Size())

	seen := make(map[testKey]bool)
	err := newSet.Stream(func(key testKey) bool {
		seen[key] = true
		return true
	})
	require.NoError(t, err)
	require.True(t, seen[key1])
	require.True(t, seen[key2])
	require.Equal(t, 2, len(seen))

	// with consume function returning false, only 1 element will be visited.
	firstSeen := make(map[testKey]bool)
	err = newSet.Stream(func(key testKey) bool {
		firstSeen[key] = true
		return false
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(firstSeen))
}
