package serializableorderedmap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds/serializableorderedmap"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func TestSerialization(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	serializableOrderedMap := serializableorderedmap.New[string, uint8]()

	serializableOrderedMap.Set("a", 0)
	serializableOrderedMap.Set("b", 1)
	serializableOrderedMap.Set("c", 2)

	bytes, err := serializableOrderedMap.Encode(serix.DefaultAPI)
	require.NoError(t, err)

	decoded := serializableorderedmap.New[string, uint8]()
	bytesRead, err := decoded.Decode(serix.DefaultAPI, bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), bytesRead)

	require.Equal(t, serializableOrderedMap, decoded)
}
