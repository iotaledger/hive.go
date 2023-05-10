package ads_test

import (
	"testing"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := ads.NewMap[testKey, testValue](store)

	// Test setting and getting a value
	key := testKey([]byte{'a'})
	value := testValueFromString("test value")
	newMap.Set(key, &value)
	exist := newMap.Has(key)
	assert.True(t, exist)
	gotValue, exists := newMap.Get(key)
	assert.True(t, exists)
	assert.ElementsMatch(t, value, *gotValue)

	// Test setting a value to empty, which should panic
	assert.Panics(t, func() { newMap.Set(key, &testValue{}) })

	// Test deleting a key
	assert.True(t, newMap.Delete(key))
	exist = newMap.Has(key)
	assert.False(t, exist)
	_, exists = newMap.Get(key)
	assert.False(t, exists)

	// Test deleting a non-existent key
	assert.False(t, newMap.Delete(key))
}

func TestStreamMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := ads.NewMap[testKey, testValue](store)

	key1 := testKey([]byte{'b'})
	key2 := testKey([]byte{'c'})
	value1 := testValueFromString("test value 1")
	value2 := testValueFromString("test value 2")
	newMap.Set(key1, &value1)
	newMap.Set(key2, &value2)

	seen := make(map[testKey]bool)
	err := newMap.Stream(func(key testKey, value *testValue) bool {
		seen[key] = true
		if key == key1 {
			assert.ElementsMatch(t, value1, *value)
		} else if key == key2 {
			assert.ElementsMatch(t, value2, *value)
		} else {
			t.Fail()
			return false
		}
		return true
	})
	assert.NoError(t, err)
	assert.True(t, seen[key1])
	assert.True(t, seen[key2])
	assert.Equal(t, 2, len(seen))
}

type testKey [1]byte

func (t testKey) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t *testKey) FromBytes(b []byte) (int, error) {
	copy(t[:], b)
	return len(t), nil
}

type testValue []byte

func testValueFromString(s string) testValue {
	return testValue([]byte(s))
}

func (t testValue) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t *testValue) FromBytes(b []byte) (int, error) {
	*t = testValue(b)
	return len(*t), nil
}
