package safemath

import (
	"errors"
	"fmt"
)

var (
	// ErrIntegerOverflow gets returned if an operation on two integers would under- or overflow.
	ErrIntegerOverflow = errors.New("integer under- or overflow")
	// ErrIntegerDivisionByZero gets returned if an integer division would cause a division by zero error.
	ErrIntegerDivisionByZero = errors.New("integer division by zero")
)

type Integer interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~int64 | ~int32 | ~int16 | ~int8
}

// Returns an error if x + y would under- or overflow, x + y otherwise.
func SafeAdd[T Integer](x T, y T) (T, error) {
	result := x + y

	if y > 0 {
		if result < x {
			return 0, fmt.Errorf("%w: %d and %d", ErrIntegerOverflow, x, y)
		}
	} else if result > x {
		return 0, fmt.Errorf("%w: %d and %d", ErrIntegerOverflow, x, y)
	}

	return result, nil
}

// Returns an error if x - y would under- or overflow, x - y otherwise.
func SafeSub[T Integer](x T, y T) (T, error) {
	result := x - y

	if y > 0 {
		if result > x {
			return 0, fmt.Errorf("%w: %d and %d", ErrIntegerOverflow, x, y)
		}
	} else if result < x {
		return 0, fmt.Errorf("%w: %d and %d", ErrIntegerOverflow, x, y)
	}

	return result, nil
}

// Returns an error if x * y would under- or overflow, x * y otherwise.
func SafeMul[T Integer](x T, y T) (T, error) {
	// Implementation inspired by:
	// https://github.com/OpenZeppelin/openzeppelin-contracts/blob/8cab922347e79732f6a532a75da5081ba7447a71/contracts/utils/math/Math.sol#L51
	result := x * y

	if x != 0 && result/x != y {
		return 0, fmt.Errorf("%w: %d and %d", ErrIntegerOverflow, x, y)
	}

	return result, nil
}

// Returns an error if y is zero and would cause a division by zero, x / y otherwise.
func SafeDiv[T Integer](x T, y T) (T, error) {
	if y == 0 {
		return 0, fmt.Errorf("%w: divisor is zero", ErrIntegerDivisionByZero)
	}

	return x / y, nil
}
