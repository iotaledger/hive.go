package ds_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func TestMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := ads.NewMap(store,
		testKey.Bytes,
		testKeyFromBytes,
		testValue.Bytes,
		testValueFromBytes,
	)
	keys := []testKey{testKey([]byte{'a'}), testKey([]byte{'b'})}
	values := []testValue{testValueFromString("test value"), testValueFromString("test value 1")}
	// Test setting and getting a value
	for i, k := range keys {
		newMap.Set(k, values[i])
	}

	for i, k := range keys {
		exist := newMap.Has(k)
		require.True(t, exist)
		gotValue, exists := newMap.Get(k)
		require.True(t, exists)
		require.ElementsMatch(t, values[i], gotValue)
	}

	// Test setting a value to empty, which should panic
	require.Panics(t, func() { newMap.Set(keys[0], testValue{}) })

	// Test getting a non-existing key
	gotValue, exists := newMap.Get(testKey([]byte{'c'}))
	require.False(t, exists)
	require.Nil(t, gotValue)

	// overwrite the value of keys[0]
	newValue := testValueFromString("test")
	newMap.Set(keys[0], newValue)
	gotValue, exists = newMap.Get(keys[0])
	require.True(t, exists)
	require.ElementsMatch(t, newValue, gotValue)

	// get the root of having 2 keys
	oldRoot := newMap.Root()

	// Test deleting a key
	require.True(t, newMap.Delete(keys[0]))
	exists = newMap.Has(keys[0])
	require.False(t, exists)
	_, exists = newMap.Get(keys[0])
	require.False(t, exists)

	// The root now should be different
	require.NotEqualValues(t, oldRoot, newMap.Root())

	// Test deleting a non-existent key
	require.False(t, newMap.Delete(keys[0]))

	// The root should be same if loading the same store to map
	newMap1 := ads.NewMap(store,
		testKey.Bytes,
		testKeyFromBytes,
		testValue.Bytes,
		testValueFromBytes,
	)
	require.EqualValues(t, newMap.Root(), newMap1.Root())
}

func TestStreamMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := ads.NewMap[testKey, testValue](store,
		testKey.Bytes,
		testKeyFromBytes,
		testValue.Bytes,
		testValueFromBytes,
	)

	kvMap := map[testKey]testValue{
		testKey([]byte{'b'}): testValueFromString("test value 1"),
		testKey([]byte{'c'}): testValueFromString("test value 2"),
	}
	for k, v := range kvMap {
		newMap.Set(k, v)
	}

	seen := make(map[testKey]testValue)
	err := newMap.Stream(func(key testKey, value testValue) bool {
		seen[key] = value
		return true
	})
	require.NoError(t, err)

	require.Equal(t, 2, len(seen))
	for k, v := range seen {
		expectedV, has := kvMap[k]
		require.True(t, has)
		require.ElementsMatch(t, expectedV, v)
	}

	// consume function returns false, only 1 element is visited.
	seenKV := make(map[testKey]testValue)
	err = newMap.Stream(func(key testKey, value testValue) bool {
		seenKV[key] = value

		return false
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(seenKV))
	for k, v := range seenKV {
		expectedV, has := kvMap[k]
		require.True(t, has)
		require.ElementsMatch(t, expectedV, v)
	}
}

type testKey [1]byte

func (t testKey) Bytes() ([]byte, error) {
	return t[:], nil
}

func testKeyFromBytes(b []byte) (testKey, int, error) {
	return testKey(b), 1, nil
}

type testValue []byte

func testValueFromString(s string) testValue {
	return testValue(s)
}

func (t testValue) Bytes() ([]byte, error) {
	return t[:], nil
}

func testValueFromBytes(b []byte) (testValue, int, error) {
	return b, len(b), nil
}
