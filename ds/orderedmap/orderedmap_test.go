package orderedmap

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func TestOrderedMap_Size(t *testing.T) {
	orderedMap := New[int, int]()

	require.Equal(t, 0, orderedMap.Size())

	orderedMap.Set(1, 1)

	require.Equal(t, 1, orderedMap.Size())

	orderedMap.Set(3, 1)
	orderedMap.Set(2, 1)

	require.Equal(t, 3, orderedMap.Size())

	orderedMap.Set(2, 2)

	require.Equal(t, 3, orderedMap.Size())

	orderedMap.Delete(2)

	require.Equal(t, 2, orderedMap.Size())
}

func TestNew(t *testing.T) {
	orderedMap := New[int, int]()
	require.NotNil(t, orderedMap)

	require.Equal(t, 0, orderedMap.Size())

	require.Nil(t, orderedMap.head)
	require.Nil(t, orderedMap.tail)
}

func TestSetGetDelete(t *testing.T) {
	orderedMap := New[string, string]()
	require.NotNil(t, orderedMap)

	// when adding the first new key,value pair, we must return false
	_, previousValueExisted := orderedMap.Set("key", "value")
	require.False(t, previousValueExisted)

	// we should be able to retrieve the just added element
	value, ok := orderedMap.Get("key")
	require.Equal(t, "value", value)
	require.True(t, ok)

	// head and tail should NOT be nil and match and size should be 1
	require.NotNil(t, orderedMap.head)
	require.Same(t, orderedMap.head, orderedMap.tail)
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
	require.Nil(t, orderedMap.head)
	require.Same(t, orderedMap.head, orderedMap.tail)

	// when deleting a NON existing element, we must get false
	deleted = orderedMap.Delete("key")
	require.False(t, deleted)
}

func TestForEach(t *testing.T) {
	orderedMap := New[string, int]()
	require.NotNil(t, orderedMap)

	testElements := []Element[string, int]{
		{key: "one", value: 1},
		{key: "two", value: 2},
		{key: "three", value: 3},
	}

	for _, element := range testElements {
		_, previousValueExisted := orderedMap.Set(element.key, element.value)
		require.False(t, previousValueExisted)
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
}

func TestConcurrencySafe(t *testing.T) {
	orderedMap := New[string, int]()
	require.NotNil(t, orderedMap)

	// initialize a slice of 100 elements
	set := make([]Element[string, int], 100)
	for i := 0; i < 100; i++ {
		element := Element[string, int]{key: fmt.Sprintf("%d", i), value: i}
		set[i] = element
	}

	// let 10 workers fill the orderedMap
	workers := 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				ele := set[i]
				orderedMap.Set(ele.key, ele.value)
			}
		}()
	}
	wg.Wait()

	// check that all the elements consumed from the set
	// have been stored in the orderedMap and its size matches
	for i := 0; i < 100; i++ {
		value, ok := orderedMap.Get(set[i].key)
		require.Equal(t, set[i].value, value)
		require.True(t, ok)
	}
	require.Equal(t, 100, orderedMap.Size())

	// let 10 workers delete elements from the orderedMAp
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				ele := set[i]
				orderedMap.Delete(ele.key)
			}
		}()
	}
	wg.Wait()

	require.Equal(t, 0, orderedMap.Size())
}

func TestSerialization(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	orderedMap := New[string, uint8]()

	orderedMap.Set("a", 0)
	orderedMap.Set("b", 1)
	orderedMap.Set("c", 2)

	bytes, err := orderedMap.Encode()
	require.NoError(t, err)

	decoded := new(OrderedMap[string, uint8])
	bytesRead, err := decoded.Decode(bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), bytesRead)

	require.Equal(t, orderedMap, decoded)
}
