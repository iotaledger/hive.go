package genericcomparator

import (
	"bytes"
	"fmt"
	"strings"
)

// Type defines the type of the generic Comparator that compares two values and returns -1 if a is smaller than b, 1 if
// a is bigger than b and 0 if both values are equal.
type Type func(a interface{}, b interface{}) int

// Comparator implements a function that compares the builtin basic types of Go and that returns -1 if a is smaller than
// b, 1 if a is bigger than b and 0 if both values are equal.
func Comparator(a, b interface{}) int {
	switch aCasted := a.(type) {
	case string:
		return strings.Compare(aCasted, b.(string))
	case []byte:
		return bytes.Compare(aCasted, b.([]byte))
	case int:
		bCasted := b.(int)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case uint:
		bCasted := b.(uint)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case int8:
		bCasted := b.(int8)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case uint8:
		bCasted := b.(uint8)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case int16:
		bCasted := b.(int16)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case uint16:
		bCasted := b.(uint16)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case int32:
		bCasted := b.(int32)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case uint32:
		bCasted := b.(uint32)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case int64:
		bCasted := b.(int64)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case uint64:
		bCasted := b.(uint64)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case float32:
		bCasted := b.(float32)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	case float64:
		bCasted := b.(float64)
		switch {
		case aCasted < bCasted:
			return -1
		case aCasted > bCasted:
			return 1
		}
	default:
		panic(fmt.Sprintf("unsupported key type: %v", a))
	}

	return 0
}
