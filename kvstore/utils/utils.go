package utils

import (
	"sort"

	"github.com/iotaledger/hive.go/kvstore"
)

// CopyBytes returns a copy of the source slice.
// If size is not passed, the result slice has same size as the source slice.
func CopyBytes(source []byte, size ...int) []byte {
	targetSize := len(source)
	if len(size) > 0 {
		targetSize = size[0]
	}
	cpy := make([]byte, targetSize)
	copy(cpy, source)

	return cpy
}

// KeyPrefixUpperBound returns the upper bound (not included in the prefix set)
// for a prefix range scan.
func KeyPrefixUpperBound(start []byte) []byte {
	end := make([]byte, len(start))
	copy(end, start)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}

	return nil // no upper-bound
}

// SortSlice sorts a slice according to the given IterDirection.
func SortSlice(slice []string, iterDirection ...kvstore.IterDirection) []string {
	switch kvstore.GetIterDirection(iterDirection...) {
	case kvstore.IterDirectionForward:
		sort.Sort(sort.StringSlice(slice))

	case kvstore.IterDirectionBackward:
		sort.Sort(sort.Reverse(sort.StringSlice(slice)))
	}

	return slice
}
