package stream

import (
	"bytes"
)

type ByteReader struct {
	*bytes.Reader
}

func NewByteReader(b []byte) *ByteReader {
	return &ByteReader{
		Reader: bytes.NewReader(b),
	}
}

func (b *ByteReader) BytesRead() int {
	return int(b.Size()) - b.Len()
}
