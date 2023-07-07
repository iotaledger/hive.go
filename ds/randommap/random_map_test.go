package randommap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/ds/randommap"
)

func TestRandomMap_Basics(t *testing.T) {
	testMap := randommap.New[string, string]()
	keysAndValues := []string{"a", "b", "c", "d"}

	// fill randomMap
	for _, keyValue := range keysAndValues {
		testMap.Set(keyValue, keyValue)
	}
	require.Equal(t, len(keysAndValues), testMap.Size())
	require.ElementsMatch(t, keysAndValues, testMap.Keys())

	for index, value := range keysAndValues {
		exists := testMap.Has(keysAndValues[index])
		require.Truef(t, exists, "%s should exists in randommap, got false", value)

		result, exists := testMap.Get(keysAndValues[index])
		require.Truef(t, exists, "get %s from randommap failed", value)
		require.Equal(t, keysAndValues[index], result)
	}

	key, exists := testMap.RandomKey()
	require.True(t, exists)
	require.Contains(t, keysAndValues, key)
	entry, exists := testMap.RandomEntry()
	require.True(t, exists)
	require.Contains(t, keysAndValues, entry)

	value, deleted := testMap.Delete(keysAndValues[0])
	require.Truef(t, deleted, "%s deleted failed", value)
	require.Equal(t, len(keysAndValues)-1, testMap.Size())

	// update existing key value
	testMap.Set(keysAndValues[1], "x")
	require.Equal(t, len(keysAndValues)-1, testMap.Size())
	result, exists := testMap.Get(keysAndValues[1])
	require.Truef(t, exists, "get %s from randommap failed", value)
	require.Equal(t, "x", result)

}

func TestRandomMap_RandomUniqueEntries(t *testing.T) {
	testMap := randommap.New[string, string]()
	// key and value are the same for the sake of the test
	keysAndValues := []string{"a", "b", "c", "d"}
	// fill randomMap
	for _, keyValue := range keysAndValues {
		testMap.Set(keyValue, keyValue)
	}

	var nilResult []string

	result := testMap.RandomUniqueEntries(0)
	assert.Equal(t, nilResult, result)

	result = testMap.RandomUniqueEntries(-5)
	assert.Equal(t, nilResult, result)

	result = testMap.RandomUniqueEntries(2)
	assert.Equal(t, 2, len(result))
	assert.True(t, containsUniqueElements(result))

	result = testMap.RandomUniqueEntries(100)
	assert.Equal(t, 4, len(result))
	assert.True(t, containsUniqueElements(result))
}

func TestRandomMap_EmptyMap(t *testing.T) {
	testMap := randommap.New[string, string]()
	var emptyResult = []string{}

	result := testMap.RandomUniqueEntries(4)
	assert.Equal(t, emptyResult, result)

	randEntry, exists := testMap.RandomEntry()
	assert.False(t, exists)
	assert.Equal(t, "", randEntry)

	_, exists = testMap.RandomKey()
	assert.False(t, exists)
}

func containsUniqueElements[V comparable](list []V) bool {
	return ds.NewSet[V](list...).Size() == len(list)
}
