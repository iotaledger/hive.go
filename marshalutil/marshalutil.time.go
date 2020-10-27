package marshalutil

import (
	"time"
)

// TimeSize contains the amount of bytes of a marshaled Time value.
const TimeSize = Int64Size

// WriteTime writes a marshaled Time value to the internal buffer.
func (util *MarshalUtil) WriteTime(timeToWrite time.Time) *MarshalUtil {
	nanoSeconds := timeToWrite.UnixNano()

	// the zero value of time translates to -6795364578871345152
	if nanoSeconds == -6795364578871345152 {
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
