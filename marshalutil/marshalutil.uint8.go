package marshalutil

// Uint8Size contains the amount of bytes of a marshaled uint8 value.
const Uint8Size = 1

// WriteUint8 writes a marshaled uint8 value to the internal buffer.
func (util *MarshalUtil) WriteUint8(value uint8) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Uint8Size)

	util.bytes[util.writeOffset] = value

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadUint8 reads an uint8 value from the internal buffer.
func (util *MarshalUtil) ReadUint8() (uint8, error) {
	readEndOffset, err := util.checkReadCapacity(Uint8Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return util.bytes[util.readOffset], nil
}
