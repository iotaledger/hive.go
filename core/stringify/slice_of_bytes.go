package stringify

import "encoding/hex"

func SliceOfBytes(value []byte) string {
	if value == nil {
		return "<nil>"
	} else if len(value) == 0 {
		return "<empty>"
	} else {
		return "0x" + hex.EncodeToString(value) + ""
	}
}
