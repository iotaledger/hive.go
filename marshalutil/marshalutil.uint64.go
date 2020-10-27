package marshalutil

import "encoding/binary"

const Uint64Size = 8

func (util *MarshalUtil) WriteUint64(value uint64) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Uint64Size)

	binary.LittleEndian.PutUint64(util.bytes[util.writeOffset:writeEndOffset], value)

	util.WriteSeek(writeEndOffset)

	return util
}

func (util *MarshalUtil) ReadUint64() (uint64, error) {
	readEndOffset, err := util.checkReadCapacity(Uint64Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return binary.LittleEndian.Uint64(util.bytes[util.readOffset:readEndOffset]), nil
}
