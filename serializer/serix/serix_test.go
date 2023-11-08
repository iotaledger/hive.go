//nolint:scopelint // we don't care about these linters in test cases
package serix_test

import (
	"context"
	"reflect"
	"testing"

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
		require.ErrorIs(t, err, test.seriErr)

		// we also need to check the json serialization
		_, err := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())
		require.ErrorIs(t, err, test.seriErr)

		return
	}
	require.NoError(t, err)

	require.Equal(t, test.size, len(serixData))

	// binary deserialize
	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget)
	require.NoError(t, err)

	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	// json serialize
	sourceJSON, err := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())
	require.NoError(t, err)

	// json deserialize
	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest, serix.WithValidation()))

	require.EqualValues(t, test.source, jsonDest)
}

func TestSerixMapSerialize(t *testing.T) {

	type MyMapType map[string]string

	type MapStruct struct {
		MyMap MyMapType `serix:",lenPrefix=uint8,minLen=2,maxLen=4,maxByteSize=50,mapKeyLenPrefix=uint16,mapKeyMinLen=2,mapKeyMaxLen=5,mapValueLenPrefix=uint32,mapValueMinLen=1,mapValueMaxLen=6"`
	}
	testAPI.RegisterTypeSettings(MapStruct{}, serix.TypeSettings{})

	tests := []serializeTest{
		{
			name: "ok",
			source: &MapStruct{
				MyMap: map[string]string{
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
				MyMap: map[string]string{
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
				MyMap: map[string]string{
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
			name: "fail - too big",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1000",
					"k2": "v2000",
					"k3": "v3000",
					"k4": "v4000",
				},
			},
			target:  &MapStruct{},
			size:    0,
			seriErr: serix.ErrValidationMaxBytesExceeded,
		},
		{
			name: "fail - key too short",
			source: &MapStruct{
				MyMap: map[string]string{
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
				MyMap: map[string]string{
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
				MyMap: map[string]string{
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
				MyMap: map[string]string{
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

type deSerializeTest struct {
	name      string
	source    any
	target    any
	size      int
	deSeriErr error
}

func (test *deSerializeTest) run(t *testing.T) {
	// binary serialize test data
	serixData, err := testAPI.Encode(context.Background(), test.source, serix.WithValidation())
	require.NoError(t, err)

	// json serialize test data
	sourceJSON, err := testAPI.JSONEncode(context.Background(), test.source, serix.WithValidation())
	require.NoError(t, err)

	// binary deserialize
	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr)

		// we also need to check the json deserialization
		jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
		err := testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest, serix.WithValidation())
		require.ErrorIs(t, err, test.deSeriErr)

		return
	}
	require.NoError(t, err)

	require.Equal(t, test.size, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	// json deserialize
	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest, serix.WithValidation()))

	require.EqualValues(t, test.source, jsonDest)
}

func TestSerixMapDeserialize(t *testing.T) {

	type MyMapType map[string]string

	// used to create test data
	type TestVectorMapStruct struct {
		MyMap MyMapType `serix:",lenPrefix=uint8,minLen=1,maxLen=5,maxByteSize=100,mapKeyLenPrefix=uint16,mapKeyMinLen=1,mapKeyMaxLen=7,mapValueLenPrefix=uint32,mapValueMinLen=0,mapValueMaxLen=10"`
	}
	testAPI.RegisterTypeSettings(TestVectorMapStruct{}, serix.TypeSettings{})

	type MapStruct struct {
		MyMap MyMapType `serix:",lenPrefix=uint8,minLen=2,maxLen=4,maxByteSize=50,mapKeyLenPrefix=uint16,mapKeyMinLen=2,mapKeyMaxLen=5,mapValueLenPrefix=uint32,mapValueMinLen=1,mapValueMaxLen=6"`
	}
	testAPI.RegisterTypeSettings(MapStruct{}, serix.TypeSettings{})

	tests := []deSerializeTest{
		{
			name: "ok",
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
					"k1": "v1",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			name: "fail - too many entries",
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
			name: "fail - too big",
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
					"k1": "v1000",
					"k2": "v2000",
					"k3": "v3000",
					"k4": "v4000",
				},
			},
			target:    &MapStruct{},
			size:      0,
			deSeriErr: serix.ErrValidationMaxBytesExceeded,
		},
		{
			name: "fail - key too short",
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
			source: &TestVectorMapStruct{
				MyMap: map[string]string{
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
