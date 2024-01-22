package serix_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func TestEncode_Slice(t *testing.T) {
	t.Parallel()
	testObj := Bools{true, false, true, true}
	ts := serix.TypeSettings{}.WithLengthPrefixType(boolsLenType)
	testEncode(t, testObj, serix.WithTypeSettings(ts))
}

func TestEncode_Struct(t *testing.T) {
	t.Parallel()
	testObj := NewSimpleStruct()
	testEncode(t, testObj)
}

func TestEncode_Interface(t *testing.T) {
	t.Parallel()
	testObj := StructWithInterface{
		Interface: &InterfaceImpl{
			interfaceImpl{
				A: 1,
				B: 2,
			},
		},
	}
	testEncode(t, testObj)
}

func TestEncode_Pointer(t *testing.T) {
	t.Parallel()
	ss := NewSimpleStruct()
	testObj := &ss
	testEncode(t, testObj)
}

func TestEncode_Optional(t *testing.T) {
	t.Parallel()
	testObj := StructWithOptionalField{Optional: nil}
	testEncode(t, testObj)
}

func TestEncode_EmbeddedStructs(t *testing.T) {
	t.Parallel()
	testObj := StructWithEmbeddedStructs{
		unexportedStruct: unexportedStruct{Foo: 1},
		ExportedStruct:   ExportedStruct{Bar: 2},
	}
	testEncode(t, testObj)
}

func TestEncode_Map(t *testing.T) {
	t.Parallel()
	testObj := Map{
		0: 2,
		1: 4,
	}
	testEncode(t, testObj, serix.WithTypeSettings(serix.TypeSettings{}.WithLengthPrefixType(mapLenType)))
}

func TestEncode_Serializable(t *testing.T) {
	t.Parallel()
	testObject := CustomSerializable(2)
	testEncode(t, testObject)
}

func TestEncode_SyntacticValidation(t *testing.T) {
	t.Parallel()
	testObj := ObjectForSyntacticValidation{}
	got, err := testAPI.Encode(ctx, testObj, serix.WithValidation())
	require.Nil(t, got)
	assert.ErrorIs(t, err, errSyntacticValidation)
}

func TestEncode_ArrayRules(t *testing.T) {
	t.Parallel()
	rules := &serix.ArrayRules{Min: 5}
	testObj := Bools{true, false, true, true}
	ts := serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint32).WithArrayRules(rules)
	got, err := testAPI.Encode(ctx, testObj, serix.WithValidation(), serix.WithTypeSettings(ts))
	require.Nil(t, got)
	require.ErrorIs(t, err, serializer.ErrArrayValidationMinElementsNotReached)
}

func testEncode(t testing.TB, testObj serializer.Serializable, opts ...serix.Option) {
	got, err := testAPI.Encode(ctx, testObj, opts...)
	require.NoError(t, err)
	expected, err := testObj.Serialize(defaultSeriMode, nil)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
