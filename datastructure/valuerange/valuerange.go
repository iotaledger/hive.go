package valuerange

// ValueRange defines the boundaries around a contiguous span of Values (i.e. "integers from 1 to 100 inclusive").
//
// It is not possible to iterate over the contained values. Each Range may be bounded or unbounded. If bounded, there is
// an associated endpoint value, and the range is considered to be either open (does not include the endpoint) or closed
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

// All returns a ValueRange that contains all possible Values.
func All() *ValueRange {
	return &ValueRange{}
}

// AtLeast returns a ValueRange that contains all Values greater than or equal to the lower EndPoint.
func AtLeast(lower Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
	}
}

// AtMost returns a ValueRange that contains all Values less than or equal to the upper EndPoint.
func AtMost(upper Value) *ValueRange {
	return &ValueRange{
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// Closed returns a ValueRange that contains all Values greater than or equal to lower and less than or equal to upper.
func Closed(lower Value, upper Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// ClosedOpen returns a ValueRange that contains all Values greater than or equal to lower and strictly less than upper.
func ClosedOpen(lower Value, upper Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeClosed},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

// GreaterThan returns a ValueRange that contains all Values strictly greater than endpoint.
func GreaterThan(lower Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
	}
}

func LessThan(upper Value) *ValueRange {
	return &ValueRange{
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

// Open returns a ValueRange that contains all Values strictly greater than lower and strictly less than upper.
func Open(lower Value, upper Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeOpen},
	}
}

func OpenClosed(lower Value, upper Value) *ValueRange {
	return &ValueRange{
		lowerEndPoint: &EndPoint{value: lower, boundType: BoundTypeOpen},
		upperEndPoint: &EndPoint{value: upper, boundType: BoundTypeClosed},
	}
}

// Compare returns 0 if the ValueRange contains the given Value, -1 if its Values are smaller and 1 if its Values are
// bigger.
func (r *ValueRange) Compare(value Value) int {
	if r.lowerEndPoint == nil {
		if cmp := r.upperEndPoint.value.Compare(value); cmp == 1 || (cmp == 0 && r.upperEndPoint.boundType == BoundTypeClosed) {
			return 0
		}

		return -1
	}

	if r.upperEndPoint == nil {
		if cmp := r.lowerEndPoint.value.Compare(value); cmp == -1 || (cmp == 0 && r.lowerEndPoint.boundType == BoundTypeClosed) {
			return 0
		}

		return 1
	}

	if cmp := r.lowerEndPoint.value.Compare(value); cmp == 1 || (cmp == 0 && r.lowerEndPoint.boundType == BoundTypeOpen) {
		return 1
	}

	if cmp := r.upperEndPoint.value.Compare(value); cmp == -1 || (cmp == 0 && r.lowerEndPoint.boundType == BoundTypeOpen) {
		return -1
	}

	return 0
}

// Contains returns true if value is within the bounds of this ValueRange.
func (r *ValueRange) Contains(value Value) bool {
	return r.Compare(value) == 0
}

// Empty returns true if this range is of the form [v..v) or (v..v].
func (r *ValueRange) Empty() bool {
	return false
}

// HasLowerBound returns true if this ValueRange has a lower EndPoint.
func (r *ValueRange) HasLowerBound() bool {
	return r.lowerEndPoint != nil
}

// HasUpperBound returns true if this ValueRange has an upper EndPoint.
func (r *ValueRange) HasUpperBound() bool {
	return r.upperEndPoint != nil
}

// LowerBoundType returns the type of this ValueRange's lower bound - BoundTypeClosed if the range includes its lower
// EndPoint and BoundTypeOpen if it does not include its lower EndPoint.
func (r *ValueRange) LowerBoundType() BoundType {
	if r.lowerEndPoint == nil {
		panic("ValueRange has no lower bound - check HasLowerBound() before calling this method")
	}

	return r.lowerEndPoint.boundType
}

// LowerEndPoint returns the lower EndPoint of this Range or nil if it is unbounded.
func (r *ValueRange) LowerEndPoint() *EndPoint {
	return r.lowerEndPoint
}

// UpperBoundType returns the type of this ValueRange's upper bound - BoundTypeClosed if the range includes its upper
// EndPoint and BoundTypeOpen if it does not include its upper EndPoint.
func (r *ValueRange) UpperBoundType() BoundType {
	if r.upperEndPoint == nil {
		panic("ValueRange has no upper bound - check HasUpperBound() before calling this method")
	}

	return r.upperEndPoint.boundType
}

// UpperEndPoint returns the upper EndPoint of this Range or nil if it is unbounded.
func (r *ValueRange) UpperEndPoint() *EndPoint {
	return r.upperEndPoint
}

func (r *ValueRange) String() string {
	var lowerEndPoint string
	switch {
	case r.lowerEndPoint == nil:
		lowerEndPoint = "ValueRange(-INF"
	case r.lowerEndPoint.boundType == BoundTypeOpen:
		lowerEndPoint = "ValueRange(" + r.lowerEndPoint.value.String()
	case r.lowerEndPoint.boundType == BoundTypeClosed:
		lowerEndPoint = "ValueRange[" + r.lowerEndPoint.value.String()
	}

	var upperEndPoint string
	switch {
	case r.upperEndPoint == nil:
		upperEndPoint = "+INF)"
	case r.upperEndPoint.boundType == BoundTypeOpen:
		upperEndPoint = r.upperEndPoint.value.String() + ")"
	case r.upperEndPoint.boundType == BoundTypeClosed:
		upperEndPoint = r.upperEndPoint.value.String() + "]"
	}

	return lowerEndPoint + " ... " + upperEndPoint
}
