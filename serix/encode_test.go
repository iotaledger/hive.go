package serix_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestEncode_Slice(t *testing.T) {
	t.Parallel()
	testObj := Slice16LengthType{true, false, true, true}
	testEncode(t, testObj)
}

func TestEncode_Struct(t *testing.T) {
	testObj := &SimpleStruct{
		Bool:   true,
		Num:    10,
		String: String16LengthType("foo"),
	}
	testEncode(t, testObj)
}

func TestEncode_Interface(t *testing.T) {

}

func TestEncode_BytesArray(t *testing.T) {

}

func TestEncode_BytesSlice(t *testing.T) {

}

func TestEncode_Pointer(t *testing.T) {

}

func TestEncode_BigInt(t *testing.T) {

}

func TestEncode_Time(t *testing.T) {

}

func TestEncode_Payload(t *testing.T) {

}

func TestEncode_EmbeddedStruct(t *testing.T) {

}

func TestEncode_Serializable(t *testing.T) {

}

func TestEncode_SyntacticValidation(t *testing.T) {

}

func TestEncode_BytesValidation(t *testing.T) {

}

func TestEncode_ArrayRules(t *testing.T) {

}

func testEncode(t testing.TB, testObj serializer.Serializable) {
	got, err := testAPI.Encode(ctx, testObj)
	require.NoError(t, err)
	expected, err := testObj.Serialize(defaultSeriMode, nil)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
