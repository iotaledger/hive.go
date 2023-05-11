package memstorage_test

import (
	"testing"

	"github.com/iotaledger/hive.go/core/memstorage"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/stretchr/testify/require"
)

func TestIndexedStorage(t *testing.T) {
	// Create a new IndexedStorage.
	storage := memstorage.NewIndexedStorage[iotago.SlotIndex, string, int]()

	// Test Get on a non-existent index without creating it.
	require.Nil(t, storage.Get(iotago.SlotIndex(1)))

	// Test Get on a non-existent index while creating it.
	subStorage := storage.Get(iotago.SlotIndex(1), true)
	require.NotNil(t, subStorage)

	// Add some values to the latestMilestoneStorage.
	subStorage.Set("tx1", 1)
	subStorage.Set("tx2", 2)

	// Test that Get returns the correct value.
	v, exists := subStorage.Get("tx1")
	require.Equal(t, 1, v)
	require.True(t, exists)

	// Test ForEach.
	count := 0
	storage.ForEach(func(index iotago.SlotIndex, store *shrinkingmap.ShrinkingMap[string, int]) {
		count += store.Size()
	})
	require.Equal(t, 2, count)

	// Test Evict.
	evictedStorage := storage.Evict(1)
	require.Equal(t, subStorage, evictedStorage)

	// Test Get on the evicted index.
	require.Nil(t, storage.Get(1, false))
}
