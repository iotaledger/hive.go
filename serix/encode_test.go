package serix_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode_Slice(t *testing.T) {
	t.Parallel()
	testObj := SliceLengthType16{true, false, true, true}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteSliceOfObjects(testObj, defaultSeriMode, nil, testObj.LengthPrefixType(), defaultArrayRules, defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_Struct(t *testing.T) {
	t.Parallel()
	testObj := &SimpleStruct{
		Bool:   true,
		Num:    10,
		String: String16LengthType("foo"),
	}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteBool(testObj.Bool, defaultErrProducer)
		s.WriteNum(testObj.Num, defaultErrProducer)
		s.WriteString(string(testObj.String), testObj.String.LengthPrefixType(), defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_Interface(t *testing.T) {
	t.Parallel()
	testObj := &StructWithInterface{
		Interface: &InterfaceImpl{},
	}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteObject(testObj.Interface.(serializer.Serializable), defaultSeriMode, nil, defaultWriteGuard, defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_BytesArray(t *testing.T) {
	t.Parallel()
	testObj := BytesArray16{1, 2, 3, 4, 5}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteBytes(testObj[:], defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_BytesSlice(t *testing.T) {
	t.Parallel()
	testObj := BytesSliceLengthType32{1, 2, 3, 4, 5}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteVariableByteSlice(testObj[:], testObj.LengthPrefixType(), defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_Pointer(t *testing.T) {

}

func TestEncode_BigInt(t *testing.T) {
	t.Parallel()
	testObj := big.NewInt(100)
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteUint256(testObj, defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
}

func TestEncode_Time(t *testing.T) {
	t.Parallel()
	testObj := time.Now()
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteTime(testObj, defaultErrProducer)
	}
	testEncode(t, testObj, manualSerialization)
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

type SerializationFunc func(s *serializer.Serializer)

func testEncode(t testing.TB, testObj interface{}, manualSerializationFn SerializationFunc) {
	got, err := testAPI.Encode(ctx, testObj)
	require.NoError(t, err)
	ser := serializer.NewSerializer()
	manualSerializationFn(ser)
	expected, err := ser.Serialize()
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
