package ads

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/typeutils"
)

func TestSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := newAuthenticatedSet[[32]byte](
		store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
		testKey.Bytes,
		testKeyFromBytes,
	)

	key := testKey([]byte{'a'})
	require.NoError(t, newSet.Add(key))
	exist, err := newSet.Has(key)
	require.NoError(t, err)
	require.True(t, exist)

	// add the same key again
	require.NoError(t, newSet.Add(key))
	exist, err = newSet.Has(key)
	require.NoError(t, err)
	require.True(t, exist)
	require.Equal(t, 1, newSet.Size())
	root := newSet.Root()

	// Test deleting a key
	require.True(t, lo.PanicOnErr(newSet.Delete(key)))
	exist, err = newSet.Has(key)
	require.NoError(t, err)
	require.False(t, exist)

	// Test deleting a non-existent key
	require.False(t, lo.PanicOnErr(newSet.Delete(key)))
	require.Equal(t, 0, newSet.Size())

	// make sure the root has changed
	root1 := newSet.Root()
	require.NotEqualValues(t, root, root1)

	// new set from old store, make sure the root is correct
	newSet1 := newAuthenticatedSet[[32]byte](store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
		testKey.Bytes,
		testKeyFromBytes,
	)
	require.EqualValues(t, newSet.Root(), newSet1.Root())
}

func TestStreamSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := newAuthenticatedSet[[32]byte](store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
		testKey.Bytes,
		testKeyFromBytes,
	)

	key1 := testKey([]byte{'b'})
	key2 := testKey([]byte{'c'})
	require.NoError(t, newSet.Add(key1))
	require.NoError(t, newSet.Add(key2))
	require.Equal(t, 2, newSet.Size())

	seen := make(map[testKey]bool)
	err := newSet.Stream(func(key testKey) error {
		seen[key] = true

		return nil
	})
	require.NoError(t, err)
	require.True(t, seen[key1])
	require.True(t, seen[key2])
	require.Equal(t, 2, len(seen))

	// with consume function returning false, only 1 element will be visited.
	firstSeen := make(map[testKey]bool)
	err = newSet.Stream(func(key testKey) error {
		firstSeen[key] = true

		return ErrStopIteration
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrStopIteration)
	require.Equal(t, 1, len(firstSeen))
}
