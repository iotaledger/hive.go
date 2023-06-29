package byteutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

func TestConcatBytes(t *testing.T) {
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, byteutils.ConcatBytes([]byte{1, 2, 3}, []byte{4, 5}, []byte{7, 8}))
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, byteutils.ConcatBytes([]byte{1, 2, 3, 4, 5, 7, 8}))
}

func TestReadAvailableBytesToBuffer(t *testing.T) {
	t.Run("AvailableBytesLessThanRequired", func(t *testing.T) {
		source := []byte{1, 2, 3, 4, 5}
		target := make([]byte, 5)

		bytesRead := byteutils.ReadAvailableBytesToBuffer(target, 0, source, 0, 10)
		require.Equalf(t, len(source), bytesRead, "Expected %d bytes to be read, but got %d", len(source), bytesRead)

	})

	t.Run("AvailableBytesGreaterThanRequired", func(t *testing.T) {
		source := []byte{1, 2, 3, 4, 5}
		target := make([]byte, 3)

		bytesRead := byteutils.ReadAvailableBytesToBuffer(target, 0, source, 0, 3)
		require.Equalf(t, 3, bytesRead, "Expected 3 bytes to be read, but got %d", bytesRead)

		expectedTarget := []byte{1, 2, 3}
		require.ElementsMatchf(t, expectedTarget, target, "Target byte slice does not match the required bytes")
	})
}

func TestConcatBytesToString(t *testing.T) {
	byteSlice1 := []byte{'h', 'e', 'l', 'l', 'o'}
	byteSlice2 := []byte{' ', 'w', 'o', 'r', 'l', 'd'}

	result := byteutils.ConcatBytesToString(byteSlice1, byteSlice2)
	expectedResult := "hello world"
	require.Equalf(t, expectedResult, result, "Concatenated string does not match the expected result")
}
