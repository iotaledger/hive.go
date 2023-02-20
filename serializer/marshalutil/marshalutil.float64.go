package marshalutil

import (
	"encoding/binary"
	"math"
)

// Float64Size contains the amount of bytes of a marshaled float64 value.
const Float64Size = 8

// WriteFloat64 writes a marshaled float64 value to the internal buffer.
func (util *MarshalUtil) WriteFloat64(value float64) *MarshalUtil {
	writeEndOffset := util.expandWriteCapacity(Float64Size)

	binary.LittleEndian.PutUint64(util.bytes[util.writeOffset:writeEndOffset], math.Float64bits(value))

	util.WriteSeek(writeEndOffset)

	return util
}

// ReadFloat64 reads a float64 value from the internal buffer.
func (util *MarshalUtil) ReadFloat64() (float64, error) {
	readEndOffset, err := util.checkReadCapacity(Float64Size)
	if err != nil {
		return 0, err
	}

	defer util.ReadSeek(readEndOffset)

	return math.Float64frombits(binary.LittleEndian.Uint64(util.bytes[util.readOffset:readEndOffset])), nil
}
