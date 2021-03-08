package marshalutil

import (
	"encoding/binary"
)

// Int16Size contains the amount of bytes of a marshaled int16 value.
const Int16Size = 2

// WriteInt16 writes a marshaled int16 value to the internal buffer.
func (util *MarshalUtil) WriteInt16(value int16) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Int16Size)

	binary.LittleEndian.PutUint16(util.bytes[util.writeOffset:writeEndOffset], uint16(value))

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadInt16 reads an int16 value from the internal buffer.
func (util *MarshalUtil) ReadInt16() (int16, error) {
	readEndOffset, err := util.checkReadCapacity(Int16Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return int16(binary.LittleEndian.Uint16(util.bytes[util.readOffset:readEndOffset])), nil
}
