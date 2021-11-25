package marshalutil

import (
	"github.com/iotaledger/hive.go/byteutils"
)

// WriteBytes appends the given bytes to the internal buffer.
// It returns the same MarshalUtil so calls can be chained.
func (util *MarshalUtil) WriteBytes(bytes []byte) *MarshalUtil {
	if bytes == nil {
		return util
	}

	writeEndOffset := util.expandWriteCapacity(len(bytes))

	copy(util.bytes[util.writeOffset:writeEndOffset], bytes)

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadBytes unmarshals the given amount of bytes from the internal read buffer and advances the read offset. If an
// optionalReadOffset parameter is provided, then the method does not modify the read offset but instead just returns a
// copy of the bytes in the provided range.
func (util *MarshalUtil) ReadBytes(length int, optionalReadOffset ...int) ([]byte, error) {
	// temporarily modify read offset if optional offset is provided
	if len(optionalReadOffset) != 0 {
		defer util.ReadSeek(util.readOffset)
		util.ReadSeek(optionalReadOffset[0])
	}

	// determine the length
	if length < 0 {
		length = len(util.bytes) - util.readOffset + length
	}

	// calculate the end offset
	readEndOffset, err := util.checkReadCapacity(length)
	if err != nil {
		return nil, err
	}

	// return a copy of the byte range if a manual offset was provided
	if len(optionalReadOffset) != 0 {
		return byteutils.ConcatBytes(util.bytes[util.readOffset:readEndOffset]), nil
	}

	// advance read offset and return read bytes
	defer util.ReadSeek(readEndOffset)
	return util.bytes[util.readOffset:readEndOffset], nil
}

// ReadRemainingBytes reads the remaining bytes from the internal buffer.
func (util *MarshalUtil) ReadRemainingBytes() []byte {
	defer util.ReadSeek(util.size)

	return util.bytes[util.readOffset:]
}
