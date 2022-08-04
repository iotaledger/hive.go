package marshalutil

import "encoding/binary"

// Int32Size contains the amount of bytes of a marshaled int32 value.
const Int32Size = 4

// WriteInt32 writes a marshaled int32 value to the internal buffer.
func (util *MarshalUtil) WriteInt32(value int32) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Int32Size)

	binary.LittleEndian.PutUint32(util.bytes[util.writeOffset:writeEndOffset], uint32(value))

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadInt32 reads an int32 value from the internal buffer.
func (util *MarshalUtil) ReadInt32() (int32, error) {
	readEndOffset, err := util.checkReadCapacity(Int32Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return int32(binary.LittleEndian.Uint32(util.bytes[util.readOffset:readEndOffset])), nil
}
