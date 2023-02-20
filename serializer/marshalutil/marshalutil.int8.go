package marshalutil

// Int8Size contains the amount of bytes of a marshaled int8 value.
const Int8Size = 1

// WriteInt8 writes a marshaled int8 value to the internal buffer.
func (util *MarshalUtil) WriteInt8(value int8) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Int8Size)

	util.bytes[util.writeOffset] = byte(value)

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadInt8 reads an int8 value from the internal buffer.
func (util *MarshalUtil) ReadInt8() (int8, error) {
	readEndOffset, err := util.checkReadCapacity(Int8Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return int8(util.bytes[util.readOffset]), nil
}
