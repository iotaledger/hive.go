package valuerange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValueRange_Open tests the Open ValueRange which contains the elements: {x | lower < x < upper}.
func TestValueRange_Open(t *testing.T) {
	// create open ValueRange for tests
	valueRange := Open(Int64Value(10), Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")
	assert.False(t, Open(Int64Value(10), Int64Value(11)).Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeOpen, "the lower bound should be Open")
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeOpen, "the lower bound should be Open")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeOpen}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeOpen}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 1, valueRange.Compare(Int64Value(10)), "the ValueRange should be larger than Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(114)), "the ValueRange should be smaller than Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller than Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.False(t, valueRange.Contains(Int64Value(10)), "the ValueRange should not contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.False(t, valueRange.Contains(Int64Value(114)), "the ValueRange should not contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_Closed tests the Closed ValueRange which contains the elements: {x | lower <= x <= upper}.
func TestValueRange_Closed(t *testing.T) {
	// create open ValueRange for tests
	valueRange := Closed(Int64Value(10), Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")
	assert.True(t, Closed(Int64Value(10), Int64Value(10)).Empty(), "the ValueRange should be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeClosed, "the lower bound should be Closed")
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeClosed, "the lower bound should be Closed")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeClosed}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeClosed}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_OpenClosed tests the OpenClosed ValueRange which contains the elements: {x | lower < x <= upper}.
func TestValueRange_OpenClosed(t *testing.T) {
	// create open ValueRange for tests
	valueRange := OpenClosed(Int64Value(10), Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")
	assert.True(t, OpenClosed(Int64Value(10), Int64Value(10)).Empty(), "the ValueRange should be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeOpen, "the lower bound should be Open")
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeClosed, "the lower bound should be Closed")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeOpen}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeClosed}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 1, valueRange.Compare(Int64Value(10)), "the ValueRange should be larger than Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller than Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.False(t, valueRange.Contains(Int64Value(10)), "the ValueRange should not contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_ClosedOpen tests the ClosedOpen ValueRange which contains the elements: {x | lower <= x < upper}.
func TestValueRange_ClosedOpen(t *testing.T) {
	// create open ValueRange for tests
	valueRange := ClosedOpen(Int64Value(10), Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")
	assert.True(t, ClosedOpen(Int64Value(10), Int64Value(10)).Empty(), "the ValueRange should be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeClosed, "the lower bound should be Closed")
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeOpen, "the lower bound should be Open")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeClosed}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeOpen}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(114)), "the ValueRange should be smaller than Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller than Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.False(t, valueRange.Contains(Int64Value(114)), "the ValueRange should not contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_GreaterThan tests the GreaterThan ValueRange which contains the elements: {x | x > lower}.
func TestValueRange_GreaterThan(t *testing.T) {
	// create open ValueRange for tests
	valueRange := GreaterThan(Int64Value(10))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.False(t, valueRange.HasUpperBound(), "the ValueRange should not have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeOpen, "the lower bound should be Open")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeOpen}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 1, valueRange.Compare(Int64Value(10)), "the ValueRange should be larger than Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.False(t, valueRange.Contains(Int64Value(10)), "the ValueRange should not contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.True(t, valueRange.Contains(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_AtLeast tests the AtLeast ValueRange which contains the elements: {x | x >= lower}.
func TestValueRange_AtLeast(t *testing.T) {
	// create open ValueRange for tests
	valueRange := AtLeast(Int64Value(10))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.True(t, valueRange.HasLowerBound(), "the ValueRange should have a lower bound")
	assert.False(t, valueRange.HasUpperBound(), "the ValueRange should not have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.LowerBoundType(), BoundTypeClosed, "the lower bound should be Closed")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(10), boundType: BoundTypeClosed}, valueRange.LowerEndPoint(), "the lower EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 1, valueRange.Compare(Int64Value(9)), "the ValueRange should be larger than Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test Contains
	assert.False(t, valueRange.Contains(Int64Value(9)), "the ValueRange should not contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.True(t, valueRange.Contains(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_LessThan tests the LessThan ValueRange which contains the elements: {x | x < upper}.
func TestValueRange_LessThan(t *testing.T) {
	// create open ValueRange for tests
	valueRange := LessThan(Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.False(t, valueRange.HasLowerBound(), "the ValueRange should not have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeOpen, "the upper bound should be Open")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeOpen}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 0, valueRange.Compare(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(114)), "the ValueRange should be smaller than Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller than Int64Value(115)")

	// test Contains
	assert.True(t, valueRange.Contains(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.False(t, valueRange.Contains(Int64Value(114)), "the ValueRange should not contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_AtMost tests the AtMost ValueRange which contains the elements: {x | x <= upper}.
func TestValueRange_AtMost(t *testing.T) {
	// create open ValueRange for tests
	valueRange := AtMost(Int64Value(114))

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.False(t, valueRange.HasLowerBound(), "the ValueRange should not have a lower bound")
	assert.True(t, valueRange.HasUpperBound(), "the ValueRange should have an upper bound")

	// test ...BoundType methods
	assert.Equal(t, valueRange.UpperBoundType(), BoundTypeClosed, "the upper bound should be Closed")

	// test ...EndPoint methods
	assert.Equal(t, &EndPoint{value: Int64Value(114), boundType: BoundTypeClosed}, valueRange.UpperEndPoint(), "the upper EndPoint should be equal to the expected value")

	// test Compare
	assert.Equal(t, 0, valueRange.Compare(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, -1, valueRange.Compare(Int64Value(115)), "the ValueRange should be smaller than Int64Value(115)")

	// test Contains
	assert.True(t, valueRange.Contains(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(11)), "the ValueRange should contain Int64Value(11)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.False(t, valueRange.Contains(Int64Value(115)), "the ValueRange should not contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}

// TestValueRange_All tests the All ValueRange which contains all elements: {x}.
func TestValueRange_All(t *testing.T) {
	// create open ValueRange for tests
	valueRange := All()

	// test Empty
	assert.False(t, valueRange.Empty(), "the ValueRange should not be empty")

	// test Has...Bound methods
	assert.False(t, valueRange.HasLowerBound(), "the ValueRange should not have a lower bound")
	assert.False(t, valueRange.HasUpperBound(), "the ValueRange should not have an upper bound")

	// test Compare
	assert.Equal(t, 0, valueRange.Compare(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.Equal(t, 0, valueRange.Compare(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test Contains
	assert.True(t, valueRange.Contains(Int64Value(9)), "the ValueRange should contain Int64Value(9)")
	assert.True(t, valueRange.Contains(Int64Value(10)), "the ValueRange should contain Int64Value(10)")
	assert.True(t, valueRange.Contains(Int64Value(50)), "the ValueRange should contain Int64Value(50)")
	assert.True(t, valueRange.Contains(Int64Value(114)), "the ValueRange should contain Int64Value(114)")
	assert.True(t, valueRange.Contains(Int64Value(115)), "the ValueRange should contain Int64Value(115)")

	// test marshaling and unmarshaling
	valueRangeBytes := valueRange.Bytes()
	restoredValueRange, consumedBytes, err := FromBytes(valueRangeBytes)
	require.NoError(t, err)
	assert.Equal(t, len(valueRangeBytes), consumedBytes, "parsing the ValueRange should consume all available bytes")
	assert.Equal(t, valueRange, restoredValueRange, "the restored ValueRange should be equal to the original one")
}
