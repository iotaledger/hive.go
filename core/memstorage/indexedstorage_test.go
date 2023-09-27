package memstorage_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/memstorage"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

type index uint32

func TestIndexedStorage(t *testing.T) {
	// Create a new IndexedStorage.
	storage := memstorage.NewIndexedStorage[index, string, int]()

	// Test Get on a non-existent index without creating it.
	require.Nil(t, storage.Get(index(1)))

	// Test Get on a non-existent index while creating it.
	subStorage := storage.Get(index(1), true)
	require.NotNil(t, subStorage)

	// Add some values to the latestMilestoneStorage.
	subStorage.Set("tx1", 1)
	subStorage.Set("tx2", 2)

	// Get an existing storage and check if values are correct.
	subStorage1 := storage.Get(index(1))
	require.NotNil(t, subStorage1)
	require.ElementsMatch(t, subStorage.Keys(), subStorage1.Keys())
	require.ElementsMatch(t, subStorage.Values(), subStorage1.Values())

	// Test that Get returns the correct value.
	v, exists := subStorage.Get("tx1")
	require.Equal(t, 1, v)
	require.True(t, exists)

	// Test ForEach.
	count := 0
	storage.ForEach(func(index index, store *shrinkingmap.ShrinkingMap[string, int]) {
		count += store.Size()
	})
	require.Equal(t, 2, count)

	// Test Evict.
	evictedStorage := storage.Evict(1)
	require.Equal(t, subStorage, evictedStorage)

	// Test Get on the evicted index.
	require.Nil(t, storage.Get(1, false))
}
