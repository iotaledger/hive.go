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

func ReadBytes(reader io.Reader, len int) ([]byte, error) {
	readBytes := make([]byte, len)

	nBytes, err := reader.Read(readBytes)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to read serialized bytes")
	}
	if nBytes != len {
		return nil, ierrors.Errorf("failed to read serialized bytes: read bytes (%d) != size (%d)", nBytes, len)
	}

	return readBytes, nil
}

// ReadBytesWithSize reads a byte slice from the reader where lenType specifies the serialization length prefix type.
func ReadBytesWithSize(reader io.Reader, lenType serializer.SeriLengthPrefixType) ([]byte, error) {
	size, err := readFixedSize(reader, lenType)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to read blob size")
	}

	return ReadBytes(reader, size)
}

// ReadObject reads a type from the reader as specified by objectFromBytesFunc. A fixed length for the deserialized type must be specified.
func ReadObject[T any](reader io.Reader, fixedLen int, objectFromBytesFunc func(bytes []byte) (T, int, error)) (T, error) {
	var result T

	readBytes, err := ReadBytes(reader, fixedLen)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to read serialized bytes")
	}

	result, consumedBytes, err := objectFromBytesFunc(readBytes)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to parse bytes of objectFromBytesFunc")
	}
	if consumedBytes != len(readBytes) {
		return result, ierrors.Errorf("failed to parse objectFromBytesFunc: consumed bytes (%d) != read bytes (%d)", consumedBytes, len(readBytes))
	}

	return result, nil
}

// ReadObjectWithSize reads a type from the reader as specified by fromBytesFunc. The serialization length prefix type must be specified.
func ReadObjectWithSize[T any](reader io.Reader, lenType serializer.SeriLengthPrefixType, objectFromBytesFunc func(bytes []byte) (T, int, error)) (T, error) {
	var result T

	readBytes, err := ReadBytesWithSize(reader, lenType)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to read serialized bytes")
	}

	result, consumedBytes, err := objectFromBytesFunc(readBytes)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to parse bytes of objectFromBytesFunc")
	}
	if consumedBytes != len(readBytes) {
		return result, ierrors.Errorf("failed to parse objectFromBytesFunc: consumed bytes (%d) != read bytes (%d)", consumedBytes, len(readBytes))
	}

	return result, nil
}

func ReadObjectFromReader[T any](reader io.Reader, objectFromReaderFunc func(reader io.Reader) (T, error)) (T, error) {
	var result T

	result, err := objectFromReaderFunc(reader)
	if err != nil {
		return result, ierrors.Wrap(err, "failed to read object from reader")
	}

	return result, nil
}

func PeekCollectionSize(reader io.ReadSeeker, lenType serializer.SeriLengthPrefixType) (int, error) {
	startOffset, err := Offset(reader)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to get start offset")
	}

	elementsCount, err := readFixedSize(reader, lenType)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to read collection count")
	}

	if _, err = GoTo(reader, startOffset); err != nil {
		return 0, ierrors.Wrap(err, "failed to go back to start offset")
	}

	return elementsCount, nil
}

// ReadCollection reads a collection from the reader where lenType specifies the serialization length prefix type.
func ReadCollection(reader io.Reader, lenType serializer.SeriLengthPrefixType, readCallback func(int) error) error {
	elementsCount, err := readFixedSize(reader, lenType)
	if err != nil {
		return ierrors.Wrap(err, "failed to read collection count")
	}

	for i := 0; i < elementsCount; i++ {
		if err = readCallback(i); err != nil {
			return ierrors.Wrapf(err, "failed to read element %d", i)
		}
	}

	return nil
}

func readFixedSize(reader io.Reader, lenType serializer.SeriLengthPrefixType) (int, error) {
	switch lenType {
	case serializer.SeriLengthPrefixTypeAsByte:
		result, err := Read[uint8](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return int(result), nil
	case serializer.SeriLengthPrefixTypeAsUint16:
		result, err := Read[uint16](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return int(result), nil
	case serializer.SeriLengthPrefixTypeAsUint32:
		result, err := Read[uint32](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return int(result), nil
	case serializer.SeriLengthPrefixTypeAsUint64:
		result, err := Read[uint64](reader)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to read length prefix")
		}

		return int(result), nil
	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
}
