package valuerange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndPoint_BoundType tests if the getter of the Value works correctly.
func TestEndPoint_Value(t *testing.T) {
	assert.Equal(t, Int8Value(1), NewEndPoint(Int8Value(1), BoundTypeOpen).Value())
	assert.Equal(t, Int8Value(0), NewEndPoint(Int8Value(0), BoundTypeOpen).Value())
	assert.Equal(t, Int8Value(-1), NewEndPoint(Int8Value(-1), BoundTypeOpen).Value())
}

// TestEndPoint_BoundType tests if the getter of the BoundType works correctly.
func TestEndPoint_BoundType(t *testing.T) {
	assert.Equal(t, BoundTypeOpen, NewEndPoint(Int8Value(1), BoundTypeOpen).BoundType())
	assert.Equal(t, BoundTypeClosed, NewEndPoint(Int8Value(1), BoundTypeClosed).BoundType())
}

// TestEndPoint_MarshalUnmarshal tests if marshaling and unmarshaling of EndPoint works correctly.
func TestEndPoint_MarshalUnmarshal(t *testing.T) {
	endPoint := NewEndPoint(Int8Value(1), BoundTypeOpen)
	marshaledEndPoint := endPoint.Bytes()
	unmarshaledEndPoint, consumedBytes, err := EndPointFromBytes(marshaledEndPoint)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledEndPoint), consumedBytes)
	assert.Equal(t, endPoint, unmarshaledEndPoint)

	endPoint = NewEndPoint(Int8Value(2), BoundTypeClosed)
	marshaledEndPoint = endPoint.Bytes()
	unmarshaledEndPoint, consumedBytes, err = EndPointFromBytes(marshaledEndPoint)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledEndPoint), consumedBytes)
	assert.Equal(t, endPoint, unmarshaledEndPoint)
}
