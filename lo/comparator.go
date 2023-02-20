package lo

import (
	"github.com/iotaledger/hive.go/constraints"
)

// Comparator is a generic comparator for two values. It returns 0 if the two values are equal, -1 if the first value is
// smaller and 1 if the first value is larger.
func Comparator[T constraints.Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
