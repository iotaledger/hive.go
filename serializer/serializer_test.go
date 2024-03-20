package serializer_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
)

func TestDeserializer_ReadObject(t *testing.T) {
	seriA := randSerializedA()

	var objA serializer.Serializable
	bytesRead, err := serializer.NewDeserializer(seriA).
		ReadObject(func(seri serializer.Serializable) { objA = seri }, serializer.DeSeriModePerformValidation, nil, serializer.TypeDenotationByte, DummyTypeSelector, func(err error) error { return err }).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.IsType(t, &A{}, objA)
	assert.Equal(t, seriA[serializer.SmallTypeDenotationByteSize:], objA.(*A).Key[:])
}

func TestDeserializer_ReadBytesInPlace(t *testing.T) {
	type Example struct {
		Out [10]byte
	}
	in := [10]byte{1}
	example := &Example{}
	_, err := serializer.NewDeserializer(in[:]).ReadBytesInPlace(example.Out[:], func(err error) error {
		return err
	}).Done()
	require.NoError(t, err)
	require.EqualValues(t, in, example.Out)
}

func TestDeserializer_ReadSliceOfObjects(t *testing.T) {
	var buf bytes.Buffer
	originObjs := serializer.Serializables{
		randA(), randA(), randB(), randA(), randB(), randB(),
	}
	assert.NoError(t, binary.Write(&buf, binary.LittleEndian, uint16(len(originObjs))))

	for _, seri := range originObjs {
		seriBytes, err := seri.Serialize(serializer.DeSeriModePerformValidation, nil)
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()

	var readObjects serializer.Serializables
	bytesRead, err := serializer.NewDeserializer(data).
		ReadSliceOfObjects(func(seri serializer.Serializables) {
			readObjects = seri
		}, serializer.DeSeriModePerformValidation, nil, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, dummyTypeArrayRules, func(err error) error { return err }).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, originObjs, readObjects)
}

type Stringers []fmt.Stringer

func (s Stringers) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(s))
	for i, x := range s {
		seris[i] = x.(serializer.Serializable)
	}

	return seris
}

func (s *Stringers) FromSerializables(seris serializer.Serializables) {
	*s = make(Stringers, len(seris))
	for i, seri := range seris {
		(*s)[i] = seri.(fmt.Stringer)
	}
}

type StructWithStringers struct {
	Objects Stringers
}

func TestDeserializer_ReadSliceOfObjectsWithSerializableSlice(t *testing.T) {
	var buf bytes.Buffer
	originObjs := serializer.Serializables{
		randA(), randA(), randB(), randA(), randB(), randB(),
	}
	assert.NoError(t, binary.Write(&buf, binary.LittleEndian, uint16(len(originObjs))))

	for _, seri := range originObjs {
		seriBytes, err := seri.Serialize(serializer.DeSeriModePerformValidation, nil)
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()

	withStringers := StructWithStringers{}

	bytesRead, err := serializer.NewDeserializer(data).
		ReadSliceOfObjects(&withStringers.Objects, serializer.DeSeriModePerformValidation, nil, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, dummyTypeArrayRules, func(err error) error { return err }).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, originObjs, withStringers.Objects.ToSerializables())
}

func TestDeserializer_ReadPayload(t *testing.T) {
	source := randA()

	data, _ := serializer.NewSerializer().WritePayload(source, serializer.DeSeriModePerformValidation, nil, dummyTypeArrayRules.Guards.WriteGuard, func(err error) error {
		require.NoError(t, err)

		return err
	}).Serialize()

	type StructWithA struct {
		A *A
	}

	type StructWithAAsInterface struct {
		A fmt.Stringer
	}

	withA := StructWithA{}
	withAAsInterface := StructWithAAsInterface{}

	bytesRead, err := serializer.NewDeserializer(data).
		ReadPayload(&withA.A, serializer.DeSeriModePerformValidation, nil, DummyTypeSelector, func(err error) error {
			return err
		}).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, source, withA.A)

	bytesRead, err = serializer.NewDeserializer(data).
		ReadPayload(&withAAsInterface.A, serializer.DeSeriModePerformValidation, nil, DummyTypeSelector, func(err error) error {
			return err
		}).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, source, withAAsInterface.A)
}

func TestDeserializer_ReadString(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				data: []byte{17, 0, 72, 101, 108, 108, 111, 44, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100},
			},
			want:    "Hello, playground",
			wantErr: false,
		},
		{
			name: "not enough (length denotation)",
			args: args{
				data: []byte{0, 1},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "not enough (actual length)",
			args: args{
				data: []byte{17, 0, 72, 101, 108, 108, 111, 44, 32, 112},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s string
			_, err := serializer.NewDeserializer(tt.args.data).
				ReadString(&s, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
					return err
				}, 0, 0).
				ConsumedAll(func(left int, err error) error { return err }).
				Done()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadStringFromBytes() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if s != tt.want {
				t.Errorf("ReadStringFromBytes() got = %v, want %v", s, tt.want)
			}
		})
	}
}

func TestReadWriteNum(t *testing.T) {
	tests := []struct {
		name        string
		x           any
		expectedErr error
	}{
		{
			name:        "ok - uint8",
			x:           uint8(1),
			expectedErr: nil,
		},
		{
			name:        "ok - uint16",
			x:           uint16(1),
			expectedErr: nil,
		},
		{
			name:        "ok - uint32",
			x:           uint32(1),
			expectedErr: nil,
		},
		{
			name:        "ok - uint64",
			x:           uint64(1),
			expectedErr: nil,
		},
		{
			name:        "ok - int8",
			x:           int8(1),
			expectedErr: nil,
		},
		{
			name:        "ok - int16",
			x:           int16(1),
			expectedErr: nil,
		},
		{
			name:        "ok - int32",
			x:           int32(1),
			expectedErr: nil,
		},
		{
			name:        "ok - int64",
			x:           int64(1),
			expectedErr: nil,
		},
		{
			name:        "ok - float32",
			x:           float32(1),
			expectedErr: nil,
		},
		{
			name:        "ok - float64",
			x:           float64(1),
			expectedErr: nil,
		},
	}

	returnErr := func(err error) error { return err }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializedBytes, err := serializer.NewSerializer().WriteNum(tt.x, returnErr).Serialize()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}
			require.NoError(t, err)

			y := reflect.New(reflect.TypeOf(tt.x)).Interface()
			_, err = serializer.NewDeserializer(serializedBytes).ReadNum(y, returnErr).Done()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.x, reflect.ValueOf(y).Elem().Interface())
		})
	}
}

func TestReadWriteUint256(t *testing.T) {
	tests := []struct {
		name        string
		x           *big.Int
		expectedErr error
	}{
		{
			name:        "ok - max uint256",
			x:           abi.MaxUint256,
			expectedErr: nil,
		},
		{
			name:        "ok - 0",
			x:           new(big.Int).SetInt64(0),
			expectedErr: nil,
		},
		{
			name:        "ok - 1337",
			x:           new(big.Int).SetInt64(1337),
			expectedErr: nil,
		},
		{
			name:        "err - negative 10",
			x:           new(big.Int).SetInt64(-10),
			expectedErr: serializer.ErrUint256NumNegative,
		},
		{
			name:        "err - too big",
			x:           new(big.Int).Add(abi.MaxUint256, abi.MaxUint256),
			expectedErr: serializer.ErrUint256TooBig,
		},
		{
			name:        "err - nil big.Int",
			x:           nil,
			expectedErr: serializer.ErrUint256Nil,
		},
	}

	returnErr := func(err error) error { return err }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializedBytes, err := serializer.NewSerializer().WriteUint256(tt.x, returnErr).Serialize()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}

			require.NoError(t, err)

			y := new(big.Int)
			_, err = serializer.NewDeserializer(serializedBytes).ReadUint256(&y, returnErr).Done()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}
			require.NoError(t, err)

			require.Equal(t, 0, tt.x.Cmp(y))
		})
	}
}

func TestReadWriteTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name              string
		timeToWrite       time.Time
		expectedTimestamp time.Time
		expectedErr       error
	}{
		{
			name:              "ok - unix epoch",
			timeToWrite:       time.Unix(0, 0),
			expectedTimestamp: time.Unix(0, 0),
			expectedErr:       nil,
		},
		{
			name:              "ok - now",
			timeToWrite:       now,
			expectedTimestamp: now,
			expectedErr:       nil,
		},
		{
			name:              "ok - maximum representable timestamp",
			timeToWrite:       time.Unix(0, math.MaxInt64),
			expectedTimestamp: time.Unix(0, math.MaxInt64),
			expectedErr:       nil,
		},
		{
			name:              "ok - time before unix epoch is truncated",
			timeToWrite:       time.Unix(-1_000_000, 0),
			expectedTimestamp: time.Unix(0, 0),
			expectedErr:       nil,
		},
		{
			name:              "ok - time nano before unix epoch is truncated",
			timeToWrite:       time.Unix(0, -1_000_000),
			expectedTimestamp: time.Unix(0, 0),
			expectedErr:       nil,
		},
		{
			name:              "ok - time before min representable is truncated",
			timeToWrite:       time.Unix(-(serializer.MaxNanoTimestampInt64Seconds + 1), 0),
			expectedTimestamp: time.Unix(0, 0),
			expectedErr:       nil,
		},
		{
			name:              "ok - time after max representable is truncated to max",
			timeToWrite:       time.Unix(serializer.MaxNanoTimestampInt64Seconds+1, 0),
			expectedTimestamp: time.Unix(0, math.MaxInt64),
			expectedErr:       nil,
		},
	}

	returnErr := func(err error) error { return err }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializedBytes, err := serializer.NewSerializer().WriteTime(tt.timeToWrite, returnErr).Serialize()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}

			require.NoError(t, err)

			timestamp := time.Time{}
			_, err = serializer.NewDeserializer(serializedBytes).ReadTime(&timestamp, returnErr).Done()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)

				return
			}
			require.NoError(t, err)

			require.True(t, tt.expectedTimestamp.Equal(timestamp))
		})
	}
}
