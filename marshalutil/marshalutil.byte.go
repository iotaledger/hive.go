package marshalutil

func (util *MarshalUtil) WriteByte(byte byte) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(1)

	util.bytes[util.writeOffset] = byte

	util.WriteSeek(writeEndOffset)

	return util
}

func (util *MarshalUtil) ReadByte() (byte, error) {
	readEndOffset, err := util.checkReadCapacity(1)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return util.bytes[util.readOffset], nil
}
