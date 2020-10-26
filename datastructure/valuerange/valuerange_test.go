package valuerange

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueRange_Open(t *testing.T) {
	// create ValueRange for tests
	valueRange := Open(Int64Value(10), Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")
	assert.False(t, Open(Int64Value(10), Int64Value(11)).Empty(), "the ValueRange should not be empty")

	// test Has...Bound method
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeOpen, "the lower bound should be Open")
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeOpen, "the lower bound should be Open")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(10)), "the ValueRange should be larger than Int64Value(10)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(114)), "the ValueRange should be smaller than Int64Value(114)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(10)), "the ValueRange should not contain Int64Value(10)")
	assert.False(t, valueRange.Contains(Int64Value(114)), "the ValueRange should not contain Int64Value(114)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange)
}

func TestValueRange_Compare(t *testing.T) {
	valueRange0 := All()
	fmt.Println(valueRange0)

	valueRangeAtMost := AtMost(Int64Value(100))
	fmt.Println(valueRangeAtMost)

	valueRange1 := Open(Int64Value(10), Int64Value(14))
	fmt.Println(valueRange1)

	valueRange2 := Closed(Int64Value(10), Int64Value(14))
	fmt.Println(valueRange2)

	valueRange3 := GreaterThan(Int64Value(10))
	fmt.Println(valueRange3)

	fmt.Print(valueRange1.Contains(Int64Value(13)))
}
