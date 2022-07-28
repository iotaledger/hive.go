package valuerange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBoundTypeOpen tests the API of the BoundTypeOpen type.
func TestBoundTypeOpen(t *testing.T) {
	boundType := BoundTypeOpen
	assert.Equal(t, "BoundTypeOpen", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledBoundType), consumedBytes)
	assert.Equal(t, boundType, unmarshaledBoundType)
}

// TestBoundTypeClosed tests the API of the BoundTypeClosed type.
func TestBoundTypeClosed(t *testing.T) {
	boundType := BoundTypeClosed
	assert.Equal(t, "BoundTypeClosed", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledBoundType), consumedBytes)
	assert.Equal(t, boundType, unmarshaledBoundType)
}

// TestBoundTypeClosed tests the API of the BoundTypeClosed type.
func TestBoundTypeUnknown(t *testing.T) {
	boundType := BoundType(17)
	assert.Equal(t, "BoundType(11)", boundType.String())

	marshaledBoundType := boundType.Bytes()
	unmarshaledBoundType, consumedBytes, err := BoundTypeFromBytes(marshaledBoundType)
	require.Error(t, err)
	assert.Equal(t, 0, consumedBytes)
	assert.Equal(t, boundType, unmarshaledBoundType)
}
