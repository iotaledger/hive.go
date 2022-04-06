package serix_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
)

func TestEncode_Slice(t *testing.T) {
	t.Parallel()
	testObj := Bools{true, false, true, true}
	lenType := serializer.SeriLengthPrefixTypeAsUint16
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteSliceOfObjects(testObj, defaultSeriMode, nil, lenType, defaultArrayRules, defaultErrProducer)
	}
	ts := serix.TypeSettings{}.WithLengthPrefixType(lenType)
	testEncode(t, manualSerialization, testObj, serix.WithTypeSettings(ts))
}

func TestEncode_Struct(t *testing.T) {
	t.Parallel()
	testObj := &SimpleStruct{
		Bool:   true,
		Num:    10,
		String: "foo",
	}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteNum(simpleStructObjectCode, defaultErrProducer)
		s.WriteBool(testObj.Bool, defaultErrProducer)
		s.WriteNum(testObj.Num, defaultErrProducer)
		s.WriteString(string(testObj.String), serializer.SeriLengthPrefixTypeAsUint16, defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_Interface(t *testing.T) {
	t.Parallel()
	testObj := &StructWithInterface{
		Interface: &InterfaceImpl{},
	}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteObject(testObj.Interface.(serializer.Serializable), defaultSeriMode, nil, defaultWriteGuard, defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_BytesArray(t *testing.T) {
	t.Parallel()
	testObj := BytesArray16{1, 2, 3, 4, 5}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteBytes(testObj[:], defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_BytesSlice(t *testing.T) {
	t.Parallel()
	testObj := []byte{1, 2, 3, 4, 5}
	lenType := serializer.SeriLengthPrefixTypeAsUint32
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteVariableByteSlice(testObj, lenType, defaultErrProducer)
	}
	ts := serix.TypeSettings{}.WithLengthPrefixType(lenType)
	testEncode(t, manualSerialization, testObj, serix.WithTypeSettings(ts))
}

func TestEncode_Pointer(t *testing.T) {

}

func TestEncode_BigInt(t *testing.T) {
	t.Parallel()
	testObj := big.NewInt(100)
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteUint256(testObj, defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_Time(t *testing.T) {
	t.Parallel()
	testObj := time.Now()
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteTime(testObj, defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_Optional(t *testing.T) {

}

func TestEncode_EmbeddedStructs(t *testing.T) {
	t.Parallel()
	testObj := &StructWithEmbeddedStructs{
		unexportedStruct: unexportedStruct{Foo: 1},
		ExportedStruct:   ExportedStruct{Bar: 2},
	}
	manualSerialization := func(s *serializer.Serializer) {
		s.WriteNum(testObj.unexportedStruct.Foo, defaultErrProducer)
		s.WriteNum(exportedStructObjectCode, defaultErrProducer)
		s.WriteNum(testObj.ExportedStruct.Bar, defaultErrProducer)
	}
	testEncode(t, manualSerialization, testObj)
}

func TestEncode_Map(t *testing.T) {

}

func TestEncode_OrderedMap(t *testing.T) {

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

func testEncode(t testing.TB, manualSerializationFn SerializationFunc, testObj interface{}, opts ...serix.Option) {
	got, err := testAPI.Encode(ctx, testObj, opts...)
	require.NoError(t, err)
	ser := serializer.NewSerializer()
	manualSerializationFn(ser)
	expected, err := ser.Serialize()
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
