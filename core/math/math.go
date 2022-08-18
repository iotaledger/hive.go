package math

import (
	"math"
)

func Abs(n int64) int64 {
	y := n >> 63

	return (n ^ y) - y
}

// Uint32Diff returns the difference between newCount and oldCount
// and catches overflows.
func Uint32Diff(newCount uint32, oldCount uint32) uint32 {
	// Catch overflows
	if newCount < oldCount {
		return (math.MaxUint32 - oldCount) + newCount
	}

	return newCount - oldCount
}
