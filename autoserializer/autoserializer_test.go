package autoserializer_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autoserializer"
	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/datastructure/orderedmap"
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

func TestAutoserializer_LengthValidationCorrect(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
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

func TestAutoserializer_SerializeLengthValidationTooLong(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")
	orderedMapOrig.Set("fourth", "value")

	origSlice := sliceStruct{
		NumSlice: []int{1, 2, 3, 4, 5, 5},
	}
	_, err := sm.Serialize(origSlice)
	assert.Error(t, err)

	orderedMapOrig.Set("fifth", "value")

	origMap := sliceStruct{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1, 2, 3},
	}
	_, err = sm.Serialize(origMap)
	assert.Error(t, err)
}

func TestAutoserializer_SerializeLengthValidationTooShort(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
	orderedMapOrig := orderedmap.New()
	orderedMapOrig.Set("first", "value")

	origSlice := sliceStruct{
		NumSlice:   []int{1, 2, 3},
		OrderedMap: orderedMapOrig,
	}
	_, err := sm.Serialize(origSlice)
	assert.Error(t, err)
	fmt.Println(err)
	orderedMapOrig.Set("second", "value")
	orderedMapOrig.Set("third", "value")

	origMap := sliceStruct{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1},
	}
	_, err = sm.Serialize(origMap)
	assert.Error(t, err)
}

func TestAutoserializer_DeserializeLengthValidationTooLong(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
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
	assert.Error(t, err)
	fmt.Println(err)

	orderedMapOrig.Set("fifth", "value")

	origMap := sliceStructNovalidation{
		OrderedMap: orderedMapOrig,
		NumSlice:   []int{1, 2, 3},
	}
	bytesMap, err := sm.Serialize(origMap)
	assert.NoError(t, err)

	var restoredMap sliceStruct
	err = sm.Deserialize(&restoredMap, bytesMap)
	assert.Error(t, err)
}

func TestAutoserializer_DeserializeLengthValidationTooShort(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
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
	assert.Error(t, err)
	fmt.Println(err)

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
	assert.Error(t, err)
	fmt.Println(err)
}

func TestAutoserializer_Int64(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := int64(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Int32(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := int32(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Int16(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := int16(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int16
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Int8(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := int8(100)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored int8
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Uint64(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := uint64(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Uint32(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := uint32(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Uint16(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := uint16(1337)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint16
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Uint8(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := uint8(137)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored uint8
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Float32(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := float32(3.14)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored float32
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Float64(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := float64(3.14)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored float64
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Bool(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := true
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored bool
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, restored)
}

func TestAutoserializer_Byte(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := byte(100)
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored byte
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_String(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := "test string value"
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored string
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_Array(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := [2]string{"test string value", "test string value 2"}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored [2]string
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}
func TestAutoserializer_StructArray(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := [3]TestImpl1{{1}, {2}, {3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored [3]TestImpl1
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestAutoserializer_InterfaceArray(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
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

func TestAutoserializer_Slice(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored []int
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestAutoserializer_StructSlice(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := []TestImpl1{{1}, {2}, {3}}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored []TestImpl1
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestAutoserializer_InterfaceSlice(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
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

func TestAutoserializer_EmptySlice(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

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
func TestAutoserializer_NilSlice(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

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

func TestAutoserializer_SliceOfEmptyStructs(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

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

func TestAutoserializer_Time(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := time.Now()
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	var restored time.Time
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, orig.Equal(restored))
}

func TestAutoserializer_BinaryMarshaler(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig, err := url.Parse("https://pkg.go.dev/encoding#BinaryMarshaler")
	assert.NoError(t, err)

	bytes, err := sm.Serialize(orig)

	assert.NoError(t, err)

	restored := &url.URL{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, orig, restored)
}

func TestAutoserializer_ZeroTime(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	var defaultTime time.Time
	bytes, err := sm.Serialize(defaultTime)
	assert.NoError(t, err)

	restored := time.Now()
	err = sm.Deserialize(&restored, bytes)
	assert.NoError(t, err)
	assert.True(t, restored.IsZero())
}

func TestAutoserializer_OrderedMap(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
	orig := orderedmap.New()
	orig.Set("first", "value")
	orig.Set("second", "value")
	orig.Set("third", "value")
	bytes, err := sm.Serialize(orig)
	fmt.Println(bytes)
	assert.NoError(t, err)
	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 3)
}

func TestAutoserializer_EmptyOrderedMap(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	orig := orderedmap.New()
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	fmt.Println(bytes)
	restored := orderedmap.New()
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Equal(t, restored.Size(), 0)
}

func TestAutoserializer_OrderedMapWithStruct(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
	sm.RegisterType(TestImpl1{})
	sm.RegisterType(TestImpl2{})
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

func TestAutoserializer_OrderedMapWithInterface(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
	sm.RegisterType("")
	sm.RegisterType(TestImpl1{})
	sm.RegisterType(TestImpl2{})
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

func TestAutoserializer_PointerToStruct(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := &TestImpl1{12}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := &TestImpl1{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, restored)
}

func TestAutoserializer_NilPointer(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := &InnerPointer{}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)

	restored := &InnerPointer{}
	err = sm.Deserialize(restored, bytes)
	assert.NoError(t, err)
	assert.Nil(t, restored.Custom)
	assert.Nil(t, orig.Custom)

}

func TestAutoserializer_PointerToInterface(t *testing.T) {
	sm := autoserializer.NewSerializationManager()
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

func TestAutoserializer_TooManyBytes(t *testing.T) {
	sm := autoserializer.NewSerializationManager()

	orig := &TestImpl1{12}
	bytes, err := sm.Serialize(orig)
	assert.NoError(t, err)
	tooManyBytes := byteutils.ConcatBytes(bytes, []byte{0x00, 0x01, 0x02})
	var restored *TestImpl1
	err = sm.Deserialize(&restored, tooManyBytes)
	assert.Error(t, err)
}

func TestAutoserializer_Struct(t *testing.T) {
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
	sm := autoserializer.NewSerializationManager()
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

func BenchmarkMessageToBytesAutoserializer(b *testing.B) {
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
	sm := autoserializer.NewSerializationManager()
	_ = sm.RegisterType("")

	_ = sm.RegisterType(TestImpl1{})

	_ = sm.RegisterType(TestImpl2{})
	var bytes []byte
	fmt.Println(b.N)

	var err error
	for n := 0; err == nil && n < b.N; n++ {
		bytes, err = sm.Serialize(orig)
	}
	fmt.Println(len(bytes))
	if err != nil {
		b.Error(err)
	}
	result = bytes
}
