package marshalutil

import (
	"encoding/binary"
	"math"
)

// Float32Size contains the amount of bytes of a marshaled float32 value.
const Float32Size = 4

// WriteFloat32 writes a marshaled float64 value to the internal buffer.
func (util *MarshalUtil) WriteFloat32(value float32) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Float64Size)

	binary.LittleEndian.PutUint32(util.bytes[util.writeOffset:writeEndOffset], math.Float32bits(value))

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadFloat32 reads a float32 value from the internal buffer.
func (util *MarshalUtil) ReadFloat32() (float32, error) {
	readEndOffset, err := util.checkReadCapacity(Float64Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return math.Float32frombits(binary.LittleEndian.Uint32(util.bytes[util.readOffset:readEndOffset])), nil
}
