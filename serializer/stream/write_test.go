package stream_test

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

func requireBufferBytes(t *testing.T, buffer *stream.ByteBuffer, expected []byte) {
	bytesInBuffer, err := buffer.Bytes()
	require.NoError(t, err)

	require.Equal(t, expected, bytesInBuffer)
}

type sampleStruct struct {
	Value0 uint16
	Value1 uint64
}

func (s sampleStruct) Bytes() ([]byte, error) {
	bytes := make([]byte, 10)
	binary.LittleEndian.PutUint16(bytes[:2], s.Value0)
	binary.LittleEndian.PutUint64(bytes[2:], s.Value1)

	return bytes, nil
}

func sampleStructFromBytes(bytes []byte) (sampleStruct, int, error) {
	value0 := binary.LittleEndian.Uint16(bytes[:2])
	value1 := binary.LittleEndian.Uint64(bytes[2:])

	return sampleStruct{value0, value1}, 10, nil
}

func sampleStructFromReader(reader io.ReadSeeker) (sampleStruct, error) {
	bytes := make([]byte, 10)
	_, err := reader.Read(bytes)
	if err != nil {
		return sampleStruct{}, err
	}

	value0 := binary.LittleEndian.Uint16(bytes[:2])
	value1 := binary.LittleEndian.Uint64(bytes[2:])

	return sampleStruct{value0, value1}, nil
}

func TestWrite(t *testing.T) {
	buffer := stream.NewByteBuffer()

	err := stream.Write(buffer, uint64(42))
	require.NoError(t, err)

	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0}
	requireBufferBytes(t, buffer, expected)
}

func TestWriteBytes(t *testing.T) {
	buffer := stream.NewByteBuffer()
	bytesToWrite := []byte{1, 2, 3, 4, 5}

	err := stream.WriteBytes(buffer, bytesToWrite)
	require.NoError(t, err)

	requireBufferBytes(t, buffer, bytesToWrite)
}

func TestWriteBytesWithSize(t *testing.T) {
	buffer := stream.NewByteBuffer()
	bytesToWrite := []byte{1, 2, 3, 4, 5}

	err := stream.WriteBytesWithSize(buffer, bytesToWrite, serializer.SeriLengthPrefixTypeAsUint16)
	require.NoError(t, err)

	expected := []byte{5, 0, 1, 2, 3, 4, 5}
	requireBufferBytes(t, buffer, expected)
}

func TestWriteObject(t *testing.T) {
	buffer := stream.NewByteBuffer()

	s := sampleStruct{42, 1337}
	err := stream.WriteObject(buffer, s, sampleStruct.Bytes)
	require.NoError(t, err)

	expected := lo.PanicOnErr(s.Bytes())
	requireBufferBytes(t, buffer, expected)
}

func TestWriteObjectWithSize(t *testing.T) {
	buffer := stream.NewByteBuffer()

	s := sampleStruct{42, 1337}
	err := stream.WriteObjectWithSize(buffer, s, serializer.SeriLengthPrefixTypeAsUint16, sampleStruct.Bytes)
	require.NoError(t, err)

	expected := byteutils.ConcatBytes([]byte{10, 0}, lo.PanicOnErr(s.Bytes()))
	requireBufferBytes(t, buffer, expected)
}

func TestWriteCollection(t *testing.T) {
	buffer := stream.NewByteBuffer()

	elementsCount := 3
	err := stream.WriteCollection(buffer, serializer.SeriLengthPrefixTypeAsUint32, func() (int, error) {
		for i := 0; i < elementsCount; i++ {
			if err := stream.Write(buffer, uint16(i+1)); err != nil {
				return 0, err
			}
		}
		return elementsCount, nil
	})
	require.NoError(t, err)

	expected := []byte{3, 0, 0, 0, 1, 0, 2, 0, 3, 0}
	requireBufferBytes(t, buffer, expected)
}
