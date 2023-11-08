package stream_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

func TestByteBuffer_Write(t *testing.T) {
	ws := stream.NewByteBuffer()
	checkWrite(t, ws, "hello", "hello")
	checkWrite(t, ws, " world", "hello world")
}

func TestByteBuffer_Seek(t *testing.T) {
	ws := stream.NewByteBuffer()
	checkWrite(t, ws, "hello", "hello")
	checkWrite(t, ws, " world", "hello world")

	checkSeek(t, ws, -2, io.SeekEnd, len("hello world")-2)
	checkWrite(t, ws, "k!", "hello work!")

	checkSeek(t, ws, 6, io.SeekStart, 6)
	checkWrite(t, ws, "gopher", "hello gopher")

	// Seek back a bit and check that we overwrite the existing buffer before growing it.
	checkSeek(t, ws, -4, io.SeekCurrent, len("hello gopher")-4)
	checkWrite(t, ws, "lang fans", "hello golang fans")

	// If we seek past the end of the buffer, the empty space should be filled with null bytes.
	checkSeek(t, ws, 4, io.SeekCurrent, len("hello golang fans")+4)
	checkWrite(t, ws, "!", "hello golang fans\x00\x00\x00\x00!")
}

func TestByteBuffer_Seek_LargeGap(t *testing.T) {
	ws := stream.NewByteBuffer()
	checkSeek(t, ws, 1024, io.SeekStart, 1024)
	checkWrite(t, ws, "hello", strings.Repeat("\x00", 1024)+"hello")
}

// checkWrite passes data to ws.Write and compares the resulting buffer against exp.
func checkWrite(t *testing.T, byteBuffer *stream.ByteBuffer, data, exp string) {
	nBytesWritten, err := byteBuffer.Write([]byte(data))
	require.NoError(t, err)

	require.EqualValues(t, len(data), nBytesWritten)

	bytes, err := byteBuffer.Bytes()
	require.NoError(t, err)

	require.EqualValues(t, exp, string(bytes))
}

// checkSeek calls ws.Seek with the supplied parameters and compares the returned offset against exp.
func checkSeek(t *testing.T, ws *stream.ByteBuffer, offset int64, whence, exp int) {
	newOffset, err := ws.Seek(offset, whence)
	require.NoError(t, err)

	require.EqualValues(t, exp, newOffset)
}
