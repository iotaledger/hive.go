package safemath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeMathAddUint8(t *testing.T) {
	var err error
	var ures uint8

	ures, err = SafeAdd(uint8(0), math.MaxUint8)
	require.NoError(t, err)
	require.Equal(t, ures, uint8(math.MaxUint8))

	ures, err = SafeAdd(math.MaxUint8, uint8(0))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(math.MaxUint8))

	ures, err = SafeAdd(uint8(100), uint8(100))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(200))

	// overflows
	_, err = SafeAdd(uint8(1), math.MaxUint8)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeAdd(math.MaxUint8, uint8(1))
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathAddInt8(t *testing.T) {
	var err error
	var ires int8

	// positive addends
	ires, err = SafeAdd(int8(0), math.MaxInt8)
	require.NoError(t, err)
	require.Equal(t, ires, int8(math.MaxInt8))

	ires, err = SafeAdd(math.MaxInt8, int8(0))
	require.NoError(t, err)
	require.Equal(t, ires, int8(math.MaxInt8))

	ires, err = SafeAdd(math.MinInt8, int8(math.MaxInt8))
	require.NoError(t, err)
	require.Equal(t, ires, int8(-1))

	ires, err = SafeAdd(int8(-100), int8(math.MaxInt8))
	require.NoError(t, err)
	require.Equal(t, ires, int8(27))

	ires, err = SafeAdd(int8(50), int8(50))
	require.NoError(t, err)
	require.Equal(t, ires, int8(100))

	// negative addends
	ires, err = SafeAdd(math.MaxInt8, int8(-math.MaxInt8))
	require.NoError(t, err)
	require.Equal(t, ires, int8(0))

	ires, err = SafeAdd(math.MaxInt8, int8(-math.MaxInt8))
	require.NoError(t, err)
	require.Equal(t, ires, int8(0))

	ires, err = SafeAdd(int8(100), int8(-50))
	require.NoError(t, err)
	require.Equal(t, ires, int8(50))

	// overflows
	_, err = SafeAdd(int8(1), math.MaxInt8)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeAdd(math.MaxInt8, int8(1))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	// underflows
	_, err = SafeAdd(int8(math.MinInt8), int8(-1))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeAdd(int8(-1), int8(math.MinInt8))
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathSubUint8(t *testing.T) {
	var err error
	var ures uint8

	ures, err = SafeSub(uint8(5), uint8(5))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(0))

	ures, err = SafeSub(uint8(5), uint8(4))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(1))

	_, err = SafeSub(uint8(4), uint8(5))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeSub(math.MaxUint8-uint8(1), math.MaxUint8)
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathSubInt8(t *testing.T) {
	var err error
	var ires int8

	// positive subtrahends
	ires, err = SafeSub(int8(5), int8(5))
	require.NoError(t, err)
	require.Equal(t, ires, int8(0))

	ires, err = SafeSub(int8(5), int8(4))
	require.NoError(t, err)
	require.Equal(t, ires, int8(1))

	ires, err = SafeSub(int8(4), int8(5))
	require.NoError(t, err)
	require.Equal(t, ires, int8(-1))

	// negative subtrahends
	ires, err = SafeSub(int8(-4), int8(-5))
	require.NoError(t, err)
	require.Equal(t, ires, int8(1))

	// underflows
	_, err = SafeSub(math.MinInt8, int8(1))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeSub(int8(-2), math.MaxInt8)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	// overflows
	_, err = SafeSub(math.MaxInt8, int8(-1))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeSub(int8(1), math.MinInt8)
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathMulUint8(t *testing.T) {
	var err error
	var ures uint8

	ures, err = SafeMul(uint8(5), uint8(5))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(25))

	_, err = SafeMul(uint8(100), uint8(3))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMul(math.MaxUint8, uint8(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMul(uint8(51), uint8(10))
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathMulInt8(t *testing.T) {
	var err error
	var ires int8

	ires, err = SafeMul(int8(5), int8(5))
	require.NoError(t, err)
	require.Equal(t, ires, int8(25))

	ires, err = SafeMul(int8(5), int8(-5))
	require.NoError(t, err)
	require.Equal(t, ires, int8(-25))

	_, err = SafeMul(int8(100), int8(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMul(math.MinInt8, int8(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMul(int8(math.MinInt8), math.MaxInt8)
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafeMathDiv(t *testing.T) {
	var err error
	var ures uint8

	ures, err = SafeDiv(uint8(10), uint8(2))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(5))

	ures, err = SafeDiv(uint8(10), uint8(8))
	require.NoError(t, err)
	require.Equal(t, ures, uint8(1))

	_, err = SafeDiv(int8(100), int8(0))
	require.ErrorIs(t, err, ErrIntegerDivisionByZero)
}

