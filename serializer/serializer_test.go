package serializer_test

import (
	"bytes"
	"encoding/binary"
	"github.com/iotaledger/hive.go/serializer"
	"testing"

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
		}, serializer.DeSeriModePerformValidation, serializer.TypeDenotationByte, DummyTypeSelector, nil, func(err error) error { return err }).
		ConsumedAll(func(left int, err error) error { return err }).
		Done()

	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesRead)
	assert.EqualValues(t, originObjs, readObjects)
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
				ReadString(&s, func(err error) error {
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
