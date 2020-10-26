package valuerange

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"golang.org/x/xerrors"
)

// EndPoint contains information about where ValueRanges start and end. It combines a threshold value with a BoundType.
type EndPoint struct {
	value     Value
	boundType BoundType
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

// EndPointFromMarshalUtil unmarshals an EndPoint using a MarshalUtil (for easier unmarshaling).
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

// Bytes returns a marshaled version of the EndPoint.
func (e *EndPoint) Bytes() []byte {
	return marshalutil.New().
		Write(e.value).
		Write(e.boundType).
		Bytes()
}

// String returns a human readable version of the EndPoint.
func (e *EndPoint) String() string {
	return stringify.Struct("EndPoint",
		stringify.StructField("value", e.value),
		stringify.StructField("boundType", e.boundType),
	)
}
