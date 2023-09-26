//nolint:scopelint // we don't care about these linters in test cases
package serix_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Identifier [blake2b.Size256]byte

type serializableStruct struct {
	bytes Identifier `serix:"0"`
	index uint64     `serix:"1"`
}

func (s serializableStruct) EncodeJSON() (any, error) {
	return fmt.Sprintf("%s:%d", base58.Encode(s.bytes[:]), s.index), nil
}

func (s *serializableStruct) DecodeJSON(val any) error {
	serialized, ok := val.(string)
	if !ok {
		return errors.New("incorrect type")
	}

	parts := strings.Split(serialized, ":")
	bytes, err := base58.Decode(parts[0])
	if err != nil {
		return err
	}
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	copy(s.bytes[:], bytes)
	s.index = uint64(idx)
	return nil
}

func TestMapEncodeDecode(t *testing.T) {
	type paras struct {
		api *serix.API
		in  any
	}

	type test struct {
		name     string
		paras    paras
		expected string
	}

	tests := []test{
		{
			name: "basic types",
			paras: func() paras {
				type example struct {
					Uint64    uint64  `serix:"0,mapKey=uint64"`
					Uint32    uint32  `serix:"1,mapKey=uint32"`
					Uint16    uint16  `serix:"2,mapKey=uint16"`
					Uint8     uint8   `serix:"3,mapKey=uint8"`
					Int64     int64   `serix:"4,mapKey=int64"`
					Int32     int32   `serix:"5,mapKey=int32"`
					Int16     int16   `serix:"6,mapKey=int16"`
					Int8      int8    `serix:"7,mapKey=int8"`
					ZeroInt32 int32   `serix:"8,mapKey=zeroInt32,omitempty"`
					Float32   float32 `serix:"9,mapKey=float32"`
					Float64   float64 `serix:"10,mapKey=float64"`
					String    string  `serix:"11,mapKey=string"`
					Bool      bool    `serix:"12,mapKey=bool"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(42))))

				return paras{
					api: api,
					in: &example{
						Uint64:    64,
						Uint32:    32,
						Uint16:    16,
						Uint8:     8,
						Int64:     -64,
						Int32:     -32,
						Int16:     -16,
						Int8:      -8,
						ZeroInt32: 0,
						Float32:   0.33,
						Float64:   0.44,
						String:    "abcd",
						Bool:      true,
					},
				}
			}(),
			expected: `{
				"type": 42,
				"uint64": "64",
				"uint32": 32,
				"uint16": 16,
				"uint8": 8,
				"int64": "-64",
				"int32": -32,
				"int16": -16,
				"int8": -8,
				"float32": "3.3000001311302185E-01",
				"float64": "4.4E-01",
				"string": "abcd",
				"bool": true
			}`,
		},
		{
			name: "big int",
			paras: func() paras {
				type example struct {
					BigInt *big.Int `serix:"0,mapKey=bigInt"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(66))))

				return paras{
					api: api,
					in: &example{
						BigInt: big.NewInt(1337),
					},
				}
			}(),
			expected: `{
				"type": 66,
 				"bigInt": "0x539"
			}`,
		},
		{
			name: "map",
			paras: func() paras {
				type example struct {
					Map map[string]string `serix:"0,mapKey=map"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(99))))

				return paras{
					api: api,
					in: &example{
						Map: map[string]string{
							"alice": "123",
						},
					},
				}
			}(),
			expected: `{
				"type": 99,
 				"map": {
					"alice": "123"
				}
			}`,
		},
		{
			name: "byte slices/arrays",
			paras: func() paras {

				type example struct {
					ByteSlice         []byte    `serix:"0,mapKey=byteSlice"`
					Array             [5]byte   `serix:"1,mapKey=array"`
					SliceOfByteSlices [][]byte  `serix:"3,mapKey=sliceOfByteSlices"`
					SliceOfByteArrays [][3]byte `serix:"4,mapKey=sliceOfByteArrays"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(5))))

				return paras{
					api: api,
					in: &example{
						ByteSlice: []byte{1, 2, 3, 4, 5},
						Array:     [5]byte{5, 4, 3, 2, 1},
						SliceOfByteSlices: [][]byte{
							{1, 2, 3},
							{3, 2, 1},
						},
						SliceOfByteArrays: [][3]byte{
							{5, 6, 7},
							{7, 6, 5},
						},
					},
				}
			}(),
			expected: `{
				"type": 5,
 				"byteSlice": "0x0102030405",
				"array": "0x0504030201",
				"sliceOfByteSlices": [
					"0x010203",
					"0x030201"
				],
				"sliceOfByteArrays": [
					"0x050607",
					"0x070605"
				]
			}`,
		},
		{
			name: "inner struct",
			paras: func() paras {
				type (
					inner struct {
						String string `serix:"0,mapKey=string"`
					}

					example struct {
						inner `serix:"0"`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(22))))

				return paras{
					api: api,
					in: &example{
						inner{String: "abcd"},
					},
				}
			}(),
			expected: `{
				"type": 22,
 				"string": "abcd"
			}`,
		},
		{
			name: "interface & direct pointer",
			paras: func() paras {
				type (
					InterfaceType      interface{}
					InterfaceTypeImpl1 [4]byte
					OtherObj           [2]byte

					example struct {
						Interface InterfaceType `serix:"0,mapKey=interface"`
						Other     *OtherObj     `serix:"1,mapKey=other"`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(33))))
				must(api.RegisterTypeSettings(InterfaceTypeImpl1{},
					serix.TypeSettings{}.WithObjectType(uint8(5)).WithMapKey("customInnerKey")),
				)
				must(api.RegisterInterfaceObjects((*InterfaceType)(nil), (*InterfaceTypeImpl1)(nil)))
				must(api.RegisterTypeSettings(OtherObj{},
					serix.TypeSettings{}.WithObjectType(uint8(2)).WithMapKey("otherObjKey")),
				)

				return paras{
					api: api,
					in: &example{
						Interface: &InterfaceTypeImpl1{1, 2, 3, 4},
						Other:     &OtherObj{1, 2},
					},
				}
			}(),
			expected: `{
				"type": 33,
 				"interface": {
					"type": 5,
					"customInnerKey": "0x01020304"
				},
				"other": {
					"type": 2,
					"otherObjKey": "0x0102"
				}
			}`,
		},
		{
			name: "slice of interface",
			paras: func() paras {
				type (
					Interface interface{}
					Impl1     struct {
						String string `serix:"0,mapKey=string"`
					}
					Impl2 struct {
						Uint16 uint16 `serix:"0,mapKey=uint16"`
					}

					example struct {
						Slice []Interface `serix:"0,mapKey=slice"`
					}
				)

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(11))))
				must(api.RegisterTypeSettings(Impl1{}, serix.TypeSettings{}.WithObjectType(uint8(0))))
				must(api.RegisterTypeSettings(Impl2{}, serix.TypeSettings{}.WithObjectType(uint8(1))))
				must(api.RegisterInterfaceObjects((*Interface)(nil), (*Impl1)(nil), (*Impl2)(nil)))

				return paras{
					api: api,
					in: &example{
						Slice: []Interface{
							&Impl1{String: "impl1"},
							&Impl2{Uint16: 1337},
						},
					},
				}
			}(),
			expected: `{
				"type": 11,
 				"slice": [
					{
						"type": 0,
						"string": "impl1"
					},
					{
						"type": 1,
						"uint16": 1337
					}
				]
			}`,
		},
		{
			name: "no map key",
			paras: func() paras {
				type example struct {
					CaptainHook string `serix:"0"`
					LiquidSoul  int64  `serix:"1"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				return paras{
					api: api,
					in: &example{
						CaptainHook: "jump",
						LiquidSoul:  30,
					},
				}
			}(),
			expected: `{
				"type": 23,
 				"captainHook": "jump",
				"liquidSoul": "30"
			}`,
		},
		{
			name: "time",
			paras: func() paras {
				type example struct {
					CreationDate time.Time `serix:"0"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				exampleTime, err := time.Parse(time.RFC3339Nano, "2022-08-12T12:51:18.120072+02:00")
				require.NoError(t, err)

				return paras{
					api: api,
					in: &example{
						CreationDate: exampleTime,
					},
				}
			}(),
			expected: `{
				"type": 23,
 				"creationDate": "2022-08-12T12:51:18.120072+02:00"
			}`,
		},

		{
			name: "serializable",
			paras: func() paras {
				type example struct {
					Entries map[serializableStruct]struct{} `serix:"0"`
				}

				api := serix.NewAPI()
				must(api.RegisterTypeSettings(example{}, serix.TypeSettings{}.WithObjectType(uint8(23))))

				return paras{
					api: api,
					in: &example{
						Entries: map[serializableStruct]struct{}{
							serializableStruct{
								bytes: blake2b.Sum256([]byte("test")),
								index: 1,
							}: struct{}{},
						},
					},
				}
			}(),
			expected: `{
				"type": 23,
				"entries": {
					"As3ZuwnL9LpoW3wz8HoDpHtZqJ4dhPFFnv87GYrnCYKj:1": {}
				}
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// encode input to a map
			out, err := test.paras.api.MapEncode(context.Background(), test.paras.in, serix.WithValidation())
			require.NoError(t, err)
			jsonOut, err := json.MarshalIndent(out, "", "\t")
			require.NoError(t, err)

			// re-arrange expected json output to conform to same indentation
			aux := orderedmap.New()
			require.NoError(t, json.Unmarshal([]byte(test.expected), aux))
			expectedJSON, err := json.MarshalIndent(aux, "", "\t")
			require.NoError(t, err)
			require.EqualValues(t, string(expectedJSON), string(jsonOut))

			mapTarget := map[string]any{}
			require.NoError(t, json.Unmarshal(expectedJSON, &mapTarget))

			dest := reflect.New(reflect.TypeOf(test.paras.in).Elem()).Interface()
			require.NoError(t, test.paras.api.MapDecode(context.Background(), mapTarget, dest))
			require.EqualValues(t, test.paras.in, dest)
		})
	}
}
