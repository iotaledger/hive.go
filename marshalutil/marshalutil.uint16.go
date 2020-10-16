package marshalutil

import "encoding/binary"

const UINT16_SIZE = 2

func (util *MarshalUtil) WriteUint16(value uint16) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(UINT16_SIZE)

	binary.LittleEndian.PutUint16(util.bytes[util.writeOffset:writeEndOffset], value)

	util.WriteSeek(writeEndOffset)

	return util
}

func (util *MarshalUtil) ReadUint16() (uint16, error) {
	readEndOffset, err := util.checkReadCapacity(UINT16_SIZE)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return binary.LittleEndian.Uint16(util.bytes[util.readOffset:readEndOffset]), nil
}
