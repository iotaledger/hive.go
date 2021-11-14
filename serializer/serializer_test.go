package serializer_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestDeserializer_ReadObject(t *testing.T) {
	seriA := randSerializedA()

	var objA serializer.Serializable
	bytesRead, err := serializer.NewDeserializer(seriA).
		ReadObject(func(seri serializer.Serializable) { objA = seri }, serializer.DeSeriModePerformValidation, serializer.TypeDenotationByte, DummyTypeSelector, func(err error) error { return err }).
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
		seriBytes, err := seri.Serialize(serializer.DeSeriModePerformValidation)
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
		}, serializer.DeSeriModePerformValidation, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, DummyTypeSelector, nil, func(err error) error { return err }).
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
		seriBytes, err := seri.Serialize(serializer.DeSeriModePerformValidation)
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()

	withStringers := StructWithStringers{}

	bytesRead, err := serializer.NewDeserializer(data).
		ReadSliceOfObjects(&withStringers.Objects, serializer.DeSeriModePerformValidation, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, DummyTypeSelector, nil, func(err error) error { return err }).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, originObjs, withStringers.Objects.ToSerializables())
}

func TestDeserializer_ReadPayload(t *testing.T) {
	source := randA()

	data, _ := serializer.NewSerializer().WritePayload(source, serializer.DeSeriModePerformValidation, func(err error) error {
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
		ReadPayload(&withA.A, serializer.DeSeriModePerformValidation, DummyTypeSelector, func(err error) error {
			return err
		}).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, source, withA.A)

	bytesRead, err = serializer.NewDeserializer(data).
		ReadPayload(&withAAsInterface.A, serializer.DeSeriModePerformValidation, DummyTypeSelector, func(err error) error {
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
				}).
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
			_, err = serializer.NewDeserializer(serializedBytes).ReadUint256(y, returnErr).Done()
			if err != nil && tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, 0, tt.x.Cmp(y))
		})
	}
}
