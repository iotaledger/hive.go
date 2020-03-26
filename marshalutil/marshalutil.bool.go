package marshalutil

const BOOL_SIZE = 1

func (util *MarshalUtil) WriteBool(bool bool) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(1)

	if bool {
		util.bytes[util.writeOffset] = 1
	} else {
		util.bytes[util.writeOffset] = 0
	}

	util.WriteSeek(writeEndOffset)

	return util
}

func (util *MarshalUtil) ReadBool() (bool, error) {
	readEndOffset, err := util.checkReadCapacity(1)
	if err != nil {
		return false, err
	}

	defer util.ReadSeek(readEndOffset)

	return util.bytes[util.readOffset] == 1, nil
}
