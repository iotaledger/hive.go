package constraints

import "fmt"

// Signed is a constraint that permits any signed integer type.
// If future releases of Go add new predeclared signed integer types,
// this constraint will be modified to include them.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint that permits any integer type.
// If future releases of Go add new predeclared integer types,
// this constraint will be modified to include them.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint that permits any floating-point type.
// If future releases of Go add new predeclared floating-point types,
// this constraint will be modified to include them.
type Float interface {
	~float32 | ~float64
}

// Complex is a constraint that permits any complex numeric type.
// If future releases of Go add new predeclared complex numeric types,
// this constraint will be modified to include them.
type Complex interface {
	~complex64 | ~complex128
}

// Numeric is a constraint that permits any numeric type: any type
// that supports the numeric operators.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type Numeric interface {
	Integer | Float
}

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type Ordered interface {
	Integer | Float | ~string
}

type Comparable[T any] interface {
	Compare(other T) int
}

// ComparableStringer is a constraint that returns a comparable type via Key()
// and a string representation via String().
type ComparableStringer[K comparable] interface {
	Key() K
	fmt.Stringer
}

// Cloneable is a constraint that permits cloning of any object.
type Cloneable[T any] interface {
	// Clone returns an exact copy of the object.
	Clone() T
}

// Equalable is a constraint that permits checking for equality of any object.
type Equalable[T any] interface {
	Equal(other T) bool
}
