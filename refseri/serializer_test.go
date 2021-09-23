package refseri_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/datastructure/orderedmap"
	"github.com/iotaledger/hive.go/refseri"
	"github.com/stretchr/testify/assert"
)

type Test interface {
	Test() int
}

// *TestImpl1 implements Test
type TestImpl1 struct {
	Val int `serialize:"true"`
}

func (m TestImpl1) Test() int {
	return 1
}

// *TestImpl2 implements Test
type TestImpl2 struct {
	Val int `serialize:"true"`
}

func (m TestImpl2) Test() int {
	return 3
}

type CustomType []string

type InnerStruct struct {
	Custom CustomType `serialize:"true"`
}

type InnerPointer struct {
	Custom *TestImpl1 `serialize:"true" allowNil:"true"`
}

type sampleStruct struct {
	Num1                      int64                  `serialize:"true"`
	Num2                      int32                  `serialize:"true"`
	Num3                      int16                  `serialize:"true"`
	Num4                      int8                   `serialize:"true"`
	Num5                      uint64                 `serialize:"true"`
	Num6                      uint32                 `serialize:"true"`
	Num7                      uint16                 `serialize:"true"`
	Num8                      uint8                  `serialize:"true"`
	Float1                    float64                `serialize:"true"`
	Float2                    float32                `serialize:"true"`
	Boolean                   bool                   `serialize:"true"`
	StringSlice               []string               `serialize:"true"`
	NumSlice                  []int64                `serialize:"true"`
	ByteSlice                 []byte                 `serialize:"true" lenPrefixBytes:"4" `
	StringArray               [32]string             `serialize:"true"`
	NumArray                  [32]int64              `serialize:"true"`
	ByteArray                 [32]byte               `serialize:"true"`
	StructType                InnerStruct            `serialize:"true"`
	PointerStructType         *InnerStruct           `serialize:"true"`
	SlicePointerStructType    []*InnerStruct         `serialize:"true" lenPrefixBytes:"1"`
	SliceStructType           []InnerStruct          `serialize:"true"`
	OrderedMap                *orderedmap.OrderedMap `serialize:"true" lenPrefixBytes:"2"`
	NilOrderedMap             *orderedmap.OrderedMap `serialize:"true" lenPrefixBytes:"2" allowNil:"true"`
	Time                      time.Time              `serialize:"true"`
	InterfaceType             Test                   `serialize:"true"`
	PointerInterfaceType      *Test                  `serialize:"true"`
	SliceInterfaceType        []Test                 `serialize:"true" lenPrefixBytes:"2"`
	SlicePointerInterfaceType []*Test                `serialize:"true"`
	NilPointerType            *InnerStruct           `serialize:"true" allowNil:"true"`
	NilPointerInterfaceType   *Test                  `serialize:"true" allowNil:"true"`
	BinaryMarshallerType      *url.URL               `serialize:"true"`
}

type sliceStruct struct {
	NumSlice   []int                  `serialize:"true" lenPrefixBytes:"4" minLen:"2" maxLen:"4"`
	OrderedMap *orderedmap.OrderedMap `serialize:"true" lenPrefixBytes:"2" minLen:"3" maxLen:"4"`
}

type sliceStructNovalidation struct {
	NumSlice   []int                  `serialize:"true" lenPrefixBytes:"4"`
	OrderedMap *orderedmap.OrderedMap `serialize:"true" lenPrefixBytes:"2"`
}

type sliceOrderStruct struct {
	NumSlice    []int    `serialize:"true" lexicalOrder:"true"`
	StringSlice []string `serialize:"true" lexicalOrder:"true"`
}

type sliceDuplicateStruct struct {
	NumSlice    []int    `serialize:"true" noDuplicates:"true"`
	StringSlice []string `serialize:"true" noDuplicates:"true"`
}
type sliceDuplicateOrderStruct struct {
	NumSlice    []int    `serialize:"true" noDuplicates:"true" lexicalOrder:"true"`
	StringSlice []string `serialize:"true" noDuplicates:"true" lexicalOrder:"true"`
}
type sliceStructNoOrder struct {
	NumSlice    []int    `serialize:"true"`
	StringSlice []string `serialize:"true"`
}

func TestReflectionSerializer_NoDuplicates(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := sliceDuplicateStruct{
		NumSlice:    []int{2, 1, 1, 2, 15, 15, -31},
		StringSlice: []string{"zebra", "elephant", "zebra", "alpaca", "lion", "alpaca", "elephant"},
	}
	origDups := sliceStructNoOrder{
		NumSlice:    []int{2, 1, 1, 2, 15, 15, -31},
		StringSlice: []string{"zebra", "elephant", "zebra", "alpaca", "lion", "alpaca", "elephant"},
	}

	_, err := sm.Serialize(orig)
	assert.ErrorIs(t, err, refseri.ErrNoDuplicatesViolated)

	bytesDups, err := sm.Serialize(origDups)
	assert.NoError(t, err)

	var restoredDups sliceDuplicateStruct

	// restoring bytes with unordered slice into ordered struct
	err = sm.Deserialize(&restoredDups, bytesDups)
	assert.ErrorIs(t, err, refseri.ErrNoDuplicatesViolated)
}

func TestReflectionSerializer_EnforceLexicalOrdering(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := sliceOrderStruct{
		NumSlice:    []int{2, 1, 0, -100, 15, -31},
		StringSlice: []string{"zebra", "elephant", "alpaca", "lion"},
	}

	origNoOrder := sliceStructNoOrder{
		NumSlice:    []int{2, 1, 0, -100, 15, -31},
		StringSlice: []string{"zebra", "elephant", "alpaca", "lion"},
	}

	_, err := sm.Serialize(orig)
	assert.ErrorIs(t, err, refseri.ErrLexicalOrderViolated)

	bytesNoOrder, err := sm.Serialize(origNoOrder)
	assert.NoError(t, err)

	// the same object with and without ordering tags should be serialized to different binary forms
	var restoredNoOrder sliceOrderStruct

	// restoring bytes with unordered slice into ordered struct
	err = sm.Deserialize(&restoredNoOrder, bytesNoOrder)
	assert.ErrorIs(t, err, refseri.ErrLexicalOrderViolated)
}

func TestReflectionSerializer_FixLexicalOrdering(t *testing.T) {
	t.Skip()
	sm := refseri.NewSerializationManager()
	expected := sliceOrderStruct{
		NumSlice:    []int{0, 1, 2, 15, -100, -31},
		StringSlice: []string{"lion", "zebra", "alpaca", "elephant"},
	}

	orig := sliceOrderStruct{
		NumSlice:    []int{2, 1, 0, -100, 15, -31},
		StringSlice: []string{"zebra", "elephant", "alpaca", "lion"},
	}

	origNoOrder := sliceStructNoOrder{
		NumSlice:    []int{2, 1, 0, -100, 15, -31},
		StringSlice: []string{"zebra", "elephant", "alpaca", "lion"},
	}

	bytesExpected, err := sm.Serialize(expected)
	assert.NoError(t, err)

	bytesOrder, err := sm.Serialize(orig)
	assert.NoError(t, err)

	bytesNoOrder, err := sm.Serialize(origNoOrder)
	assert.NoError(t, err)

	// the same object with and without ordering tags should be serialized to different binary forms
	assert.NotEqual(t, bytesOrder, bytesNoOrder)
	assert.Equal(t, bytesExpected, bytesOrder)
	var restoredOrderRaw sliceStructNoOrder
	var restoredOrder sliceOrderStruct
	var restoredNoOrder sliceOrderStruct

	// restore bytes into structure without order checking and see if the bytes are correctly serialized
	err = sm.Deserialize(&restoredOrderRaw, bytesOrder)
	assert.NoError(t, err)
	assert.NotEqual(t, restoredOrderRaw, orig)

	err = sm.Deserialize(&restoredOrder, bytesOrder)
	assert.NoError(t, err)
	// restoring bytes with unordered slice into ordered struct
	err = sm.Deserialize(&restoredNoOrder, bytesNoOrder)
	assert.NoError(t, err)

	// both should be deserialized into exactly the same structure
	assert.EqualValues(t, expected, restoredNoOrder)
	assert.EqualValues(t, expected, restoredOrderRaw)
	assert.EqualValues(t, expected, restoredOrder)
}

func TestReflectionSerializer_SkipDuplicates(t *testing.T) {
	t.Skip()

	sm := refseri.NewSerializationManager()
	expected := sliceOrderStruct{
		NumSlice:    []int{1, 2, 15, -31},
		StringSlice: []string{"lion", "zebra", "alpaca", "elephant"},
	}

	orig := sliceOrderStruct{
		NumSlice:    []int{2, 1, 1, 2, 15, 15, -31},
		StringSlice: []string{"zebra", "elephant", "zebra", "alpaca", "lion", "alpaca", "elephant"},
	}
	origDups := sliceStructNoOrder{
		NumSlice:    []int{2, 1, 1, 2, 15, 15, -31},
		StringSlice: []string{"zebra", "elephant", "zebra", "alpaca", "lion", "alpaca", "elephant"},
	}

	bytesNoDups, err := sm.Serialize(orig)
	assert.NoError(t, err)

	bytesDups, err := sm.Serialize(origDups)
	assert.NoError(t, err)

	bytesExpected, err := sm.Serialize(expected)
	assert.NoError(t, err)
	// the same object with and without ordering tags should be serialized to different binary forms
	assert.NotEqual(t, bytesNoDups, bytesDups)
	assert.Equal(t, bytesExpected, bytesNoDups)

	var restoredNoDupsRaw sliceStructNoOrder
	var restoredNoDups sliceOrderStruct
	var restoredDups sliceOrderStruct

	err = sm.Deserialize(&restoredNoDupsRaw, bytesNoDups)
	assert.NoError(t, err)

	err = sm.Deserialize(&restoredNoDups, bytesNoDups)
	assert.NoError(t, err)
	// restoring bytes with unordered slice into ordered struct
	err = sm.Deserialize(&restoredDups, bytesDups)
	assert.NoError(t, err)

	// both should be deserialized into exactly the same structure
	assert.EqualValues(t, expected, restoredDups)
	assert.EqualValues(t, expected, restoredNoDupsRaw)
	assert.EqualValues(t, expected, restoredNoDups)
}

func TestReflectionSerializer_LengthValidationCorrect(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)
	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")
	orderedMapOrig.Set("fourth", "value")
	orig := sliceStruct{
		NumSlice:   []int{1, 2},
		OrderedMap: orderedMapOrig,
	}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored sliceStruct
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_SerializeLengthValidationTooLong(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)

	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")
	orderedMapOrig.Set("fourth", "value")

	origSlice := sliceStruct{
		NumSlice: []int{1, 2, 3, 4, 5, 5},
	}
	_, err = sm.Serialize(origSlice)
	assert.ErrorIs(t, err, refseri.ErrSliceMaxLength)

	orderedMapOrig.Set("fifth", "value")

	origMap := sliceStruct{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1, 2, 3},
	}
	_, err = sm.Serialize(origMap)
	assert.ErrorIs(t, err, refseri.ErrSliceMaxLength)
}

func TestReflectionSerializer_SerializeLengthValidationTooShort(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)

	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")

	origSlice := sliceStruct{
		NumSlice:   []int{1, 2, 3},
		OrderedMap: orderedMapOrig,
	}
	_, err = sm.Serialize(origSlice)
	assert.ErrorIs(t, err, refseri.ErrSliceMinLength)
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")

	origMap := sliceStruct{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1},
	}
	_, err = sm.Serialize(origMap)
	assert.ErrorIs(t, err, refseri.ErrSliceMinLength)
}

func TestReflectionSerializer_DeserializeLengthValidationTooLong(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)

	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")
	orderedMapOrig.Set("fourth", "value")

	origSlice := sliceStructNovalidation{
		NumSlice:   []int{1, 2, 3, 4, 5, 5},
		OrderedMap: orderedMapOrig,
	}
	bytesSlice, err := sm.Serialize(origSlice)
	assert.NoError(t, err)

	var restoredSlice sliceStruct
	err = sm.Deserialize(&restoredSlice, bytesSlice)
	assert.ErrorIs(t, err, refseri.ErrSliceMaxLength)

	orderedMapOrig.Set("fifth", "value")

	origMap := sliceStructNovalidation{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1, 2, 3},
	}
	bytesMap, err := sm.Serialize(origMap)
	assert.NoError(t, err)

	var restoredMap sliceStruct
	err = sm.Deserialize(&restoredMap, bytesMap)
	assert.ErrorIs(t, err, refseri.ErrSliceMaxLength)
}

func TestReflectionSerializer_DeserializeLengthValidationTooShort(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)

	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")

	origMap := sliceStructNovalidation{
		NumSlice:   []int{1, 2, 3},
		OrderedMap: orderedMapOrig,
	}
	bytesMap, err := sm.Serialize(origMap)
	assert.NoError(t, err)

	var restoredMap sliceStruct
	err = sm.Deserialize(&restoredMap, bytesMap)
	assert.ErrorIs(t, err, refseri.ErrSliceMinLength)

	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")

	origSlice := sliceStructNovalidation{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1},
	}
	bytesSlice, err := sm.Serialize(origSlice)
	assert.NoError(t, err)

	var restoredSlice sliceStruct
	err = sm.Deserialize(&restoredSlice, bytesSlice)
	assert.ErrorIs(t, err, refseri.ErrSliceMinLength)
}

func TestReflectionSerializer_Int64(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := int64(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Int32(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := int32(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Int16(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := int16(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int16
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Int8(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := int8(100)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int8
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Uint64(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := uint64(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Uint32(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := uint32(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Uint16(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := uint16(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint16
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Uint8(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := uint8(137)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint8
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Float32(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := float32(3.14)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored float32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Float64(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := float64(3.14)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored float64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Bool(t *testing.T) {
	sm := refseri.NewSerializationManager()

	bytes, err := sm.Serialize(true)
	assert.NoError(t, err)

	var restored bool
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, restored)
}

func TestReflectionSerializer_Byte(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := byte(100)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored byte
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_String(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := "test string value"
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored string
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_Array(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := [2]string{"test string value", "test string value 2"}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored [2]string
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_StructArray(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := [3]TestImpl1{{1}, {2}, {3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored [3]TestImpl1
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_InterfaceArray(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)

	err = sm.RegisterType(TestImpl2{})
	assert.NoError(t, err)

	orig := [3]Test{TestImpl1{1}, TestImpl2{2}, TestImpl1{3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored [3]Test
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_Slice(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored []int
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_StructSlice(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := []TestImpl1{{1}, {2}, {3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored []TestImpl1
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_InterfaceSlice(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)

	err = sm.RegisterType(TestImpl2{})
	assert.NoError(t, err)

	orig := []Test{TestImpl1{1}, TestImpl2{2}, TestImpl1{3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored []Test
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_EmptySlice(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := make([]string, 0)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	expected := []byte{0}
	assert.Equal(t, expected, bytes)
	var restored []int
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Len(t, restored, 0)
}

func TestReflectionSerializer_NilSlice(t *testing.T) {
	sm := refseri.NewSerializationManager()

	var orig []int
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	expected := []byte{0}
	assert.Equal(t, expected, bytes)
	var restored []int

	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Len(t, restored, 0)
}

func TestReflectionSerializer_SliceOfEmptyStructs(t *testing.T) {
	sm := refseri.NewSerializationManager()

	type emptyStruct struct{}

	orig := make([]emptyStruct, 255)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	expected := []byte{0xff}
	assert.Equal(t, expected, bytes)
	var restored []emptyStruct
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Len(t, restored, 255)
}

func TestReflectionSerializer_Time(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := time.Now()
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored time.Time
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, orig.Equal(restored))
}

func TestReflectionSerializer_BinaryMarshaler(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig, err := url.Parse("https://pkg.go.dev/encoding#BinaryMarshaler")
	assert.NoError(t, err)

	bytes, err := sm.Serialize(orig)

	assert.NoError(t, err)

	restored := &url.URL{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestReflectionSerializer_ZeroTime(t *testing.T) {
	sm := refseri.NewSerializationManager()

	var defaultTime time.Time
	bytes, err := sm.Serialize(defaultTime)
	assert.NoError(t, err)

	restored := time.Now()
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, restored.IsZero())
}

func TestReflectionSerializer_OrderedMap(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)

	orig := orderedmap.New()
	orig.Set("first", "value")
	orig.Set("second", "value")
	orig.Set("third", "value")
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 3)
}

func TestReflectionSerializer_EmptyOrderedMap(t *testing.T) {
	sm := refseri.NewSerializationManager()
	orig := orderedmap.New()
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 0)
}

func TestReflectionSerializer_OrderedMapWithStruct(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)
	err = sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)
	err = sm.RegisterType(TestImpl2{})
	assert.NoError(t, err)
	orig := orderedmap.New()
	orig.Set("first", TestImpl2{2})
	orig.Set("second", TestImpl2{33})
	orig.Set("third", TestImpl2{12})
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 3)
	first, exists := restored.Get("first")
	assert.True(t, exists)
	assert.Equal(t, first, TestImpl2{2})
	second, exists := restored.Get("second")
	assert.True(t, exists)
	assert.Equal(t, second, TestImpl2{33})
	third, exists := restored.Get("third")
	assert.True(t, exists)
	assert.Equal(t, third, TestImpl2{12})
}

func TestReflectionSerializer_OrderedMapWithInterface(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType("")
	assert.NoError(t, err)
	err = sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)
	err = sm.RegisterType(TestImpl2{})
	assert.NoError(t, err)
	orig := orderedmap.New()
	orig.Set("first", Test(TestImpl1{2}))
	orig.Set("second", Test(TestImpl2{33}))
	orig.Set("third", Test(TestImpl1{12}))
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 3)
	first, exists := restored.Get("first")
	assert.True(t, exists)
	assert.Equal(t, first, Test(TestImpl1{2}))
	second, exists := restored.Get("second")
	assert.True(t, exists)
	assert.Equal(t, second, Test(TestImpl2{33}))
	third, exists := restored.Get("third")
	assert.True(t, exists)
	assert.Equal(t, third, Test(TestImpl1{12}))
}

func TestReflectionSerializer_PointerToStruct(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := &TestImpl1{12}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := &TestImpl1{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestReflectionSerializer_NilPointer(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := &InnerPointer{}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := &InnerPointer{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Nil(t, restored.Custom)
	assert.Nil(t, orig.Custom)
}

func TestReflectionSerializer_PointerToInterface(t *testing.T) {
	sm := refseri.NewSerializationManager()
	err := sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)

	orig := Test(TestImpl1{12})
	t2 := &orig
	bytes, err := sm.Serialize(t2)
	assert.NoError(t, err)

	restored := Test(TestImpl1{})
	t3 := &restored
	err = sm.Deserialize(t3, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, t2, t3)
}

func TestReflectionSerializer_TooManyBytes(t *testing.T) {
	sm := refseri.NewSerializationManager()

	orig := &TestImpl1{12}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	tooManyBytes := byteutils.ConcatBytes(bytes, []byte{0x00, 0x01, 0x02})
	var restored *TestImpl1
	err = sm.Deserialize(&restored, tooManyBytes)
	assert.ErrorIs(t, err, refseri.ErrNotAllBytesRead)
}

func TestReflectionSerializer_Struct(t *testing.T) {
	interfacePointer := Test(TestImpl2{Val: 10})
	sampleOrderedMap := orderedmap.New()
	sampleOrderedMap.Set("first", TestImpl2{2})
	sampleOrderedMap.Set("second", TestImpl2{33})
	sampleOrderedMap.Set("third", TestImpl2{12})
	urlValue, err := url.Parse("https://pkg.go.dev/encoding#BinaryMarshaler")
	assert.NoError(t, err)
	orig := sampleStruct{
		Num1:                      1,
		Num2:                      2,
		Num3:                      3,
		Num4:                      4,
		Num5:                      5,
		Num6:                      6,
		Num7:                      7,
		Num8:                      8,
		Float1:                    1.23,
		Float2:                    2.65,
		Boolean:                   true,
		StringSlice:               []string{"one", "two", "three", "four"},
		NumSlice:                  []int64{1, 2, 3, 4},
		ByteSlice:                 []byte{1, 2, 3, 4},
		StringArray:               [32]string{"one", "two", "three"},
		NumArray:                  [32]int64{1, 2, 3},
		ByteArray:                 [32]byte{1, 2, 3},
		StructType:                InnerStruct{[]string{"one", "two", "three", "four"}},
		PointerStructType:         &InnerStruct{[]string{"one", "two", "three", "four"}},
		SlicePointerStructType:    []*InnerStruct{{[]string{"one", "two", "three", "four"}}},
		SliceStructType:           []InnerStruct{{[]string{"one", "two", "three", "four"}}},
		InterfaceType:             TestImpl1{Val: 10},
		PointerInterfaceType:      &interfacePointer,
		SliceInterfaceType:        []Test{TestImpl1{Val: 10}, TestImpl1{Val: 12}, TestImpl2{Val: 13}, TestImpl2{Val: 15}},
		SlicePointerInterfaceType: []*Test{&interfacePointer},
		Time:                      time.Now(),
		OrderedMap:                sampleOrderedMap,
		BinaryMarshallerType:      urlValue,
	}
	sm := refseri.NewSerializationManager()
	err = sm.RegisterType("")
	assert.NoError(t, err)

	err = sm.RegisterType(TestImpl1{})
	assert.NoError(t, err)

	err = sm.RegisterType(TestImpl2{})
	assert.NoError(t, err)

	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := sampleStruct{}
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, orig.Time.Equal(restored.Time))

	assert.Equal(t, restored.OrderedMap.Size(), 3)
	first, exists := restored.OrderedMap.Get("first")
	assert.True(t, exists)
	assert.Equal(t, first, TestImpl2{2})
	second, exists := restored.OrderedMap.Get("second")
	assert.True(t, exists)
	assert.Equal(t, second, TestImpl2{33})
	third, exists := restored.OrderedMap.Get("third")
	assert.True(t, exists)
	assert.Equal(t, third, TestImpl2{12})
	assert.Nil(t, restored.NilPointerType)
	assert.Nil(t, restored.NilPointerInterfaceType)
	assert.Nil(t, restored.NilOrderedMap)
	assert.Equal(t, orig.BinaryMarshallerType, restored.BinaryMarshallerType)

	// time is correctly restored, remove it from the struct to use automatic equality check
	restored.Time = time.Time{}
	orig.Time = time.Time{}
	assert.EqualValues(t, orig, restored)
	assert.True(t, reflect.DeepEqual(orig, restored))
}

var result []byte

func BenchmarkMessageToBytesReflectionSerializer(b *testing.B) {
	interfacePointer := Test(TestImpl2{Val: 10})
	sampleOrderedMap := orderedmap.New()
	sampleOrderedMap.Set("first", TestImpl2{2})
	sampleOrderedMap.Set("second", TestImpl2{33})
	sampleOrderedMap.Set("third", TestImpl2{12})

	orig := sampleStruct{
		Num1:                      1,
		Num2:                      2,
		Num3:                      3,
		Num4:                      4,
		Num5:                      5,
		Num6:                      6,
		Num7:                      7,
		Num8:                      8,
		Float1:                    1.23,
		Float2:                    2.65,
		Boolean:                   true,
		StringSlice:               []string{"one", "two", "three", "four"},
		NumSlice:                  []int64{1, 2, 3, 4},
		ByteSlice:                 []byte{1, 2, 3, 4},
		StringArray:               [32]string{"one", "two", "three"},
		NumArray:                  [32]int64{1, 2, 3},
		ByteArray:                 [32]byte{1, 2, 3},
		StructType:                InnerStruct{[]string{"one", "two", "three", "four"}},
		PointerStructType:         &InnerStruct{[]string{"one", "two", "three", "four"}},
		SlicePointerStructType:    []*InnerStruct{{[]string{"one", "two", "three", "four"}}},
		SliceStructType:           []InnerStruct{{[]string{"one", "two", "three", "four"}}},
		InterfaceType:             TestImpl1{Val: 10},
		PointerInterfaceType:      &interfacePointer,
		SliceInterfaceType:        []Test{TestImpl1{Val: 10}, TestImpl1{Val: 12}, TestImpl2{Val: 13}, TestImpl2{Val: 15}},
		SlicePointerInterfaceType: []*Test{&interfacePointer},
		Time:                      time.Now(),
		OrderedMap:                sampleOrderedMap,
	}
	sm := refseri.NewSerializationManager()
	_ = sm.RegisterType("")

	_ = sm.RegisterType(TestImpl1{})

	_ = sm.RegisterType(TestImpl2{})
	var bytes []byte

	var err error
	for n := 0; err == nil && n < b.N; n++ {
		bytes, err = sm.Serialize(orig)
	}
	if err != nil {
		b.Error(err)
	}
	result = bytes
}
