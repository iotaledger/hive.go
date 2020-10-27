package valuerange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInt8Value tests the public API of Int8Values.
func TestInt8Value(t *testing.T) {
	assert.Equal(t, -1, Int8Value(0).Compare(Int8Value(1)), "Int8Value(0) should be smaller than Int8Value(1)")
	assert.Equal(t, 0, Int8Value(0).Compare(Int8Value(0)), "Int8Value(0) should be equal to Int8Value(0)")
	assert.Equal(t, 1, Int8Value(1).Compare(Int8Value(0)), "Int8Value(1) should be bigger than Int8Value(0)")

	value := Int8Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, "Int8Value(7)", value.String())
}

// TestInt16Value tests the public API of Int16Values.
func TestInt16Value(t *testing.T) {
	assert.Equal(t, -1, Int16Value(0).Compare(Int16Value(1)), "Int16Value(0) should be smaller than Int16Value(1)")
	assert.Equal(t, 0, Int16Value(0).Compare(Int16Value(0)), "Int16Value(0) should be equal to Int16Value(0)")
	assert.Equal(t, 1, Int16Value(1).Compare(Int16Value(0)), "Int16Value(1) should be bigger than Int16Value(0)")

	value := Int16Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, "Int16Value(7)", value.String())
}

// TestInt32Value tests the public API of Int32Values.
func TestInt32Value(t *testing.T) {
	assert.Equal(t, -1, Int32Value(0).Compare(Int32Value(1)), "Int32Value(0) should be smaller than Int32Value(1)")
	assert.Equal(t, 0, Int32Value(0).Compare(Int32Value(0)), "Int32Value(0) should be equal to Int32Value(0)")
	assert.Equal(t, 1, Int32Value(1).Compare(Int32Value(0)), "Int32Value(1) should be bigger than Int32Value(0)")

	value := Int32Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, "Int32Value(7)", value.String())
}

// TestInt64Value tests the public API of Int64Values.
func TestInt64Value(t *testing.T) {
	assert.Equal(t, -1, Int64Value(0).Compare(Int64Value(1)), "Int64Value(0) should be smaller than Int64Value(1)")
	assert.Equal(t, 0, Int64Value(0).Compare(Int64Value(0)), "Int64Value(0) should be equal to Int64Value(0)")
	assert.Equal(t, 1, Int64Value(1).Compare(Int64Value(0)), "Int64Value(1) should be bigger than Int64Value(0)")

	value := Int64Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, "Int64Value(7)", value.String())
}
