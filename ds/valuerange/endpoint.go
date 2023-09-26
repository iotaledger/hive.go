package valuerange

import (
	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
	"github.com/izuc/zipp.foundation/stringify"
)

// EndPoint contains information about where ValueRanges start and end. It combines a threshold value with a BoundType.
type EndPoint struct {
	value     Value
	boundType BoundType
}

// NewEndPoint create a new EndPoint from the given details.
func NewEndPoint(value Value, boundType BoundType) *EndPoint {
	return &EndPoint{
		value:     value,
		boundType: boundType,
	}
}

// EndPointFromBytes unmarshals an EndPoint from a sequence of bytes.
func EndPointFromBytes(endPointBytes []byte) (endPoint *EndPoint, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(endPointBytes)
	if endPoint, err = EndPointFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse EndPoint from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// EndPointFromMarshalUtil unmarshals an EndPoint using a MarshalUtil (for easier unmarshalling).
func EndPointFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (endPoint *EndPoint, err error) {
	endPoint = &EndPoint{}
	if endPoint.value, err = ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Value from MarshalUtil: %w", err)

		return
	}
	if endPoint.boundType, err = BoundTypeFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse BoundType from MarshalUtil: %w", err)

		return
	}

	return
}

// Value returns the Value of the EndPoint.
func (e *EndPoint) Value() Value {
	return e.value
}

// BoundType returns the BoundType of the EndPoint.
func (e *EndPoint) BoundType() BoundType {
	return e.boundType
}

// Bytes returns a marshaled version of the EndPoint.
func (e *EndPoint) Bytes() []byte {
	return marshalutil.New().
		Write(e.value).
		Write(e.boundType).
		Bytes()
}

// String returns a human-readable version of the EndPoint.
func (e *EndPoint) String() string {
	return stringify.Struct("EndPoint",
		stringify.NewStructField("value", e.value),
		stringify.NewStructField("boundType", e.boundType),
	)
}
