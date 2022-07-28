package randommap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/core/datastructure/set"
)

func TestRandomMap_RandomUniqueEntries(t *testing.T) {
	testMap := New[string, string]()
	// key and value are the same for the sake of the test
	keysAndValues := []string{"a", "b", "c", "d"}
	// fill randomMap
	for _, keyValue := range keysAndValues {
		testMap.Set(keyValue, keyValue)
	}

	var emptyResult = []string{}

	result := testMap.RandomUniqueEntries(0)
	assert.Equal(t, emptyResult, result)

	result = testMap.RandomUniqueEntries(-5)
	assert.Equal(t, emptyResult, result)

	result = testMap.RandomUniqueEntries(2)
	assert.Equal(t, 2, len(result))
	assert.True(t, containsUniqueElements(result))

	result = testMap.RandomUniqueEntries(100)
	assert.Equal(t, 4, len(result))
	assert.True(t, containsUniqueElements(result))
}
func TestRandomMap_EmptyMap(t *testing.T) {
	testMap := New[string, string]()
	var emptyResult = []string{}

	result := testMap.RandomUniqueEntries(4)
	assert.Equal(t, emptyResult, result)

	randEntry, exists := testMap.RandomEntry()
	assert.False(t, exists)
	assert.Equal(t, "", randEntry)

	assert.Panics(t, func() {
		testMap.RandomKey()
	})
}
func containsUniqueElements[V any](list []V) bool {
	elementSet := set.New(false)
	for _, element := range list {
		elementSet.Add(element)
	}
	return elementSet.Size() == len(list)
}
