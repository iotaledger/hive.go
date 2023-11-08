package stream_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

func TestByteReaderBytesRead(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected int
	}{
		{[]byte{1, 2, 3, 4, 5}, 0}, // No bytes read initially
		{[]byte{1, 2, 3, 4, 5}, 3}, // 3 bytes read after reading 3 bytes
		{[]byte{1, 2, 3, 4, 5}, 5}, // All 5 bytes read after reading all bytes
	}

	for _, tc := range testCases {
		bytes := make([]byte, tc.expected)

		reader := stream.NewByteReader(tc.input)

		nBytes, err := reader.Read(bytes)
		require.NoError(t, err)
		require.Equal(t, tc.expected, nBytes)

		require.Equal(t, tc.input[:tc.expected], bytes[:tc.expected])

		require.Equal(t, tc.expected, reader.BytesRead())
	}
}
