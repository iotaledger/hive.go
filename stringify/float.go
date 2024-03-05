package stringify

import "strconv"

func Float64(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func Float32(f float32) string {
	return strconv.FormatFloat(float64(f), 'g', -1, 32)
}
