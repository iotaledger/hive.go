package ads_test

import (
	"testing"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := ads.NewSet[testKey](store)

	key := testKey([]byte{'a'})
	newSet.Add(key)
	exist := newSet.Has(key)
	assert.True(t, exist)

	// Test deleting a key
	assert.True(t, newSet.Delete(key))
	exist = newSet.Has(key)
	assert.False(t, exist)

	// Test deleting a non-existent key
	assert.False(t, newSet.Delete(key))

	assert.Equal(t, 0, newSet.Size())
}

func TestStreamSet(t *testing.T) {
	store := mapdb.NewMapDB()
	newSet := ads.NewSet[testKey](store)

	key1 := testKey([]byte{'b'})
	key2 := testKey([]byte{'c'})
	newSet.Add(key1)
	newSet.Add(key2)
	assert.Equal(t, 2, newSet.Size())

	seen := make(map[testKey]bool)
	err := newSet.Stream(func(key testKey) bool {
		seen[key] = true
		return true
	})
	assert.NoError(t, err)
	assert.True(t, seen[key1])
	assert.True(t, seen[key2])
	assert.Equal(t, 2, len(seen))
}
