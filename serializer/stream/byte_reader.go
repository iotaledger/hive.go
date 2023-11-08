package stream

import (
	"bytes"
)

// ByteReader is a wrapper around bytes.Reader that provides the BytesRead method to get the number of bytes read.
type ByteReader struct {
	*bytes.Reader
}

// NewByteReader creates a new ByteReader from a byte slice.
func NewByteReader(b []byte) *ByteReader {
	return &ByteReader{
		Reader: bytes.NewReader(b),
	}
}

// BytesRead returns the number of bytes read from the underlying reader.
func (b *ByteReader) BytesRead() int {
	return int(b.Size()) - b.Len()
}
