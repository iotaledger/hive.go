package safemath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	benchmarkResult  uint64
	ibenchmarkResult int64
	// Pick large numbers that are not a multiple of two and which when multiplied, fit in an int64.
	testFactor1 = (1 << 16) - 1
	testFactor2 = (1 << 15) - 1
)

func BenchmarkMultiplicationSafeMul(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, _ := SafeMul(uint64(testFactor1), uint64(testFactor2))
		benchmarkResult = res
	}
}

func BenchmarkMultiplicationSafeMulUint64(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, _ := SafeMulUint64(uint64(testFactor1), uint64(testFactor2))
		benchmarkResult = res
	}
}

func BenchmarkMultiplicationSafeMulInt64(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, _ := SafeMulInt64(int64(testFactor1), int64(testFactor2))
		ibenchmarkResult = res
	}
}

// This simply benchmarks raw uint64 multiplication so we can check how much slower
// the safe functions are in comparison.
func BenchmarkMultiplication(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res := uint64(testFactor1) * uint64(testFactor2)
		benchmarkResult = res
	}
}

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

func TestSafeMathMulUint64(t *testing.T) {
	type safeMulFunc func(x, y uint64) (uint64, error)

	safeMulFuncs := []safeMulFunc{
		SafeMul[uint64],
		SafeMulUint64,
	}

	for _, fn := range safeMulFuncs {
		var err error
		var ures uint64

		ures, err = fn(uint64(5), uint64(5))
		require.NoError(t, err)
		require.Equal(t, uint64(25), ures)

		_, err = fn(math.MaxUint64, uint64(2))
		require.ErrorIs(t, err, ErrIntegerOverflow)

		_, err = fn(uint64(2), math.MaxUint64)
		require.ErrorIs(t, err, ErrIntegerOverflow)

		_, err = fn(uint64(math.MaxUint64), uint64(math.MaxUint64))
		require.ErrorIs(t, err, ErrIntegerOverflow)
	}
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

func TestSafeMathMulInt64(t *testing.T) {
	var err error
	var ires int64

	ires, err = SafeMulInt64(int64(-5), int64(0))
	require.NoError(t, err)
	require.Equal(t, int64(0), ires)

	ires, err = SafeMulInt64(int64(0), int64(-5))
	require.NoError(t, err)
	require.Equal(t, int64(0), ires)

	ires, err = SafeMulInt64(int64(-1), int64(-5))
	require.NoError(t, err)
	require.Equal(t, int64(5), ires)

	ires, err = SafeMulInt64(int64(-5), int64(1))
	require.NoError(t, err)
	require.Equal(t, int64(-5), ires)

	ires, err = SafeMulInt64(int64(5), int64(5))
	require.NoError(t, err)
	require.Equal(t, ires, int64(25))

	ires, err = SafeMulInt64(int64(5), int64(-5))
	require.NoError(t, err)
	require.Equal(t, int64(-25), ires)

	ires, err = SafeMulInt64(int64(-5), int64(-5))
	require.NoError(t, err)
	require.Equal(t, int64(25), ires)

	ires, err = SafeMulInt64(int64(5), int64(5))
	require.NoError(t, err)
	require.Equal(t, int64(25), ires)

	ires, err = SafeMulInt64(int64(math.MinInt64)/2, 2)
	require.NoError(t, err)
	require.Equal(t, int64(math.MinInt64), ires)

	ires, err = SafeMulInt64(int64(math.MaxInt64-1)/2, 2)
	require.NoError(t, err)
	require.Equal(t, int64(math.MaxInt64-1), ires)

	_, err = SafeMulInt64(math.MinInt64, int64(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMulInt64(math.MaxInt64, int64(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMulInt64(int64(math.MaxInt64), math.MaxInt64)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMulInt64(int64(math.MinInt64), math.MaxInt64)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	_, err = SafeMulInt64(int64(math.MinInt64), math.MinInt64)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	// The result of this computation would be 3 larger than MaxInt64.
	_, err = SafeMulInt64(int64(math.MaxInt64/2)+1, 2)
	require.ErrorIs(t, err, ErrIntegerOverflow)

	// The result of this computation would be 2 smaller than MinInt64.
	_, err = SafeMulInt64(int64(math.MinInt64/2)-1, 2)
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

func TestSafe64MulDiv(t *testing.T) {
	var err error
	var ures uint64

	ures, err = Safe64MulDiv(uint64(0), uint64(10), uint64(2))
	require.NoError(t, err)
	require.Equal(t, ures, uint64(0))

	ures, err = Safe64MulDiv(uint64(10), uint64(10), uint64(8))
	require.NoError(t, err)
	require.Equal(t, ures, uint64(12))

	ures, err = Safe64MulDiv(uint64(1), uint64(1<<43-1), uint64(1<<37-1))
	require.NoError(t, err)
	require.Equal(t, ures, uint64(64))

	ures, err = Safe64MulDiv(uint64(1<<63-1), uint64(1<<13-1), uint64(1<<50-1))
	require.NoError(t, err)
	require.Equal(t, ures, uint64(67100672))

	_, err = Safe64MulDiv(uint64(0), uint64(100), uint64(0))
	require.ErrorIs(t, err, ErrIntegerDivisionByZero)

	_, err = Safe64MulDiv(uint64(1<<63-1), uint64(1<<63-1), uint64(2))
	require.ErrorIs(t, err, ErrIntegerOverflow)
}

