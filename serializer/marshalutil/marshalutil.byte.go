package marshalutil

// WriteByte writes a marshaled byte value to the internal buffer.
func (util *MarshalUtil) WriteByte(b byte) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(1)

	util.bytes[util.writeOffset] = b

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadByte reads a byte value from the internal buffer.
func (util *MarshalUtil) ReadByte() (byte, error) {
	readEndOffset, err := util.checkReadCapacity(1)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return util.bytes[util.readOffset], nil
}
