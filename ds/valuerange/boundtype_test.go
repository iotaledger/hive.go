package valuerange

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBoundTypeOpen tests the API of the BoundTypeOpen type.
func TestBoundTypeOpen(t *testing.T) {
	boundType := BoundTypeOpen
	require.Equal(t, "BoundTypeOpen", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.NoError(t, err)
	require.Equal(t, len(marshaledBoundType), consumedBytes)
	require.Equal(t, boundType, unmarshaledBoundType)
}

// TestBoundTypeClosed tests the API of the BoundTypeClosed type.
func TestBoundTypeClosed(t *testing.T) {
	boundType := BoundTypeClosed
	require.Equal(t, "BoundTypeClosed", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.NoError(t, err)
	require.Equal(t, len(marshaledBoundType), consumedBytes)
	require.Equal(t, boundType, unmarshaledBoundType)
}

// TestBoundTypeClosed tests the API of the BoundTypeClosed type.
func TestBoundTypeUnknown(t *testing.T) {
	boundType := BoundType(17)
	require.Equal(t, "BoundType(11)", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.Error(t, err)
	require.Equal(t, 0, consumedBytes)
	require.Equal(t, boundType, unmarshaledBoundType)
}
