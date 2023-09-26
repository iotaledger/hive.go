package stringify

import (
	"strconv"
)

func UInt(value uint64) string {
	return strconv.FormatUint(value, 10)
}
