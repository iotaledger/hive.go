package marshalutil

import (
	"encoding/binary"
)

// Int64Size contains the amount of bytes of a marshaled int64 value.
const Int64Size = 8

// WriteInt64 writes a marshaled int64 value to the internal buffer.
func (util *MarshalUtil) WriteInt64(value int64) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Int64Size)

	binary.LittleEndian.PutUint64(util.bytes[util.writeOffset:writeEndOffset], uint64(value))

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadInt64 reads an int64 value from the internal buffer.
func (util *MarshalUtil) ReadInt64() (int64, error) {
	readEndOffset, err := util.checkReadCapacity(Int64Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return int64(binary.LittleEndian.Uint64(util.bytes[util.readOffset:readEndOffset])), nil
}
