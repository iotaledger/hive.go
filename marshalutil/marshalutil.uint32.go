package marshalutil

import (
	"encoding/binary"
)

const UINT32_SIZE = 4

func (util *MarshalUtil) WriteUint32(value uint32) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(UINT32_SIZE)

	binary.LittleEndian.PutUint32(util.bytes[util.writeOffset:writeEndOffset], value)

	util.WriteSeek(writeEndOffset)

	return util
}

func (util *MarshalUtil) ReadUint32() (uint32, error) {
	readEndOffset, err := util.checkReadCapacity(UINT32_SIZE)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return binary.LittleEndian.Uint32(util.bytes[util.readOffset:readEndOffset]), nil
}
