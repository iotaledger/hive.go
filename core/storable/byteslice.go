package storable

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

const SliceOffsetAuto = ^uint64(0)

type ByteSlice struct {
	fileHandle  *os.File
	startOffset uint64
	entrySize   uint64

	sync.RWMutex
}

func NewByteSlice(fileName string, entrySize uint64, opts ...options.Option[ByteSlice]) (indexedFile *ByteSlice, err error) {
	return options.Apply(new(ByteSlice), opts, func(i *ByteSlice) {
		if i.fileHandle, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0o666); err != nil {
			err = ierrors.Wrap(err, "failed to open file")
			return
		}

		i.entrySize = entrySize

		if err = i.readHeader(); err != nil {
			err = ierrors.Wrap(err, "failed to read header")
			return
		}
	}), err
}

func (b *ByteSlice) EntrySize() uint64 {
	return b.entrySize
}

func (b *ByteSlice) Set(index uint64, entry []byte) (err error) {
	b.Lock()
	defer b.Unlock()

	if uint64(len(entry)) != b.entrySize {
		return ierrors.Errorf("entry has wrong length %d vs %d", len(entry), b.entrySize)
	}

	if b.startOffset == SliceOffsetAuto {
		b.startOffset = index

		if err = b.writeHeader(); err != nil {
			return ierrors.Wrap(err, "failed to write header")
		}
	}

	relativeIndex := index - b.startOffset
	if relativeIndex < 0 {
		return ierrors.Errorf("index %d is out of bounds", index)
	}

	if _, err = b.fileHandle.WriteAt(entry, int64(8+relativeIndex*b.entrySize)); err != nil {
		return ierrors.Wrap(err, "failed to write entry")
	}

	return b.fileHandle.Sync()
}

func (b *ByteSlice) Get(index uint64) (entry []byte, err error) {
	relativeIndex := index - b.startOffset
	if relativeIndex < 0 {
		return nil, ierrors.Errorf("index %d is out of bounds", index)
	}

	entryBytes := make([]byte, b.entrySize)
	if _, err = b.fileHandle.ReadAt(entryBytes, int64(8+relativeIndex*b.entrySize)); err != nil {
		return nil, ierrors.Wrap(err, "failed to read entry")
	}

	return entryBytes, nil
}

func (b *ByteSlice) Close() (err error) {
	return b.fileHandle.Close()
}

func (b *ByteSlice) readHeader() (err error) {
	startOffsetBytes := make([]byte, 8)
	if _, err = b.fileHandle.ReadAt(startOffsetBytes, 0); err != nil {
		if ierrors.Is(err, io.EOF) {
			return nil
		}

		return ierrors.Wrap(err, "failed to read start offset")
	}

	var startOffset uint64
	if _, err = serix.DefaultAPI.Decode(context.Background(), startOffsetBytes, &startOffset); err != nil {
		return ierrors.Wrap(err, "failed to decode start offset")
	}

	if b.startOffset != 0 && b.startOffset != SliceOffsetAuto {
		if startOffset != b.startOffset {
			return ierrors.Errorf("start offset %d does not match existing offset %d in file", b.startOffset, startOffset)
		}
	}

	b.startOffset = startOffset

	return nil
}

func (b *ByteSlice) writeHeader() (err error) {
	startOffsetBytes, err := serix.DefaultAPI.Encode(context.Background(), uint64(b.startOffset))
	if err != nil {
		return ierrors.Wrap(err, "failed to encode startOffset")
	}

	entrySizeBytes, err := serix.DefaultAPI.Encode(context.Background(), uint64(b.entrySize))
	if err != nil {
		return ierrors.Wrap(err, "failed to encode entrySize")
	}

	if _, err = b.fileHandle.WriteAt(startOffsetBytes, 0); err != nil {
		return ierrors.Wrap(err, "failed to write startOffset")
	} else if _, err = b.fileHandle.WriteAt(entrySizeBytes, 8); err != nil {
		return ierrors.Wrap(err, "failed to write entrySize")
	}

	return b.fileHandle.Sync()
}

func WithOffset(offset uint64) options.Option[ByteSlice] {
	return func(s *ByteSlice) {
		s.startOffset = offset
	}
}
