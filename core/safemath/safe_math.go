package safemath

import (
	"math/bits"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrIntegerOverflow gets returned if an operation on two integers would under- or overflow.
	ErrIntegerOverflow = ierrors.New("integer under- or overflow")
	// ErrIntegerDivisionByZero gets returned if an integer division would cause a division by zero error.
	ErrIntegerDivisionByZero = ierrors.New("integer division by zero")
)

type Integer interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~int64 | ~int32 | ~int16 | ~int8
}

// Returns x + y or an error if that computation would under- or overflow.
func SafeAdd[T Integer](x T, y T) (T, error) {
	result := x + y

	if y > 0 {
		if result < x {
			return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
		}
	} else if result > x {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	return result, nil
}

// Returns x - y or an error if that computation would under- or overflow.
func SafeSub[T Integer](x T, y T) (T, error) {
	result := x - y

	if y > 0 {
		if result > x {
			return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
		}
	} else if result < x {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	return result, nil
}

// Returns x * y or an error if that computation would overflow.
//
// According to benchmarks, this function is about 73% faster than SafeMul.
func SafeMulUint64(x, y uint64) (uint64, error) {
	hi, lo := bits.Mul64(x, y)

	if hi != 0 {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	return lo, nil
}

// Returns x * y or an error if that computation would under- or overflow.
//
// According to benchmarks, this function is about 27% faster than SafeMul.
func SafeMulInt64(x, y int64) (int64, error) {
	// This function stores the sign of the resulting int64 multiplication
	// and then executes the multiplication with two uint64s, in 128-bit space.
	// If any of the upper 64-bits are non-zero, an overflow has occurred.
	// If the result should be positive, but the sign bit is set, an overflow has occurred.
	// If the result should be negative, but the sign bit is unset, an overflow has occurred.
	//
	// The following examples use uint16/int8 as a more easily understandable example.
	//
	// *Positive Overflow*
	//
	// As a positive overflow example, take int8(66) * 2.
	// The result as a uint16 will be 132, or in binary: 0000 0000 1000 0100.
	// Interpreted as an int8 the high part is cut off and so the sign bit (most significant bit) is set,
	// but we expected the result to be positive, so we detect this as an overflow.
	//
	// *Negative Overflow*
	//
	// As a negative overflow example, take int8(-66) * 2.
	// The result as a uint16 will 132, or in binary 0000 0000 1000 0100, since we calculate 66*2 without the sign.
	// Interpreted as an int8 the high part is cut off and the result multiplied with -1, to account for the
	// expected sign of the result, so we get 0111 1100.
	// The sign bit (most significant bit) is unset, but we expected the result to be negative,
	// so we detect this as an overflow.
	//
	// *Overflowing 8-bit space*
	//
	// As a larger negative overflow example using 8-bit integers, take int8(-128) * 32.
	// The multiplication would take place using 16-bits, so the result as a uint16 will be
	// binary 0001 0000 0000 0000, since we calculate 128*32 without the sign.
	// Interpreted as an int8 and multiplied with -1, to account for the expected sign of the result, we simply get
	// 0000 0000, since the least significant byte of the result is 0.
	// However, since the most significant byte of the result is non-zero, we detect the overflow.

	xNegative := x < 0
	yNegative := y < 0
	xPositive := x > 0
	yPositive := y > 0
	resultIsPositive := true
	var resultSign int64 = 1

	// Store the sign of the resulting computation and negate any negative integers,
	// so we can correctly convert to uint64.
	if xNegative {
		if yPositive {
			resultIsPositive = false
			resultSign = -1
		}
		x = -x
	}

	if yNegative {
		if xPositive {
			resultIsPositive = false
			resultSign = -1
		}
		y = -y
	}

	// Execute the multiplication in 128-bit space using unsigned integers.
	hi, lo := bits.Mul64(uint64(x), uint64(y))
	// Interpret the lo result as an int64, then correct for the expected sign.
	loSigned := int64(lo) * resultSign

	// If the computation overflowed a uint64, it would also overflow an int64.
	if hi != 0 {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	// Extract the most significant bit, signaling if the number is negative (1) or not (0).
	signBitSet := ((loSigned >> 63) & 1) == 1

	// If the result is expected to be positive but the sign bit is set, it's an overflow.
	if resultIsPositive && signBitSet {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	// If the result is expected to be negative but the sign bit is not set, it's an underflow.
	if !resultIsPositive && !signBitSet {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	return loSigned, nil
}

// Returns x * y or an error if that computation would under- or overflow.
func SafeMul[T Integer](x T, y T) (T, error) {
	result := x * y

	if x != 0 && result/x != y {
		return 0, ierrors.Wrapf(ErrIntegerOverflow, "%d and %d", x, y)
	}

	return result, nil
}

// Returns x / y or an error if that computation would cause a division by zero.
func SafeDiv[T Integer](x T, y T) (T, error) {
	if y == 0 {
		return 0, ierrors.Wrapf(ErrIntegerDivisionByZero, "%d and %d", x, y)
	}

	return x / y, nil
}
