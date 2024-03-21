package serix_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

const defaultSeriMode = serializer.DeSeriModePerformValidation

var (
	testAPI            = serix.NewAPI()
	ctx                = context.Background()
	defaultArrayRules  = &serializer.ArrayRules{}
	defaultErrProducer = func(err error) error { return err }
	defaultWriteGuard  = func(seri serializer.Serializable) error { return nil }
)

func TestMinMax(t *testing.T) {
	type paras struct {
		api         *serix.API
		encodeInput any
		decodeInput []byte
	}

	type test struct {
		name  string
		paras paras
		error bool
	}
	tests := []test{
		{
			name: "ok - string in bounds",
			paras: func() paras {
				type example struct {
					Str string `serix:",minLen=5,maxLen=10,lenPrefix=uint8"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(0))))

				return paras{
					api:         api,
					encodeInput: &example{Str: "abcde"},
					decodeInput: []byte{0, 5, 97, 98, 99, 100, 101},
				}
			}(),
			error: false,
		},
		{
			name: "err - string out of bounds",
			paras: func() paras {
				type example struct {
					Str string `serix:",minLen=5,maxLen=10,lenPrefix=uint8"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(0))))

				return paras{
					api:         api,
					encodeInput: &example{Str: "abc"},
					decodeInput: []byte{0, 3, 97, 98, 99},
				}
			}(),
			error: true,
		},
		{
			name: "ok - slice in bounds",
			paras: func() paras {
				type example struct {
					Slice []byte `serix:",minLen=0,maxLen=10,lenPrefix=uint8"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(0))))

				return paras{
					api:         api,
					encodeInput: &example{Slice: []byte{1, 2, 3, 4, 5}},
					decodeInput: []byte{0, 5, 1, 2, 3, 4, 5},
				}
			}(),
			error: false,
		},
		{
			name: "err - slice out of bounds",
			paras: func() paras {
				type example struct {
					Slice []byte `serix:",minLen=0,maxLen=3,lenPrefix=uint8"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(0))))

				return paras{
					api:         api,
					encodeInput: &example{Slice: []byte{1, 2, 3, 4}},
					decodeInput: []byte{0, 4, 1, 2, 3, 4},
				}
			}(),
			error: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("encode", func(t *testing.T) {
				_, err := test.paras.api.Encode(context.Background(), test.paras.encodeInput, serix.WithValidation())
				if test.error {
					require.Error(t, err)

					return
				}
				require.NoError(t, err)
			})
			t.Run("decode", func(t *testing.T) {
				dest := reflect.New(reflect.TypeOf(test.paras.encodeInput).Elem()).Interface()
				_, err := test.paras.api.Decode(context.Background(), test.paras.decodeInput, dest, serix.WithValidation())
				if test.error {
					require.Error(t, err)

					return
				}
				require.NoError(t, err)
			})
		})
	}
}

type serializeTest struct {
	name    string
	source  any
	target  any
	size    int
	seriErr error
}

func (test *serializeTest) run(t *testing.T) {
	// binary serialize
	serixData, err := testAPI.Encode(context.Background(), test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "binary serialization failed")

		// we also need to check the json serialization
		_, err := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())
		require.ErrorIs(t, err, test.seriErr, "json serialization failed")

		return
	}
	require.NoError(t, err, "binary serialization failed")

	require.Equal(t, test.size, len(serixData))

	// binary deserialize
	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget)
	require.NoError(t, err, "binary deserialization failed")

	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget, "binary")

	// json serialize
	sourceJSON, err := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())
	require.NoError(t, err, "json serialization failed")

	// json deserialize
	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest, serix.WithValidation()), "json deserialization failed")

	require.EqualValues(t, test.source, jsonDest, "json")
}

func TestSerixSerializeMap(t *testing.T) {

	type MyMapTypeKey string
	type MyMapTypeValue string
	type MyMapType map[MyMapTypeKey]MyMapTypeValue
	type MapStruct struct {
		MyMap MyMapType `serix:",lenPrefix=uint8,minLen=2,maxLen=4"`
	}

	testAPI.RegisterTypeSettings(MyMapTypeKey(""), serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithMinLen(2).WithMaxLen(5))
	testAPI.RegisterTypeSettings(MyMapTypeValue(""), serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint32).WithMinLen(1).WithMaxLen(6))
	testAPI.RegisterTypeSettings(MapStruct{}, serix.TypeSettings{})

	tests := []serializeTest{
		{
			name: "ok",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
					"k2": "v2",
				},
			},
			target:  &MapStruct{},
			size:    21,
			seriErr: nil,
		},
		{
			name: "fail - not enough entries",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - too many entries",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
					"k4": "v4",
					"k5": "v5",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - key too short",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
					"k":  "v2",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - key too long",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1":     "v1",
					"k20000": "v2",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - value too short",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
					"k2": "",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - value too long",
			source: &MapStruct{
				MyMap: MyMapType{
					"k1": "v1",
					"k2": "v200000",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSerixSerializeString(t *testing.T) {

	type TestStruct struct {
		TestString string `serix:",lenPrefix=uint8"`
	}

	testAPI.RegisterTypeSettings(TestStruct{}, serix.TypeSettings{})

	tests := []serializeTest{
		{
			name: "ok",
			source: &TestStruct{
				TestString: "hello world!",
			},
			target:  &TestStruct{},
			size:    13,
			seriErr: nil,
		},
		{
			name: "fail - invalid utf8 string",
			source: &TestStruct{
				TestString: string([]byte{0xff, 0xfe, 0xfd}),
			},
			target:  &TestStruct{},
			size:    0,
			seriErr: serix.ErrNonUTF8String,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

type deSerializeTest struct {
	name      string
	source    any
	target    any
	size      int
	deSeriErr error
}

// convert all *orderedmap.OrderedMap to map[string]any
func convertOrderedMapToMap(m *orderedmap.OrderedMap) map[string]any {
	for k, v := range m.Values() {
		if v, ok := v.(*orderedmap.OrderedMap); ok {
			m.Set(k, convertOrderedMapToMap(v))
		}
	}

	return m.Values()
}

func (test *deSerializeTest) run(t *testing.T) {
	// binary serialize test data (without validation)
	serixData, err := testAPI.Encode(context.Background(), test.source)
	require.NoError(t, err, "binary serialization failed")

	// "map" serialize test data (without validation)
	// we don't use the json serialization here, because we want to test serix, and be able to inject malicous data
	serixMapData, err := testAPI.MapEncode(context.Background(), test.source)
	require.NoError(t, err, "map serialization failed")

	// convert all *orderedmap.OrderedMap in serixMapData to map[string]any
	serixMapDataUnordered := convertOrderedMapToMap(serixMapData)

	// binary deserialize
	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "binary deserialization failed")

		// we also need to check the "map" deserialization
		mapDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
		err := testAPI.MapDecode(context.Background(), serixMapDataUnordered, mapDest, serix.WithValidation())
		require.ErrorIs(t, err, test.deSeriErr, "map deserialization failed")

		return
	}
	require.NoError(t, err, "binary deserialization failed")

	require.Equal(t, test.size, bytesRead)
	require.EqualValues(t, test.source, serixTarget, "binary")

	// "map" deserialize
	mapDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.MapDecode(context.Background(), serixMapDataUnordered, mapDest, serix.WithValidation()), "map deserialization failed")

	require.EqualValues(t, test.source, mapDest, "map")
}

func TestSerixDeserializeMap(t *testing.T) {

	type MyMapTypeKey string
	type MyMapTypeValue string
	type MapStruct struct {
		MyMap map[MyMapTypeKey]MyMapTypeValue `serix:",lenPrefix=uint8,minLen=2,maxLen=4"`
	}

	testAPI.RegisterTypeSettings(MyMapTypeKey(""), serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithMinLen(2).WithMaxLen(5))
	testAPI.RegisterTypeSettings(MyMapTypeValue(""), serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint32).WithMinLen(1).WithMaxLen(6))
	testAPI.RegisterTypeSettings(MapStruct{}, serix.TypeSettings{})

	tests := []deSerializeTest{
		{
			name: "ok",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
					"k2": "v2",
				},
			},
			target:    &MapStruct{},
			size:      21,
			deSeriErr: nil,
		},
		{
			name: "fail - not enough entries",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - too many entries",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
					"k4": "v4",
					"k5": "v5",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - key too short",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
					"k":  "v2",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - key too long",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1":     "v1",
					"k20000": "v2",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - value too short",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
					"k2": "",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - value too long",
			source: &MapStruct{
				MyMap: map[MyMapTypeKey]MyMapTypeValue{
					"k1": "v1",
					"k2": "v200000",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSerixDeserializeString(t *testing.T) {

	type TestStruct struct {
		TestString string `serix:",lenPrefix=uint8"`
	}

	testAPI.RegisterTypeSettings(TestStruct{}, serix.TypeSettings{})

	tests := []deSerializeTest{
		{
			name: "ok",
			source: &TestStruct{
				TestString: "hello world!",
			},
			target:    &TestStruct{},
			size:      13,
			deSeriErr: nil,
		},
		{
			name: "fail - invalid utf8 string",
			source: &TestStruct{
				TestString: string([]byte{0xff, 0xfe, 0xfd}),
			},
			target:    &TestStruct{},
			size:      0,
			deSeriErr: serix.ErrNonUTF8String,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSerixFieldKeyString(t *testing.T) {
	type test struct {
		name   string
		source string
		target string
	}

	tests := []*test{
		{
			name:   "single char",
			source: "A",
			target: "a",
		},
		{
			name:   "all upper case",
			source: "MYTEST",
			target: "mYTEST",
		},
		{
			name:   "all lower case",
			source: "mytest",
			target: "mytest",
		},
		{
			name:   "mixed case",
			source: "MyTest",
			target: "myTest",
		},
		{
			name:   "mixed case with numbers",
			source: "MyTest123",
			target: "myTest123",
		},
		{
			name:   "mixed case with numbers and underscore",
			source: "MyTest_123",
			target: "myTest_123",
		},
		{
			name:   "mixed case with numbers and underscore and dash",
			source: "MyTest_123-",
			target: "myTest_123-",
		},
		{
			name:   "mixed case with special keyword 'id'",
			source: "MyTestID",
			target: "myTestId",
		},
		{
			name:   "mixed case with special keyword 'URL'",
			source: "MyTestURL",
			target: "myTestUrl",
		},
		{
			name:   "mixed case with special keyword 'ID' but lowercase",
			source: "MyTestid",
			target: "myTestid",
		},
		{
			name:   "mixed case with special keyword 'URL' but lowercase",
			source: "MyTesturl",
			target: "myTesturl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.target, serix.FieldKeyString(tt.source))
		})
	}
}

func TestSerixMustOccur(t *testing.T) {
	const (
		ShapeSquare    byte = 100
		ShapeRectangle byte = 101
		ShapeTriangle  byte = 102
	)

	type (
		Shape interface {
		}
		Square struct {
			Size uint8 `serix:""`
		}
		Rectangle struct {
			Size uint8 `serix:""`
		}
		Triangle struct {
			Size uint16 `serix:""`
		}
		Container struct {
			Shapes []Shape `serix:""`
		}
	)

	var shapesArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 10,
		MustOccur: serializer.TypePrefixes{
			uint32(ShapeSquare):    struct{}{},
			uint32(ShapeRectangle): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	must(testAPI.RegisterTypeSettings(Triangle{}, serix.TypeSettings{}.WithObjectType(uint8(ShapeTriangle))))
	must(testAPI.RegisterTypeSettings(Square{}, serix.TypeSettings{}.WithObjectType(uint8(ShapeSquare))))
	must(testAPI.RegisterTypeSettings(Rectangle{}, serix.TypeSettings{}.WithObjectType(uint8(ShapeRectangle))))
	must(testAPI.RegisterTypeSettings(Container{}, serix.TypeSettings{}.WithObjectType(uint8(5))))

	must(testAPI.RegisterTypeSettings([]Shape{},
		serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(shapesArrRules),
	))

	must(testAPI.RegisterInterfaceObjects((*Shape)(nil), (*Triangle)(nil)))
	must(testAPI.RegisterInterfaceObjects((*Shape)(nil), (*Square)(nil)))
	must(testAPI.RegisterInterfaceObjects((*Shape)(nil), (*Rectangle)(nil)))

	tests := []encodingTest{
		{
			name: "ok encoding",
			source: &Container{
				Shapes: []Shape{
					&Square{Size: 10},
					&Rectangle{Size: 5},
					&Triangle{Size: 3},
				},
			},
			target:  &Container{},
			seriErr: nil,
		},
		{
			name: "fail encoding - square must occur",
			source: &Container{
				Shapes: []Shape{
					&Rectangle{Size: 5},
					&Triangle{Size: 3},
				},
			},
			target:  &Container{},
			seriErr: serializer.ErrArrayValidationTypesNotOccurred,
		},
		{
			name: "fail encoding - square & rectangle must occur - empty slice",
			source: &Container{
				Shapes: []Shape{},
			},
			target:  &Container{},
			seriErr: serializer.ErrArrayValidationTypesNotOccurred,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}

	deSeriTests := []decodingTest{
		{
			name: "ok decoding",
			source: &Container{
				Shapes: []Shape{
					&Square{Size: 10},
					&Rectangle{Size: 5},
					&Triangle{Size: 3},
				},
			},
			target:    &Container{},
			deSeriErr: nil,
		},
		{
			name: "fail decoding - square must occur",
			source: &Container{
				Shapes: []Shape{
					&Rectangle{Size: 5},
					&Triangle{Size: 3},
				},
			},
			target:    &Container{},
			deSeriErr: serializer.ErrArrayValidationTypesNotOccurred,
		},
		{
			name: "fail decoding - square & rectangle must occur - empty slice",
			source: &Container{
				Shapes: []Shape{},
			},
			target:    &Container{},
			deSeriErr: serializer.ErrArrayValidationTypesNotOccurred,
		},
	}

	for _, tt := range deSeriTests {
		t.Run(tt.name, tt.run)
	}
}

func TestSerixInterfaceObjects(t *testing.T) {
	const (
		TestType1 byte = iota
		TestType2
		TestType3
	)

	type (
		Interface               interface{}
		Interfaces[T Interface] []T
		Interface1              interface{ Interface }
		Interface2              interface{ Interface }
		Interfaces1             = Interfaces[Interface1]
		Interfaces2             = Interfaces[Interface2]

		TestObject1 struct{}
		TestObject2 struct{}
		TestObject3 struct{}

		Container struct {
			Interfaces1 Interfaces1 `serix:""`
			Interfaces2 Interfaces2 `serix:""`
		}
	)

	registerTypes := func(api *serix.API) {
		must(api.RegisterTypeSettings(TestObject1{}, serix.TypeSettings{}.WithObjectType(uint8(TestType1))))
		must(api.RegisterTypeSettings(TestObject2{}, serix.TypeSettings{}.WithObjectType(uint8(TestType2))))
		must(api.RegisterTypeSettings(TestObject3{}, serix.TypeSettings{}.WithObjectType(uint8(TestType3))))

		must(api.RegisterTypeSettings(Interfaces1{}, serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte)))
		must(api.RegisterInterfaceObjects((*Interface1)(nil), (*TestObject1)(nil)))
		must(api.RegisterInterfaceObjects((*Interface1)(nil), (*TestObject2)(nil)))

		// we also register the objects in another interface, just to make sure that the interface objects are not mixed up
		must(api.RegisterTypeSettings(Interfaces2{}, serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte)))
		must(api.RegisterInterfaceObjects((*Interface2)(nil), (*TestObject1)(nil)))
		must(api.RegisterInterfaceObjects((*Interface2)(nil), (*TestObject2)(nil)))
		must(api.RegisterInterfaceObjects((*Interface2)(nil), (*TestObject3)(nil)))
	}

	registerTypes(testAPI)

	tests := []encodingTest{
		{
			name: "ok",
			source: &Container{
				Interfaces1: Interfaces1{
					&TestObject1{},
					&TestObject2{},
				},
				Interfaces2: Interfaces2{},
			},
			target:  &Container{},
			seriErr: nil,
		},
		{
			name: "fail - invalid object in Interfaces1",
			source: &Container{
				Interfaces1: Interfaces1{
					&TestObject1{},
					&TestObject2{},
					&TestObject3{},
				},
				Interfaces2: Interfaces2{},
			},
			target:  &Container{},
			seriErr: serix.ErrInterfaceUnderlyingTypeNotRegistered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}

	// we need to create a new API to create valid test data for the deserialization tests
	testVectorAPI := serix.NewAPI()
	registerTypes(testVectorAPI)

	// register TestObject3 to be able to create a valid test vector
	must(testVectorAPI.RegisterInterfaceObjects((*Interface1)(nil), (*TestObject3)(nil)))

	deSeriTests := []decodingTest{
		{
			name: "ok",
			source: &Container{
				Interfaces1: Interfaces1{
					&TestObject1{},
					&TestObject2{},
				},
				Interfaces2: Interfaces2{},
			},
			target:    &Container{},
			deSeriErr: nil,
		},
		{
			name: "fail - invalid types in Interfaces1",
			source: &Container{
				Interfaces1: Interfaces1{
					&TestObject1{},
					&TestObject2{},
					&TestObject3{},
				},
				Interfaces2: Interfaces2{},
			},
			target:    &Container{},
			deSeriErr: serix.ErrInterfaceUnderlyingTypeNotRegistered,
		},
	}

	for _, tt := range deSeriTests {
		t.Run(tt.name, func(t *testing.T) { tt.runWithTestVectorAPI(t, testVectorAPI) })
	}
}

type encodingTest struct {
	name    string
	source  any
	target  any
	seriErr error
}

func (test *encodingTest) run(t *testing.T) {
	serixData, err := testAPI.Encode(context.Background(), test.source, serix.WithValidation())
	jsonData, jsonErr := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())

	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr)
		require.ErrorIs(t, jsonErr, test.seriErr)

		return
	}
	require.NoError(t, err)
	require.NoError(t, jsonErr)

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget)

	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.JSONDecode(context.Background(), jsonData, jsonDest))

	require.EqualValues(t, test.source, jsonDest)
}

type decodingTest struct {
	name      string
	source    any
	target    any
	deSeriErr error
}

// runWithTestVectorAPI runs the decoding test with a new API to create valid test data.
func (test *decodingTest) runWithTestVectorAPI(t *testing.T, testVectorAPI *serix.API) {
	serixData, err := testVectorAPI.Encode(context.Background(), test.source)
	require.NoError(t, err)

	sourceJSON, err := testVectorAPI.JSONEncode(context.Background(), test.source)
	require.NoError(t, err)

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget, serix.WithValidation())

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	jsonErr := testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest, serix.WithValidation())

	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr)
		require.ErrorIs(t, jsonErr, test.deSeriErr)

		return
	}
	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	require.NoError(t, jsonErr)

	require.EqualValues(t, test.source, jsonDest)
}

func (test *decodingTest) run(t *testing.T) {
	// the normal tests uses the same API to create the test vectors as used for decoding
	test.runWithTestVectorAPI(t, testAPI)
}

func TestSerixOmitEmpty(t *testing.T) {
	type Numbers struct {
		Bytes []uint8 `serix:",omitempty"`
	}
	type omitEmptyTest struct {
		name        string
		expectEmpty bool
		source      Numbers
	}

	tests := []omitEmptyTest{
		{
			name:        "ok - slice empty",
			expectEmpty: true,
			source: Numbers{
				Bytes: []uint8{},
			},
		},
		{
			name:        "ok - nil slice",
			expectEmpty: true,
			source: Numbers{
				Bytes: nil,
			},
		},
		{
			name:        "ok - non-empty slice",
			expectEmpty: false,
			source: Numbers{
				Bytes: []uint8{0xff},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encodedJson, err := testAPI.JSONEncode(context.Background(), test.source)
			require.NoError(t, err)

			dec := json.NewDecoder(strings.NewReader(string(encodedJson)))
			var decoded Numbers
			err = dec.Decode(&decoded)
			require.NoError(t, err)

			if test.expectEmpty {
				require.Empty(t, decoded.Bytes)
			} else {
				require.NotEmpty(t, decoded.Bytes)
			}
		})
	}
}
