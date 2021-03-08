package marshalutil

import (
	"encoding/binary"
)

// Uint32Size contains the amount of bytes of a marshaled uint32 value.
const Uint32Size = 4

// WriteUint32 writes a marshaled uint32 value to the internal buffer.
func (util *MarshalUtil) WriteUint32(value uint32) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Uint32Size)

	binary.LittleEndian.PutUint32(util.bytes[util.writeOffset:writeEndOffset], value)

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadUint32 reads an uint32 value from the internal buffer.
func (util *MarshalUtil) ReadUint32() (uint32, error) {
	readEndOffset, err := util.checkReadCapacity(Uint32Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return binary.LittleEndian.Uint32(util.bytes[util.readOffset:readEndOffset]), nil
}
