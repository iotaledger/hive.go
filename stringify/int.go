package stringify

import (
	"strconv"
)

func Int(value int64) string {
	return strconv.FormatInt(value, 10)
}
