package marshalutil

import "encoding/binary"

// Uint16Size contains the amount of bytes of a marshaled uint16 value.
const Uint16Size = 2

// WriteUint16 writes a marshaled uint16 value to the internal buffer.
func (util *MarshalUtil) WriteUint16(value uint16) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Uint16Size)

	binary.LittleEndian.PutUint16(util.bytes[util.writeOffset:writeEndOffset], value)

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadUint16 reads an uint16 value from the internal buffer.
func (util *MarshalUtil) ReadUint16() (uint16, error) {
	readEndOffset, err := util.checkReadCapacity(Uint16Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return binary.LittleEndian.Uint16(util.bytes[util.readOffset:readEndOffset]), nil
}
