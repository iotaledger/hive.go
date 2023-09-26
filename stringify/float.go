package stringify

import "strconv"

func Float64(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}

func Float32(f float32) string {
	return strconv.FormatFloat(float64(f), 'E', -1, 32)
}
