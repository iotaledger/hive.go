//nolint:gocritic // we don't care about these linters in test cases
package valuerange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInt8Value tests the public API of Int8Values.
func TestInt8Value(t *testing.T) {
	assert.Equal(t, -1, Int8Value(-2).Compare(Int8Value(-1)), "Int8Value(-2) should be smaller than Int8Value(-1)")
	assert.Equal(t, 0, Int8Value(-1).Compare(Int8Value(-1)), "Int8Value(-1) should be equal to Int8Value(-1)")
	assert.Equal(t, 1, Int8Value(-1).Compare(Int8Value(-2)), "Int8Value(-1) should be bigger than Int8Value(-2)")
	assert.Equal(t, -1, Int8Value(0).Compare(Int8Value(1)), "Int8Value(0) should be smaller than Int8Value(1)")
	assert.Equal(t, 0, Int8Value(0).Compare(Int8Value(0)), "Int8Value(0) should be equal to Int8Value(0)")
	assert.Equal(t, 1, Int8Value(1).Compare(Int8Value(0)), "Int8Value(1) should be bigger than Int8Value(0)")

	value := Int8Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Int8ValueType, value.Type())
	assert.Equal(t, "Int8Value(7)", value.String())
}

// TestInt16Value tests the public API of Int16Values.
func TestInt16Value(t *testing.T) {
	assert.Equal(t, -1, Int16Value(-2).Compare(Int16Value(-1)), "Int16Value(-2) should be smaller than Int16Value(-1)")
	assert.Equal(t, 0, Int16Value(-1).Compare(Int16Value(-1)), "Int16Value(-1) should be equal to Int16Value(-1)")
	assert.Equal(t, 1, Int16Value(-1).Compare(Int16Value(-2)), "Int16Value(-1) should be bigger than Int16Value(-2)")
	assert.Equal(t, -1, Int16Value(0).Compare(Int16Value(1)), "Int16Value(0) should be smaller than Int16Value(1)")
	assert.Equal(t, 0, Int16Value(0).Compare(Int16Value(0)), "Int16Value(0) should be equal to Int16Value(0)")
	assert.Equal(t, 1, Int16Value(1).Compare(Int16Value(0)), "Int16Value(1) should be bigger than Int16Value(0)")

	value := Int16Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Int16ValueType, value.Type())
	assert.Equal(t, "Int16Value(7)", value.String())
}

// TestInt32Value tests the public API of Int32Values.
func TestInt32Value(t *testing.T) {
	assert.Equal(t, -1, Int32Value(-2).Compare(Int32Value(-1)), "Int32Value(-2) should be smaller than Int32Value(-1)")
	assert.Equal(t, 0, Int32Value(-1).Compare(Int32Value(-1)), "Int32Value(-1) should be equal to Int32Value(-1)")
	assert.Equal(t, 1, Int32Value(-1).Compare(Int32Value(-2)), "Int32Value(-1) should be bigger than Int32Value(-2)")
	assert.Equal(t, -1, Int32Value(0).Compare(Int32Value(1)), "Int32Value(0) should be smaller than Int32Value(1)")
	assert.Equal(t, 0, Int32Value(0).Compare(Int32Value(0)), "Int32Value(0) should be equal to Int32Value(0)")
	assert.Equal(t, 1, Int32Value(1).Compare(Int32Value(0)), "Int32Value(1) should be bigger than Int32Value(0)")

	value := Int32Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Int32ValueType, value.Type())
	assert.Equal(t, "Int32Value(7)", value.String())
}

// TestInt64Value tests the public API of Int64Values.
func TestInt64Value(t *testing.T) {
	assert.Equal(t, -1, Int64Value(-2).Compare(Int64Value(-1)), "Int64Value(-2) should be smaller than Int64Value(-1)")
	assert.Equal(t, 0, Int64Value(-1).Compare(Int64Value(-1)), "Int64Value(-1) should be equal to Int64Value(-1)")
	assert.Equal(t, 1, Int64Value(-1).Compare(Int64Value(-2)), "Int64Value(-1) should be bigger than Int64Value(-2)")
	assert.Equal(t, -1, Int64Value(0).Compare(Int64Value(1)), "Int64Value(0) should be smaller than Int64Value(1)")
	assert.Equal(t, 0, Int64Value(0).Compare(Int64Value(0)), "Int64Value(0) should be equal to Int64Value(0)")
	assert.Equal(t, 1, Int64Value(1).Compare(Int64Value(0)), "Int64Value(1) should be bigger than Int64Value(0)")

	value := Int64Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Int64ValueType, value.Type())
	assert.Equal(t, "Int64Value(7)", value.String())
}

// TestUint8Value tests the public API of Uint8Values.
func TestUint8Value(t *testing.T) {
	assert.Equal(t, -1, Uint8Value(0).Compare(Uint8Value(1)), "Uint8Value(0) should be smaller than Uint8Value(1)")
	assert.Equal(t, 0, Uint8Value(0).Compare(Uint8Value(0)), "Uint8Value(0) should be equal to Uint8Value(0)")
	assert.Equal(t, 1, Uint8Value(1).Compare(Uint8Value(0)), "Uint8Value(1) should be bigger than Uint8Value(0)")

	value := Uint8Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Uint8ValueType, value.Type())
	assert.Equal(t, "Uint8Value(7)", value.String())
}

// TestUint16Value tests the public API of Uint16Values.
func TestUint16Value(t *testing.T) {
	assert.Equal(t, -1, Uint16Value(0).Compare(Uint16Value(1)), "Uint16Value(0) should be smaller than Uint16Value(1)")
	assert.Equal(t, 0, Uint16Value(0).Compare(Uint16Value(0)), "Uint16Value(0) should be equal to Uint16Value(0)")
	assert.Equal(t, 1, Uint16Value(1).Compare(Uint16Value(0)), "Uint16Value(1) should be bigger than Uint16Value(0)")

	value := Uint16Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Uint16ValueType, value.Type())
	assert.Equal(t, "Uint16Value(7)", value.String())
}

// TestUint32Value tests the public API of Uint32Values.
func TestUint32Value(t *testing.T) {
	assert.Equal(t, -1, Uint32Value(0).Compare(Uint32Value(1)), "Uint32Value(0) should be smaller than Uint32Value(1)")
	assert.Equal(t, 0, Uint32Value(0).Compare(Uint32Value(0)), "Uint32Value(0) should be equal to Uint32Value(0)")
	assert.Equal(t, 1, Uint32Value(1).Compare(Uint32Value(0)), "Uint32Value(1) should be bigger than Uint32Value(0)")

	value := Uint32Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Uint32ValueType, value.Type())
	assert.Equal(t, "Uint32Value(7)", value.String())
}

// TestUint64Value tests the public API of Uint64Values.
func TestUint64Value(t *testing.T) {
	assert.Equal(t, -1, Uint64Value(0).Compare(Uint64Value(1)), "Uint64Value(0) should be smaller than Uint64Value(1)")
	assert.Equal(t, 0, Uint64Value(0).Compare(Uint64Value(0)), "Uint64Value(0) should be equal to Uint64Value(0)")
	assert.Equal(t, 1, Uint64Value(1).Compare(Uint64Value(0)), "Uint64Value(1) should be bigger than Uint64Value(0)")

	value := Uint64Value(7)
	marshaledValue := value.Bytes()
	unmarshaledValue, consumedBytes, err := ValueFromBytes(marshaledValue)
	require.NoError(t, err)
	assert.Equal(t, len(marshaledValue), consumedBytes)
	assert.Equal(t, value, unmarshaledValue)
	assert.Equal(t, Uint64ValueType, value.Type())
	assert.Equal(t, "Uint64Value(7)", value.String())
}
