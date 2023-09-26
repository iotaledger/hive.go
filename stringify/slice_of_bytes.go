package stringify

import "encoding/hex"

func SliceOfBytes(value []byte) string {
	switch {
	case value == nil:
		return "<nil>"

	case len(value) == 0:
		return "<empty>"

	default:
		return "0x" + hex.EncodeToString(value) + ""
	}
}
