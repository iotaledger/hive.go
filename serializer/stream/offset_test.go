package stream_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

func TestOffset(t *testing.T) {
	buffer := stream.NewByteReader([]byte{1, 2, 3, 4, 5})

	offset, err := stream.Offset(buffer)
	require.NoError(t, err)
	require.EqualValues(t, 0, offset)

	{
		_, err = buffer.Read([]byte{1})
		require.NoError(t, err)

		offset, err = stream.Offset(buffer)
		require.NoError(t, err)
		require.EqualValues(t, 1, offset)
	}

	{
		newOffset, err := stream.Skip(buffer, 3)
		require.NoError(t, err)
		require.EqualValues(t, 4, newOffset)

		offset, err = stream.Offset(buffer)
		require.NoError(t, err)
		require.EqualValues(t, 4, offset)
	}

	{
		newOffset, err := stream.GoTo(buffer, 2)
		require.NoError(t, err)
		require.EqualValues(t, 2, newOffset)

		offset, err = stream.Offset(buffer)
		require.NoError(t, err)
		require.EqualValues(t, 2, offset)
	}
}
