package orderedmap_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds/orderedmap"
)

func TestOrderedMap_Size(t *testing.T) {
	orderedMap := orderedmap.New[int, int]()

	require.Equal(t, 0, orderedMap.Size())
	require.True(t, orderedMap.IsEmpty())

	orderedMap.Set(1, 1)

	require.Equal(t, 1, orderedMap.Size())

	orderedMap.Set(3, 1)
	orderedMap.Set(2, 1)

	require.Equal(t, 3, orderedMap.Size())
	require.False(t, orderedMap.IsEmpty())

	orderedMap.Set(2, 2)

	require.Equal(t, 3, orderedMap.Size())

	orderedMap.Delete(2)

	require.Equal(t, 2, orderedMap.Size())

	expectedMap := map[int]int{}
	orderedMap.ForEach(func(key int, value int) bool {
		expectedMap[key] = value
		return true
	})

	clone := orderedMap.Clone()

	clonedMap := map[int]int{}
	clone.ForEach(func(key int, value int) bool {
		clonedMap[key] = value
		return true
	})

	// We compare the maps because EqualValues does a deep compare and the original has a deleted items count inside the ShrinkingMap
	require.EqualValues(t, expectedMap, clonedMap)

	clone.Clear()
	require.True(t, clone.IsEmpty())
	require.False(t, orderedMap.IsEmpty())
}

func TestNew(t *testing.T) {
	orderedMap := orderedmap.New[int, int]()
	require.NotNil(t, orderedMap)

	require.Equal(t, 0, orderedMap.Size())

	_, _, exists := orderedMap.Head()
	require.False(t, exists)

	_, _, exists = orderedMap.Tail()
	require.False(t, exists)
}

func TestSetGetDelete(t *testing.T) {
	orderedMap := orderedmap.New[string, string]()
	require.NotNil(t, orderedMap)

	// when adding the first new key,value pair, we must return false
	_, previousValueExisted := orderedMap.Set("key", "value")
	require.False(t, previousValueExisted)

	// we should be able to retrieve the just added element
	value, ok := orderedMap.Get("key")
	require.Equal(t, "value", value)
	require.True(t, ok)

	// head and tail should NOT be nil and match and size should be 1
	k, v, exists := orderedMap.Head()
	require.True(t, exists)
	require.Equal(t, "key", k)
	require.Equal(t, "value", v)

	k, v, exists = orderedMap.Tail()
	require.True(t, exists)
	require.Equal(t, "key", k)
	require.Equal(t, "value", v)

	require.Equal(t, 1, orderedMap.Size())

	// when adding the same key,value pair must return true
	// and size should not change;
	_, previousValueExisted = orderedMap.Set("key", "value")
	require.True(t, previousValueExisted)
	require.Equal(t, 1, orderedMap.Size())

	// when retrieving something that does not exist we
	// should get nil, false
	value, ok = orderedMap.Get("keyNotStored")
	require.Empty(t, value)
	require.False(t, ok)

	// when deleting an existing element, we must get true,
	// the element must be removed, and size decremented.
	deleted := orderedMap.Delete("key")
	require.True(t, deleted)
	value, ok = orderedMap.Get("key")
	require.Empty(t, value)
	require.False(t, ok)
	require.Equal(t, 0, orderedMap.Size())

	// if we delete the only element, head and tail should be both nil
	_, _, exists = orderedMap.Head()
	require.False(t, exists)

	_, _, exists = orderedMap.Tail()
	require.False(t, exists)

	// when deleting a NON existing element, we must get false
	deleted = orderedMap.Delete("key")
	require.False(t, deleted)
}

func TestForEach(t *testing.T) {
	orderedMap := orderedmap.New[string, int]()
	require.NotNil(t, orderedMap)

	keys := []string{"one", "two", "three"}
	values := []int{1, 2, 3}

	for i := 0; i < len(keys); i++ {
		orderedMap.Set(keys[i], values[i])
	}

	// test that all elements are positive via ForEach
	testPositive := orderedMap.ForEach(func(key string, value int) bool {
		return value > 0
	})
	require.True(t, testPositive)

	testNegative := orderedMap.ForEach(func(key string, value int) bool {
		return value < 0
	})
	require.False(t, testNegative)

	j := len(keys) - 1
	revKeys := make([]string, len(keys))
	revValues := make([]int, len(keys))
	orderedMap.ForEachReverse(func(key string, value int) bool {
		revKeys[j] = key
		revValues[j] = value
		j--

		return true
	})

	require.ElementsMatch(t, keys, revKeys)
	require.ElementsMatch(t, values, revValues)
}

func TestConcurrencySafe(t *testing.T) {
	orderedMap := orderedmap.New[string, int]()
	require.NotNil(t, orderedMap)

	count := 100
	keys := make([]string, count)
	values := make([]int, count)

	// initialize a slice of 100 elements
	for i := 0; i < 100; i++ {
		keys[i] = fmt.Sprintf("%d", i)
		values[i] = i
	}

	// let 10 workers fill the orderedMap
	workers := 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				orderedMap.Set(keys[i], values[i])
			}
		}()
	}
	wg.Wait()

	// check that all the elements consumed from the set
	// have been stored in the orderedMap and its size matches
	for i := 0; i < 100; i++ {
		value, ok := orderedMap.Get(keys[i])
		require.Equal(t, values[i], value)
		require.True(t, ok)
	}
	require.Equal(t, 100, orderedMap.Size())

	// let 10 workers delete elements from the orderedMAp
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				orderedMap.Delete(keys[i])
			}
		}()
	}
	wg.Wait()

	require.Equal(t, 0, orderedMap.Size())
}
