// Adapted (rename and added methods) from https://github.com/orcaman/writerseeker.
//
// The MIT License (MIT)
//
// Copyright (c) 2017 Or Hiltch
// Copyright (c) 2017 icza (https://stackoverflow.com/users/1705598/icza)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package stream

import (
	"bytes"
	"io"

	"github.com/iotaledger/hive.go/ierrors"
)

// ByteBuffer is an in-memory io.WriteSeeker implementation.
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

// Write writes to the buffer of this ByteBuffer instance.
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

// Seek seeks in the buffer of this ByteBuffer instance.
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
		return 0, ierrors.New("negative result pos")
	}
	w.pos = newPos

	return int64(newPos), nil
}

func (w *ByteBuffer) Reader() *ByteReader {
	return NewByteReader(w.buf.Bytes())
}

func (w *ByteBuffer) Close() error {
	return nil
}

func (w *ByteBuffer) Bytes() ([]byte, error) {
	return w.buf.Bytes(), nil
}
