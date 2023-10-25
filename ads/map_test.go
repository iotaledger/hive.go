package ads

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/typeutils"
)

var ErrStopIteration = ierrors.New("stop")

func TestMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := newAuthenticatedMap[[32]byte](store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
		testKey.Bytes,
		testKeyFromBytes,
		testValue.Bytes,
		testValueFromBytes,
	)
	keys := []testKey{testKey([]byte{'a'}), testKey([]byte{'b'})}
	values := []testValue{testValueFromString("test value"), testValueFromString("test value 1")}
	// Test setting and getting a value
	require.Equal(t, 0, newMap.Size())
	require.False(t, newMap.WasRestoredFromStorage())

	for i, k := range keys {
		require.NoError(t, newMap.Set(k, values[i]))
	}

	for i, k := range keys {
		exist, err := newMap.Has(k)
		require.NoError(t, err)
		require.True(t, exist)
		gotValue, exists, err := newMap.Get(k)
		require.NoError(t, err)
		require.True(t, exists)
		require.ElementsMatch(t, values[i], gotValue)
	}

	require.Equal(t, len(keys), newMap.Size())

	// Test setting a value to empty, which should be just fine
	require.NoError(t, newMap.Set(keys[0], testValue{}))

	// Test getting a non-existing key
	gotValue, exists, err := newMap.Get(testKey([]byte{'c'}))
	require.NoError(t, err)
	require.False(t, exists)
	require.Nil(t, gotValue)

	// overwrite the value of keys[0]
	newValue := testValueFromString("test")
	require.NoError(t, newMap.Set(keys[0], newValue))
	gotValue, exists, err = newMap.Get(keys[0])
	require.NoError(t, err)
	require.True(t, exists)
	require.ElementsMatch(t, newValue, gotValue)

	// get the root of having 2 keys
	oldRoot := newMap.Root()

	// Test deleting a key
	require.True(t, lo.PanicOnErr(newMap.Delete(keys[0])))
	exists, err = newMap.Has(keys[0])
	require.NoError(t, err)
	require.False(t, exists)
	_, exists, err = newMap.Get(keys[0])
	require.NoError(t, err)
	require.False(t, exists)

	// The root now should be different
	require.NotEqualValues(t, oldRoot, newMap.Root())

	// Test deleting a non-existent key
	require.False(t, lo.PanicOnErr(newMap.Delete(keys[0])))

	require.NoError(t, newMap.Commit())

	// The root should be same if loading the same store to map
	newMap1 := newAuthenticatedMap[[32]byte](store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
		testKey.Bytes,
		testKeyFromBytes,
		testValue.Bytes,
		testValueFromBytes,
	)

	require.True(t, newMap.WasRestoredFromStorage())
	require.EqualValues(t, newMap.Root(), newMap1.Root())
}

func TestStreamMap(t *testing.T) {
	store := mapdb.NewMapDB()
	newMap := newAuthenticatedMap[[32]byte, testKey, testValue](store,
		typeutils.ByteArray32ToBytes,
		typeutils.ByteArray32FromBytes,
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
		require.NoError(t, newMap.Set(k, v))
	}

	seen := make(map[testKey]testValue)
	err := newMap.Stream(func(key testKey, value testValue) error {
		seen[key] = value

		return nil
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
	err = newMap.Stream(func(key testKey, value testValue) error {
		seenKV[key] = value

		return ErrStopIteration
	})
	// the error is expected because we stopped the iteration early
	require.Error(t, err)
	require.ErrorIs(t, err, ErrStopIteration)
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
