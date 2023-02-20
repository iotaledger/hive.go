package marshalutil

import (
	"time"
)

// TimeSize contains the amount of bytes of a marshaled Time value.
const TimeSize = Int64Size

// WriteTime writes a marshaled Time value to the internal buffer.
func (util *MarshalUtil) WriteTime(timeToWrite time.Time) *MarshalUtil {
	if timeToWrite.IsZero() {
		util.WriteInt64(0)
	} else {
		util.WriteInt64(timeToWrite.UnixNano())
	}

	return util
}

// ReadTime reads a Time value from the internal buffer.
func (util *MarshalUtil) ReadTime() (result time.Time, err error) {
	nanoSeconds, err := util.ReadInt64()
	if err != nil {
		return
	}

	if nanoSeconds == 0 {
		result = time.Time{}
	} else {
		result = time.Unix(0, nanoSeconds)
	}

	return
}
