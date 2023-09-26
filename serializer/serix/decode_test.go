package serix_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/serializer/v2"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func TestDecode_Slice(t *testing.T) {
	t.Parallel()
	testObj := Bools{true, false, true, true}
	ts := serix.TypeSettings{}.WithLengthPrefixType(boolsLenType)
	testDecode(t, testObj, serix.WithTypeSettings(ts))
}

func TestDecode_Struct(t *testing.T) {
	t.Parallel()
	testObj := NewSimpleStruct()
	testDecode(t, testObj)
}

func TestDecode_Interface(t *testing.T) {
	t.Parallel()
	testObj := StructWithInterface{
		Interface: &InterfaceImpl{
			interfaceImpl{
				A: 1,
				B: 2,
			},
		},
	}
	testDecode(t, testObj)
}

func TestDecode_Pointer(t *testing.T) {
	t.Parallel()
	ss := NewSimpleStruct()
	testObj := &ss
	testDecode(t, testObj)
}

func TestDecode_Optional(t *testing.T) {
	t.Parallel()
	testObj := StructWithOptionalField{Optional: nil}
	testDecode(t, testObj)
}

func TestDecode_EmbeddedStructs(t *testing.T) {
	t.Parallel()
	testObj := StructWithEmbeddedStructs{
		unexportedStruct: unexportedStruct{Foo: 1},
		ExportedStruct:   ExportedStruct{Bar: 2},
	}
	testDecode(t, testObj)
}

func TestDecode_Map(t *testing.T) {
	t.Parallel()
	testObj := Map{
		0: 2,
		1: 4,
	}
	testDecode(t, testObj, serix.WithTypeSettings(serix.TypeSettings{}.WithLengthPrefixType(mapLenType)))
}

func TestDecode_Deserializable(t *testing.T) {
	t.Parallel()
	testObject := CustomSerializable(2)
	testDecode(t, testObject)
}

func TestDecode_DeserializablePointer(t *testing.T) {
	t.Parallel()
	cs := CustomSerializable(2)
	testObject := &cs
	testDecode(t, testObject)
}

func TestDecode_SyntacticValidation(t *testing.T) {
	t.Parallel()
	testObj := &ObjectForSyntacticValidation{}
	bytesRead, err := testAPI.Decode(ctx, nil, testObj, serix.WithValidation())
	require.Zero(t, bytesRead)
	assert.ErrorIs(t, err, errSyntacticValidation)
}

func TestDecode_BytesValidation(t *testing.T) {
	t.Parallel()
	testObj := &ObjectForBytesValidation{}
	bytesRead, err := testAPI.Decode(ctx, nil, testObj, serix.WithValidation())
	require.Zero(t, bytesRead)
	assert.ErrorIs(t, err, errBytesValidation)
}

func TestDecode_ArrayRules(t *testing.T) {
	t.Parallel()
	testObj := &Bools{true, false, true, true}
	bytes, err := testObj.Serialize(defaultSeriMode, nil)
	require.NoError(t, err)
	rules := &serix.ArrayRules{Min: 5}
	ts := serix.TypeSettings{}.WithLengthPrefixType(boolsLenType).WithArrayRules(rules)
	bytesRead, err := testAPI.Decode(ctx, bytes, testObj, serix.WithValidation(), serix.WithTypeSettings(ts))
	require.Zero(t, bytesRead)
	assert.Contains(t, err.Error(), "min count of elements within the array not reached")
}

func testDecode(t testing.TB, expected serializer.Serializable, opts ...serix.Option) {
	bytes, err := expected.Serialize(defaultSeriMode, nil)
	require.NoError(t, err)
	got := reflect.New(reflect.TypeOf(expected)).Elem()
	bytesRead, err := testAPI.Decode(ctx, bytes, got.Addr().Interface(), opts...)
	require.NoError(t, err)
	assert.Equal(t, expected, got.Interface())
	assert.Equal(t, len(bytes), bytesRead)
}
