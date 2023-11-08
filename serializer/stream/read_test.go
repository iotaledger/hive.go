package stream_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

func TestRead(t *testing.T) {
	buffer := bytes.NewReader([]byte{42, 0, 0, 0, 0, 0, 0, 0})

	result, err := stream.Read[uint64](buffer)

	require.NoError(t, err)
	require.EqualValues(t, 42, result)
}

func TestReadBytes(t *testing.T) {
	initialBytes := []byte{1, 2, 3, 4, 5}
	buffer := bytes.NewReader(initialBytes)

	readBytes, err := stream.ReadBytes(buffer, 5)
	require.NoError(t, err)
	require.EqualValues(t, initialBytes, readBytes)
}

func TestReadBytesWithSize(t *testing.T) {
	initialBytes := []byte{5, 0, 1, 2, 3, 4, 5}
	buffer := bytes.NewReader(initialBytes)

	readBytes, err := stream.ReadBytesWithSize(buffer, serializer.SeriLengthPrefixTypeAsUint16)
	require.NoError(t, err)

	require.EqualValues(t, []byte{1, 2, 3, 4, 5}, readBytes)
}

func TestReadObject(t *testing.T) {
	buffer := bytes.NewReader([]byte{42, 0, 57, 5, 0, 0, 0, 0, 0, 0})

	result, err := stream.ReadObject(buffer, 10, sampleStructFromBytes)
	require.NoError(t, err)

	expected := sampleStruct{42, 1337}
	require.EqualValues(t, expected, result)
}

func TestReadObjectWithSize(t *testing.T) {
	buffer := bytes.NewReader([]byte{10, 0, 42, 0, 57, 5, 0, 0, 0, 0, 0, 0})

	result, err := stream.ReadObjectWithSize(buffer, serializer.SeriLengthPrefixTypeAsUint16, sampleStructFromBytes)
	require.NoError(t, err)

	expected := sampleStruct{42, 1337}
	require.EqualValues(t, expected, result)
}

func TestReadObjectFromReader(t *testing.T) {
	buffer := bytes.NewReader([]byte{42, 0, 57, 5, 0, 0, 0, 0, 0, 0})

	result, err := stream.ReadObjectFromReader(buffer, sampleStructFromReader)
	require.NoError(t, err)

	expected := sampleStruct{42, 1337}
	require.EqualValues(t, expected, result)
}

func TestPeek(t *testing.T) {
	buffer := bytes.NewReader([]byte{3, 0, 0, 0, 1, 0, 2, 0, 3, 0})

	elementsCount, err := stream.PeekSize(buffer, serializer.SeriLengthPrefixTypeAsUint32)
	require.NoError(t, err)
	require.EqualValues(t, 3, elementsCount)
}

func TestReadCollection(t *testing.T) {
	buffer := bytes.NewReader([]byte{3, 0, 0, 0, 1, 0, 2, 0, 3, 0})

	count, err := stream.PeekSize(buffer, serializer.SeriLengthPrefixTypeAsUint32)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)

	results := make([]uint16, count)

	err = stream.ReadCollection(buffer, serializer.SeriLengthPrefixTypeAsUint32, func(i int) error {
		result, err := stream.Read[uint16](buffer)
		require.NoError(t, err)

		results[i] = result

		return nil
	})
	require.NoError(t, err)

	expected := []uint16{1, 2, 3}
	require.EqualValues(t, expected, results)
}
