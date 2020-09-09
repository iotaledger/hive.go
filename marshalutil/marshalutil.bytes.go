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

// ReadBytes unmarshals the given amount of bytes from the internal read buffer and advances the read offset.
func (util *MarshalUtil) ReadBytes(length int, optionalReadOffset ...int) ([]byte, error) {
	// determine the read offset
	readOffset := util.readOffset
	if len(optionalReadOffset) >= 1 {
		readOffset = optionalReadOffset[0]
	}

	// determine the length
	if length < 0 {
		length = len(util.bytes) - readOffset + length
	}

	// calculate the end offset
	readEndOffset, err := util.checkReadCapacity(length)
	if err != nil {
		return nil, err
	}

	// return a copy of the byte range if a manual offset was provided
	if len(optionalReadOffset) != 0 {
		return byteutils.ConcatBytes(util.bytes[readOffset:readEndOffset]), nil
	}

	// advance read offset and return read bytes
	util.ReadSeek(readEndOffset)
	return util.bytes[readOffset:readEndOffset], nil
}

func (util *MarshalUtil) ReadRemainingBytes() []byte {
	defer util.ReadSeek(util.size)

	return util.bytes[util.readOffset:]
}
