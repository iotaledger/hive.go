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
					Str string `serix:"0,minLen=5,maxLen=10,lengthPrefixType=uint8"`
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
					Str string `serix:"0,minLen=5,maxLen=10,lengthPrefixType=uint8"`
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
					Slice []byte `serix:"0,minLen=0,maxLen=10,lengthPrefixType=uint8"`
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
					Slice []byte `serix:"0,minLen=0,maxLen=3,lengthPrefixType=uint8"`
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

type deSerializeTest struct {
	name      string
	source    any
	target    any
	size      int
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	serixData, err := testAPI.Encode(context.Background(), test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr)

		return
	}
	require.NoError(t, err)

	require.Equal(t, test.size, len(serixData))

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := testAPI.Decode(context.Background(), serixData, serixTarget)
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr)

		return
	}
	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	sourceJSON, err := testAPI.JSONEncode(context.Background(), test.source)
	require.NoError(t, err)

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, testAPI.JSONDecode(context.Background(), sourceJSON, jsonDest))

	require.EqualValues(t, test.source, jsonDest)
}

func TestSerixMap(t *testing.T) {

	type MyMapType map[string]string

	type MapStruct struct {
		MyMap MyMapType `serix:"0,lengthPrefixType=uint8,mapMinEntries=2,mapMaxEntries=4,mapMaxByteSize=50,mapKeyLengthPrefixType=uint16,mapKeyMinLen=2,mapKeyMaxLen=5,mapValueLengthPrefixType=uint32,mapValueMinLen=1,mapValueMaxLen=6"`
	}

	testAPI.RegisterTypeSettings(MapStruct{}, serix.TypeSettings{}.WithObjectType(uint8(0)))

	tests := []deSerializeTest{
		{
			name: "ok",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			target:    &MapStruct{},
			size:      22,
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "fail - not enough entries",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1",
				},
			},
			target:    &MapStruct{},
			size:      0,
			seriErr:   serix.ErrMapValidationMinElementsNotReached,
			deSeriErr: nil,
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
			target:    &MapStruct{},
			size:      0,
			seriErr:   serix.ErrMapValidationMaxElementsExceeded,
			deSeriErr: nil,
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
			target:    &MapStruct{},
			size:      0,
			seriErr:   serix.ErrMapValidationMaxBytesExceeded,
			deSeriErr: nil,
		},
		{
			name: "fail - key too short",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1",
					"k":  "v2",
				},
			},
			target:    &MapStruct{},
			size:      0,
			seriErr:   serializer.ErrArrayValidationMinElementsNotReached,
			deSeriErr: nil,
		},
		{
			name: "fail - key too long",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1":     "v1",
					"k20000": "v2",
				},
			},
			target:    &MapStruct{},
			size:      0,
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: nil,
		},
		{
			name: "fail - value too short",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1",
					"k2": "",
				},
			},
			target:    &MapStruct{},
			size:      0,
			seriErr:   serializer.ErrArrayValidationMinElementsNotReached,
			deSeriErr: nil,
		},
		{
			name: "fail - value too long",
			source: &MapStruct{
				MyMap: map[string]string{
					"k1": "v1",
					"k2": "v200000",
				},
			},
			target:    &MapStruct{},
			size:      0,
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
