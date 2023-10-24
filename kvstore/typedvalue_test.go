package kvstore_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func TestTypedValue(t *testing.T) {
	kvStore := mapdb.NewMapDB()
	defer kvStore.Close()

	increase := func(currentValue int, exists bool) (newValue int, err error) {
		if !exists {
			return 1337, nil
		}

		return currentValue + 1, nil
	}

	typedValue := kvstore.NewTypedValue[int](kvStore, []byte("key"), intToBytes, bytesToInt)

	value, err := typedValue.Get()
	require.Equal(t, 0, value)
	require.ErrorIs(t, err, kvstore.ErrKeyNotFound)

	has, err := typedValue.Has()
	require.False(t, has)
	require.NoError(t, err)

	value, err = typedValue.Get()
	require.Equal(t, 0, value)
	require.ErrorIs(t, err, kvstore.ErrKeyNotFound)

	value, err = typedValue.Compute(increase)
	require.Equal(t, 1337, value)
	require.NoError(t, err)

	value, err = typedValue.Compute(increase)
	require.Equal(t, 1338, value)
	require.NoError(t, err)

	value, err = typedValue.Compute(increase)
	require.Equal(t, 1339, value)
	require.NoError(t, err)

	value, err = typedValue.Get()
	require.Equal(t, 1339, value)
	require.NoError(t, err)

	has, err = typedValue.Has()
	require.True(t, has)
	require.NoError(t, err)

	require.NoError(t, typedValue.Delete())

	value, err = typedValue.Get()
	require.Equal(t, 0, value)
	require.ErrorIs(t, err, kvstore.ErrKeyNotFound)

	has, err = typedValue.Has()
	require.False(t, has)
	require.NoError(t, err)

	typedValue.Set(42)
	value, err = typedValue.Get()
	require.Equal(t, 42, value)
	require.NoError(t, err)

	typedValueRestored := kvstore.NewTypedValue[int](kvStore, []byte("key"), intToBytes, bytesToInt)
	has, err = typedValueRestored.Has()
	require.True(t, has)
	require.NoError(t, err)

	value, err = typedValueRestored.Get()
	require.Equal(t, 42, value)
	require.NoError(t, err)
}

func intToBytes(value int) (encoded []byte, err error) {
	encoded = make([]byte, 4)

	binary.LittleEndian.PutUint32(encoded, uint32(value))

	return encoded, nil
}

func bytesToInt(encoded []byte) (value int, consumed int, err error) {
	value = int(binary.LittleEndian.Uint32(encoded))

	return value, 4, nil
}
