package stream

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// Write writes one of the allowedGenericTypes basic type to the writer.
func Write[T allowedGenericTypes](writer io.Writer, value T) error {
	return binary.Write(writer, binary.LittleEndian, value)
}

func WriteBytes(writer io.Writer, bytes []byte) error {
	return lo.Return2(writer.Write(bytes))
}

// WriteBytesWithSize writes bytes to the writer where lenType specifies the serialization length prefix type.
func WriteBytesWithSize(writer io.Writer, bytes []byte, lenType serializer.SeriLengthPrefixType) error {
	if err := writeFixedSize(writer, len(bytes), lenType); err != nil {
		return ierrors.Wrap(err, "failed to write bytes length")
	}

	if _, err := writer.Write(bytes); err != nil {
		return ierrors.Wrap(err, "failed to write bytes")
	}

	return nil
}

// WriteObject writes a type to the writer as specified by the objectToBytesFunc.
func WriteObject[T any](writer io.Writer, target T, objectToBytesFunc func(T) ([]byte, error)) error {
	serializedBytes, err := objectToBytesFunc(target)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize target")
	}

	if _, err = writer.Write(serializedBytes); err != nil {
		return ierrors.Wrap(err, "failed to write target")
	}

	return nil
}

// WriteObjectWithSize writes an object to the writer as specified by the objectToBytesFunc. The serialization length prefix type must be specified.
func WriteObjectWithSize[T any](writer io.Writer, target T, lenType serializer.SeriLengthPrefixType, objectToBytesFunc func(T) ([]byte, error)) error {
	serializedBytes, err := objectToBytesFunc(target)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize object")
	}

	if err = WriteBytesWithSize(writer, serializedBytes, lenType); err != nil {
		return ierrors.Wrap(err, "failed to write serialized bytes")
	}

	return nil
}

// WriteCollection writes a collection to the writer where lenType specifies the serialization length prefix type.
func WriteCollection(writer io.WriteSeeker, lenType serializer.SeriLengthPrefixType, writeCallback func() (elementsCount int, err error)) error {
	var elementsCount int
	var startOffset, endOffset int64
	var err error

	if startOffset, err = Offset(writer); err != nil {
		return ierrors.Wrap(err, "failed to get start offset")
	}

	if err = writeFixedSize(writer, 0, lenType); err != nil {
		return ierrors.Wrap(err, "failed to skip elements count")
	}

	if elementsCount, err = writeCallback(); err != nil {
		return ierrors.Wrap(err, "failed to write collection")
	}

	if endOffset, err = Offset(writer); err != nil {
		return ierrors.Wrap(err, "failed to read end offset of collection")
	}

	if _, err = GoTo(writer, startOffset); err != nil {
		return ierrors.Wrap(err, "failed to seek to start of attestors")
	}

	if err = writeFixedSize(writer, elementsCount, lenType); err != nil {
		return ierrors.Wrap(err, "failed to write elements count")
	}

	if _, err = GoTo(writer, endOffset); err != nil {
		return ierrors.Wrap(err, "failed to seek to end of attestors")
	}

	return nil
}

func writeFixedSize(writer io.Writer, l int, lenType serializer.SeriLengthPrefixType) error {
	switch lenType {
	case serializer.SeriLengthPrefixTypeAsByte:
		if l > math.MaxUint8 {
			return ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint8)
		}
		if err := Write(writer, uint8(l)); err != nil {
			return ierrors.Wrap(err, "unable to write length")
		}

		return nil

	case serializer.SeriLengthPrefixTypeAsUint16:
		if l > math.MaxUint16 {
			return ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint16)
		}
		if err := Write(writer, uint16(l)); err != nil {
			return ierrors.Wrap(err, "unable to write length")
		}

		return nil

	case serializer.SeriLengthPrefixTypeAsUint32:
		if l > math.MaxUint32 {
			return ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint32)
		}
		if err := Write(writer, uint32(l)); err != nil {
			return ierrors.Wrap(err, "unable to write length")
		}

		return nil

	case serializer.SeriLengthPrefixTypeAsUint64:
		if err := Write(writer, uint64(l)); err != nil {
			return ierrors.Wrap(err, "unable to write length")
		}

		return nil

	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
}
