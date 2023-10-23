package stream

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// Read reads a generic basic type from the reader.
func Read[T allowedGenericTypes](reader io.Reader) (result T, err error) {
	return result, binary.Read(reader, binary.LittleEndian, &result)
}

// ReadBytesVariable reads a byte slice from the reader where lenType specifies the serialization length prefix type.
func ReadBytesVariable(reader io.Reader, lenType serializer.SeriLengthPrefixType) ([]byte, error) {
	size, err := readFixedSize(reader, lenType)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to read blob size")
	}

	readBytes := make([]byte, size)

	nBytes, err := reader.Read(readBytes)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to read serialized bytes")
	}
	if uint64(nBytes) != size {
		return nil, ierrors.Errorf("failed to read serialized bytes: read bytes (%d) != size (%d)", nBytes, size)
	}

	return readBytes, nil
}

// ReadFixedFunc reads a type from the reader as specified by fromBytesFunc. A fixed length for the deserialized type must be specified.
func ReadFixedFunc[T any](reader io.Reader, fixedLen int, fromBytesFunc func(bytes []byte) (T, int, error)) (T, error) {
	var result T
	readBytes := make([]byte, fixedLen)

	nBytes, err := reader.Read(readBytes)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to read serialized bytes")
	}
	if nBytes != fixedLen {
		return result, ierrors.Errorf("failed to read serialized bytes: read bytes (%d) != fixed size (%d)", nBytes, fixedLen)
	}

	result, consumedBytes, err := fromBytesFunc(readBytes)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to parse bytes of fromBytesFunc")
	}
	if consumedBytes != len(readBytes) {
		return result, ierrors.Errorf("failed to parse fromBytesFunc: consumed bytes (%d) != read bytes (%d)", consumedBytes, len(readBytes))
	}

	return result, nil
}

// ReadVariableFunc reads a type from the reader as specified by fromBytesFunc. The serialization length prefix type must be specified.
func ReadVariableFunc[T any](reader io.Reader, lenType serializer.SeriLengthPrefixType, fromBytesFunc func(bytes []byte) (T, int, error)) (T, error) {
	var result T

	readBytes, err := ReadBytesVariable(reader, lenType)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to read serialized bytes")
	}

	result, consumedBytes, err := fromBytesFunc(readBytes)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to parse bytes of fromBytesFunc")
	}
	if consumedBytes != len(readBytes) {
		return result, ierrors.Errorf("failed to parse fromBytesFunc: consumed bytes (%d) != read bytes (%d)", consumedBytes, len(readBytes))
	}

	return result, nil
}

// ReadCollection reads a collection from the reader where lenType specifies the serialization length prefix type.
func ReadCollection(reader io.Reader, lenType serializer.SeriLengthPrefixType, readCallback func(uint64) error) (err error) {
	var elementsCount uint64

	if elementsCount, err = readFixedSize(reader, lenType); err != nil {
		return ierrors.Wrap(err, "failed to read collection count")
	}

	for i := uint64(0); i < elementsCount; i++ {
		if err = readCallback(i); err != nil {
			return ierrors.Wrapf(err, "failed to read element %d", i)
		}
	}

	return nil
}

func readFixedSize(reader io.Reader, lenType serializer.SeriLengthPrefixType) (uint64, error) {
	switch lenType {
	case serializer.SeriLengthPrefixTypeAsByte:
		result, err := Read[uint8](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return uint64(result), nil
	case serializer.SeriLengthPrefixTypeAsUint16:
		result, err := Read[uint16](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return uint64(result), nil
	case serializer.SeriLengthPrefixTypeAsUint32:
		result, err := Read[uint32](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return uint64(result), nil
	case serializer.SeriLengthPrefixTypeAsUint64:
		result, err := Read[uint64](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return result, nil
	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
}
