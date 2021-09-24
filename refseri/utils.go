package refseri

import (
	"bytes"
	"fmt"
	"go/types"
	"sort"

	"github.com/iotaledger/hive.go/marshalutil"
)

// ReadLen reads length of a collection from the buffer according to lenPrefixType
func ReadLen(lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) (int, error) {
	switch lenPrefixType {
	case types.Uint8:
		lengthUint8, err := buffer.ReadUint8()
		return int(lengthUint8), err
	case types.Uint16:
		lengthUint16, err := buffer.ReadUint16()
		return int(lengthUint16), err
	case types.Uint32:
		lengthUint32, err := buffer.ReadUint32()
		return int(lengthUint32), err
	default:
		return 0, fmt.Errorf("%w: %d", ErrUnknownLengthPrefix, lenPrefixType)
	}
}

// WriteLen writes length of a collection from the buffer according to lenPrefixType
func WriteLen(length int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
	switch lenPrefixType {
	case types.Uint8:
		buffer.WriteUint8(uint8(length))
	case types.Uint16:
		buffer.WriteUint16(uint16(length))
	case types.Uint32:
		buffer.WriteUint32(uint32(length))
	default:
		return fmt.Errorf("%w: %d", ErrUnknownLengthPrefix, lenPrefixType)
	}
	return nil
}

// ValidateLength is used to make sure that the length of a collection is within bounds specified in struct tags.
func ValidateLength(length int, minSliceLen int, maxSliceLen int) (err error) {
	if length < minSliceLen {
		err = fmt.Errorf("%w: min %d elements instead of %d", ErrSliceMinLength, minSliceLen, length)
		return
	}
	if maxSliceLen > 0 && length > maxSliceLen {
		err = fmt.Errorf("%w: max %d elements instead of %d", ErrSliceMaxLength, maxSliceLen, length)
		return
	}
	return
}

// InsertSorted inserts new byte slice into a sorted slice
func InsertSorted(ss [][]byte, s []byte) [][]byte {
	i := sort.Search(len(ss), func(i int) bool {
		return bytes.Compare(s, ss[i]) < 0
	})
	ss = append(ss, []byte{})
	copy(ss[i+1:], ss[i:])
	ss[i] = s
	return ss
}
