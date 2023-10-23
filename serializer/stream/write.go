package stream

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// Write writes one of the allowedGenericTypes basic type to the writer.
func Write[T allowedGenericTypes](writer io.Writer, value T) {
	if err := binary.Write(writer, binary.LittleEndian, value); err != nil {
		// This should never happen as we cover only basic types.
		panic(err)
	}
}

// WriteBytesVariable writes bytes to the writer where lenType specifies the serialization length prefix type.
func WriteBytesVariable(writer io.Writer, bytes []byte, lenType serializer.SeriLengthPrefixType) error {
	if err := writeFixedSize(writer, len(bytes), lenType); err != nil {
		return ierrors.Wrap(err, "failed to write bytes length")
	}

	if _, err := writer.Write(bytes); err != nil {
		return ierrors.Wrap(err, "failed to write bytes")
	}

	return nil
}

// WriteFixedFunc writes a type to the writer as specified by the toBytesFunc. A fixed length for the serialized type must be specified.
func WriteFixedFunc[T any](writer io.Writer, target T, fixedLen int, toBytesFunc func(T) ([]byte, error)) error {
	serializedBytes, err := toBytesFunc(target)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize target")
	}

	if fixedLen != len(serializedBytes) {
		return ierrors.Errorf("serialized bytes length (%d) != fixed size (%d)", len(serializedBytes), fixedLen)
	}

	if _, err = writer.Write(serializedBytes); err != nil {
		return ierrors.Wrap(err, "failed to write target")
	}

	return nil
}

// WriteVariableFunc writes a type to the writer as specified by the writeFunc. The serialization length prefix type must be specified.
func WriteVariableFunc[T any](writer io.Writer, target T, lenType serializer.SeriLengthPrefixType, writeFunc func(T) ([]byte, error)) error {
	serializedBytes, err := writeFunc(target)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize target")
	}

	if err = WriteBytesVariable(writer, serializedBytes, lenType); err != nil {
		return ierrors.Wrap(err, "failed to write serialized bytes")
	}

	return nil
}

// WriteCollection writes a collection to the writer where lenType specifies the serialization length prefix type.
func WriteCollection(writer io.WriteSeeker, lenType serializer.SeriLengthPrefixType, writeCollection func() (elementsCount int, err error)) error {
	var elementsCount int
	var startOffset, endOffset int64
	var err error

	if startOffset, err = Offset(writer); err != nil {
		return ierrors.Wrap(err, "failed to get start offset")
	}

	if err = writeFixedSize(writer, 0, lenType); err != nil {
		return ierrors.Wrap(err, "failed to skip elements count")
	}

	if elementsCount, err = writeCollection(); err != nil {
		return ierrors.Wrap(err, "failed to write collection")
	}

	if endOffset, err = Offset(writer); err != nil {
		return ierrors.Wrap(err, "failed to read end offset of collection")
	}

	if _, err = GoTo(writer, startOffset); err != nil {
		return ierrors.Wrap(err, "failed to seek to start of attestors")
	}

	if err = writeFixedSize(writer, elementsCount, lenType); err != nil {
		return ierrors.Wrap(err, "failed to write attestors count")
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
		Write(writer, uint8(l))

		return nil

	case serializer.SeriLengthPrefixTypeAsUint16:
		if l > math.MaxUint16 {
			return ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint16)
		}
		Write(writer, uint16(l))

		return nil

	case serializer.SeriLengthPrefixTypeAsUint32:
		if l > math.MaxUint32 {
			return ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint32)
		}
		Write(writer, uint32(l))

		return nil

	case serializer.SeriLengthPrefixTypeAsUint64:
		Write(writer, uint64(l))

		return nil

	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
}
