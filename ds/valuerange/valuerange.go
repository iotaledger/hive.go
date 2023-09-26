package valuerange

import (
	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/ds/bitmask"
	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

// ValueRange defines the boundaries around a contiguous span of Values (i.e. "integers from 1 to 100 inclusive").
//
// It is not possible to iterate over the contained values. Each Range may be bounded or unbounded. If bounded, there is
// an associated endpoint value and the range is considered to be either open (does not include the endpoint) or closed
// (includes the endpoint) on that side.
//
// With three possibilities on each side, this yields nine basic types of ranges, enumerated below:
//
// Notation         Definition          Factory method
// (a .. b)         {x | a < x < b}     Open
// [a .. b]         {x | a <= x <= b}   Closed
// (a .. b]         {x | a < x <= b}    OpenClosed
// [a .. b)         {x | a <= x < b}    ClosedOpen
// (a .. +INF)      {x | x > a}         GreaterThan
// [a .. +INF)      {x | x >= a}        AtLeast
// (-INF .. b)      {x | x < b}         LessThan
// (-INF .. b]      {x | x <= b}        AtMost
// (-INF .. +INF)   {x}        			All
//
// When both endpoints exist, the upper endpoint may not be less than the lower. The endpoints may be equal only if at
// least one of the bounds is closed.
type ValueRange struct {
	lowerEndPoint *EndPoint
	upperEndPoint *EndPoint
}

// FromBytes unmarshals a ValueRange from a sequence of bytes.
func FromBytes(valueRangeBytes []byte) (valueRange *ValueRange, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(valueRangeBytes)
	if valueRange, err = FromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse ValueRange from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// FromMarshalUtil unmarshals a ValueRange using a MarshalUtil (for easier unmarshalling).
func FromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (valueRange *ValueRange, err error) {
	endPointExistsMaskByte, err := marshalUtil.ReadByte()
	if err != nil {
		err = xerrors.Errorf("failed to read endpoint exists mask (%v): %w", err, ErrParseBytesFailed)

		return
	}

	valueRange = &ValueRange{}
	endPointExistsMask := bitmask.BitMask(endPointExistsMaskByte)
	if endPointExistsMask.HasBit(0) {
		if valueRange.lowerEndPoint, err = EndPointFromMarshalUtil(marshalUtil); err != nil {
			err = xerrors.Errorf("failed to parse lower EndPoint from MarshalUtil: %w", ErrParseBytesFailed)

			return
		}
	}
	if endPointExistsMask.HasBit(1) {
		if valueRange.upperEndPoint, err = EndPointFromMarshalUtil(marshalUtil); err != nil {
			err = xerrors.Errorf("failed to parse upper EndPoint from MarshalUtil: %w", ErrParseBytesFailed)

			return
		}
	}

	return
}

// All returns a ValueRange that contains all possible Values.
func All() *ValueRange {
	return &ValueRange{}
}

// AtLeast returns a ValueRange that contains all Values greater than or equal to lower.
func AtLeast(lower Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
	}
}

// AtMost returns a ValueRange that contains all Values less than or equal to upper.
func AtMost(upper Value) *ValueRange {
	return &ValueRange{
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// Closed returns a ValueRange that contains all Values greater than or equal to lower and less than or equal to upper.
func Closed(lower Value, upper Value) *ValueRange {
	if lower.Compare(upper) == 1 {
		panic("lower needs to be smaller or equal than upper")
	}

	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// ClosedOpen returns a ValueRange that contains all Values greater than or equal to lower and strictly less than upper.
func ClosedOpen(lower Value, upper Value) *ValueRange {
	if lower.Compare(upper) == 1 {
		panic("lower needs to be smaller or equal than upper")
	}

	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

// GreaterThan returns a ValueRange that contains all Values strictly greater than lower.
func GreaterThan(lower Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
	}
}

// LessThan returns a ValueRange that contains all values strictly less than upper.
func LessThan(upper Value) *ValueRange {
	return &ValueRange{
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

// Open returns a ValueRange that contains all Values strictly greater than lower and strictly less than upper.
func Open(lower Value, upper Value) *ValueRange {
	if lower.Compare(upper) != -1 {
		panic("lower needs to be smaller than upper")
	}

	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

// OpenClosed returns a ValueRange that contains all values strictly greater than lower and less than or equal to upper.
func OpenClosed(lower Value, upper Value) *ValueRange {
	if lower.Compare(upper) == 1 {
		panic("lower needs to be smaller or equal than upper")
	}

	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// Compare returns 0 if the ValueRange contains the given Value, -1 if its contained Values are smaller and 1 if they
// are bigger.
func (v *ValueRange) Compare(value Value) int {
	if v.lowerEndPoint == nil {
		if v.upperEndPoint == nil {
			return 0
		}

		if cmp := v.upperEndPoint.value.Compare(value); cmp == 1 || (cmp == 0 && v.upperEndPoint.boundType == BoundTypeClosed) {
			return 0
		}

		return -1
	}

	if v.upperEndPoint == nil {
		if v.lowerEndPoint == nil {
			return 0
		}

		if cmp := v.lowerEndPoint.value.Compare(value); cmp == -1 || (cmp == 0 && v.lowerEndPoint.boundType == BoundTypeClosed) {
			return 0
		}

		return 1
	}

	if cmp := v.lowerEndPoint.value.Compare(value); cmp == 1 || (cmp == 0 && v.lowerEndPoint.boundType == BoundTypeOpen) {
		return 1
	}

	if cmp := v.upperEndPoint.value.Compare(value); cmp == -1 || (cmp == 0 && v.upperEndPoint.boundType == BoundTypeOpen) {
		return -1
	}

	return 0
}

// Contains returns true if value is within the bounds of this ValueRange.
func (v *ValueRange) Contains(value Value) bool {
	return v.Compare(value) == 0
}

// Empty returns true if this range is of the form [v..v) or (v..v]. This does not encompass ranges of the form (v..v),
// because such ranges are invalid and can't be constructed at all.
//
// Note that certain discrete ranges such as the integer range (3..4) are not considered empty, even though they contain
// no actual values.
func (v *ValueRange) Empty() bool {
	return v.lowerEndPoint != nil && v.upperEndPoint != nil && v.lowerEndPoint.value.Compare(v.upperEndPoint.value) == 0
}

// HasLowerBound returns true if this ValueRange has a lower EndPoint.
func (v *ValueRange) HasLowerBound() bool {
	return v.lowerEndPoint != nil
}

// HasUpperBound returns true if this ValueRange has an upper EndPoint.
func (v *ValueRange) HasUpperBound() bool {
	return v.upperEndPoint != nil
}

// LowerBoundType returns the type of this ValueRange's lower bound - BoundTypeClosed if the range includes its lower
// EndPoint and BoundTypeOpen if it does not include its lower EndPoint.
func (v *ValueRange) LowerBoundType() BoundType {
	if v.lowerEndPoint == nil {
		panic("ValueRange has no lower bound - check HasLowerBound() before calling this method")
	}

	return v.lowerEndPoint.boundType
}

// LowerEndPoint returns the lower EndPoint of this ValueRange. It panics if the ValueRange has no lower EndPoint.
func (v *ValueRange) LowerEndPoint() *EndPoint {
	if v.lowerEndPoint == nil {
		panic("ValueRange has no lower EndPoint - check HasLowerBound() before calling this method")
	}

	return v.lowerEndPoint
}

// UpperBoundType returns the type of this ValueRange's upper bound - BoundTypeClosed if the range includes its upper
// EndPoint and BoundTypeOpen if it does not include its upper EndPoint.
func (v *ValueRange) UpperBoundType() BoundType {
	if v.upperEndPoint == nil {
		panic("ValueRange has no upper bound - check HasUpperBound() before calling this method")
	}

	return v.upperEndPoint.boundType
}

// UpperEndPoint returns the upper EndPoint of this Range or nil if it is unbounded.
func (v *ValueRange) UpperEndPoint() *EndPoint {
	if v.upperEndPoint == nil {
		panic("ValueRange has no upper EndPoint - check HasUpperBound() before calling this method")
	}

	return v.upperEndPoint
}

// Bytes returns a marshaled version of the ValueRange.
func (v *ValueRange) Bytes() []byte {
	var endPointExistsMask bitmask.BitMask
	if v.lowerEndPoint != nil {
		endPointExistsMask = endPointExistsMask.SetBit(0)
	}
	if v.upperEndPoint != nil {
		endPointExistsMask = endPointExistsMask.SetBit(1)
	}

	marshalUtil := marshalutil.New()
	marshalUtil.WriteByte(byte(endPointExistsMask))
	if endPointExistsMask.HasBit(0) {
		marshalUtil.Write(v.lowerEndPoint)
	}
	if endPointExistsMask.HasBit(1) {
		marshalUtil.Write(v.upperEndPoint)
	}

	return marshalUtil.Bytes()
}

// String returns a human-readable version of the ValueRange.
func (v *ValueRange) String() string {
	var lowerEndPoint string
	switch {
	case v.lowerEndPoint == nil:
		lowerEndPoint = "(-INF"
	case v.lowerEndPoint.boundType == BoundTypeOpen:
		lowerEndPoint = "(" + v.lowerEndPoint.value.String()
	case v.lowerEndPoint.boundType == BoundTypeClosed:
		lowerEndPoint = "[" + v.lowerEndPoint.value.String()
	}

	var upperEndPoint string
	switch {
	case v.upperEndPoint == nil:
		upperEndPoint = "+INF)"
	case v.upperEndPoint.boundType == BoundTypeOpen:
		upperEndPoint = v.upperEndPoint.value.String() + ")"
	case v.upperEndPoint.boundType == BoundTypeClosed:
		upperEndPoint = v.upperEndPoint.value.String() + "]"
	}

	return "ValueRange" + lowerEndPoint + " ... " + upperEndPoint
}
