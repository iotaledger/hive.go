package storable

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type Slice[A any, B serializer.MarshalablePtr[A]] struct {
	byteSlice *ByteSlice
}

func NewSlice[A any, B serializer.MarshalablePtr[A]](fileName string, entrySize int, opts ...options.Option[ByteSlice]) (indexedFile *Slice[A, B], err error) {
	byteSlice, err := NewByteSlice(fileName, entrySize, opts...)
	if err != nil {
		return nil, err
	}

	return &Slice[A, B]{
		byteSlice: byteSlice,
	}, nil
}

func (s *Slice[A, B]) Set(index int, entry B) (err error) {
	serializedEntry, err := entry.Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to serialize entry")
	}

	return s.byteSlice.Set(index, serializedEntry)
}

func (s *Slice[A, B]) Get(index int) (entry B, err error) {
	entryBytes, err := s.byteSlice.Get(index)
	if err != nil {
		return entry, err
	}

	var newEntry B = new(A)
	if _, err = newEntry.FromBytes(entryBytes); err != nil {
		return entry, errors.Wrap(err, "failed to deserialize entry")
	}
	entry = newEntry

	return
}

func (s *Slice[A, B]) Close() (err error) {
	return s.byteSlice.Close()
}
