package randommap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/datastructure/randommap"
	"github.com/iotaledger/hive.go/datastructure/set"
)

func TestRandomMap_RandomUniqueEntries(t *testing.T) {
	testMap := New[string, string](randommap.New())
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

func containsUniqueElements[V any](list []V) bool {
	elementSet := set.New(false)
	for _, element := range list {
		elementSet.Add(element)
	}
	return elementSet.Size() == len(list)
}
