package stream

import (
	"bytes"
	"errors"
	"io"
)

// ByteBuffer is an in-memory io.WriteSeeker implementation.
// Adapted from https://github.com/orcaman/writerseeker.
type ByteBuffer struct {
	buf *bytes.Buffer
	pos int
}

func NewByteBuffer(initialLength ...int) *ByteBuffer {
	var length int
	if len(initialLength) > 0 {
		length = initialLength[0]
	}

	return &ByteBuffer{
		buf: bytes.NewBuffer(make([]byte, length)),
	}
}

// Write writes to the buffer of this ByteBuffer instance
func (w *ByteBuffer) Write(p []byte) (n int, err error) {
	// If the offset is past the end of the buffer, grow the buffer with null bytes.
	if extra := w.pos - w.buf.Len(); extra > 0 {
		if _, err := w.buf.Write(make([]byte, extra)); err != nil {
			return n, err
		}
	}

	// If the offset isn't at the end of the buffer, write as much as we can.
	if w.pos < w.buf.Len() {
		n = copy(w.buf.Bytes()[w.pos:], p)
		p = p[n:]
	}

	// If there are remaining bytes, append them to the buffer.
	if len(p) > 0 {
		var bn int
		bn, err = w.buf.Write(p)
		n += bn
	}

	w.pos += n
	return n, err
}

// Seek seeks in the buffer of this ByteBuffer instance
func (w *ByteBuffer) Seek(offset int64, whence int) (int64, error) {
	newPos, offs := 0, int(offset)

	switch whence {
	case io.SeekStart:
		newPos = offs
	case io.SeekCurrent:
		newPos = w.pos + offs
	case io.SeekEnd:
		newPos = w.buf.Len() + offs
	}

	if newPos < 0 {
		return 0, errors.New("negative result pos")
	}
	w.pos = newPos

	return int64(newPos), nil
}

func (w *ByteBuffer) Reader() *ByteReader {
	return NewByteReader(w.buf.Bytes())
}

// Close :
func (w *ByteBuffer) Close() error {
	return nil
}

func (w *ByteBuffer) Bytes() ([]byte, error) {
	return w.buf.Bytes(), nil
}
