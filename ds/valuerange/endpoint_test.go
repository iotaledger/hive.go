package valuerange

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEndPoint_BoundType tests if the getter of the Value works correctly.
func TestEndPoint_Value(t *testing.T) {
	require.Equal(t, Int8Value(1), NewEndPoint(Int8Value(1), BoundTypeOpen).Value())
	require.Equal(t, Int8Value(0), NewEndPoint(Int8Value(0), BoundTypeOpen).Value())
	require.Equal(t, Int8Value(-1), NewEndPoint(Int8Value(-1), BoundTypeOpen).Value())
}

// TestEndPoint_BoundType tests if the getter of the BoundType works correctly.
func TestEndPoint_BoundType(t *testing.T) {
	require.Equal(t, BoundTypeOpen, NewEndPoint(Int8Value(1), BoundTypeOpen).BoundType())
	require.Equal(t, BoundTypeClosed, NewEndPoint(Int8Value(1), BoundTypeClosed).BoundType())
}

// TestEndPoint_MarshalUnmarshal tests if marshaling and unmarshalling of EndPoint works correctly.
func TestEndPoint_MarshalUnmarshal(t *testing.T) {
	endPoint := NewEndPoint(Int8Value(1), BoundTypeOpen)
	marshaledEndPoint := endPoint.Bytes()
	unmarshaledEndPoint, consumedBytes, err := EndPointFromBytes(marshaledEndPoint)
	require.NoError(t, err)
	require.Equal(t, len(marshaledEndPoint), consumedBytes)
	require.Equal(t, endPoint, unmarshaledEndPoint)

	endPoint = NewEndPoint(Int8Value(2), BoundTypeClosed)
	marshaledEndPoint = endPoint.Bytes()
	unmarshaledEndPoint, consumedBytes, err = EndPointFromBytes(marshaledEndPoint)
	require.NoError(t, err)
	require.Equal(t, len(marshaledEndPoint), consumedBytes)
	require.Equal(t, endPoint, unmarshaledEndPoint)
}
