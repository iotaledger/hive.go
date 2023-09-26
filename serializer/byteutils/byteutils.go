package byteutils

import (
	"bytes"
	"strings"
)

func ReadAvailableBytesToBuffer(target []byte, targetOffset int, source []byte, sourceOffset int, sourceLength int) int {
	availableBytes := sourceLength - sourceOffset
	requiredBytes := len(target) - targetOffset

	var bytesToRead int
	if availableBytes < requiredBytes {
		bytesToRead = availableBytes
	} else {
		bytesToRead = requiredBytes
	}

	copy(target[targetOffset:], source[sourceOffset:sourceOffset+bytesToRead])

	return bytesToRead
}

// ConcatBytes concatenates the byte slices into a new byte slice.
func ConcatBytes(byteSlices ...[]byte) []byte {
	var b bytes.Buffer
	for _, byteSlice := range byteSlices {
		b.Write(byteSlice)
	}

	return b.Bytes()
}

// ConcatBytesToString concatenates the byte slices into a string.
func ConcatBytesToString(byteSlices ...[]byte) string {
	var b strings.Builder
	for _, byteSlice := range byteSlices {
		b.Write(byteSlice)
	}

	return b.String()
}
