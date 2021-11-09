package serializer_test

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/hive.go/testutil"

	"github.com/stretchr/testify/assert"
)

const (
	TypeA       byte = 0
	TypeB       byte = 1
	aKeyLength       = 16
	bNameLength      = 32
	typeALength      = serializer.SmallTypeDenotationByteSize + aKeyLength
	typeBLength      = serializer.SmallTypeDenotationByteSize + bNameLength
)

var (
	ErrUnknownDummyType = errors.New("unknown example type")
)

func DummyTypeSelector(dummyType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(dummyType) {
	case TypeA:
		seri = &A{}
	case TypeB:
		seri = &B{}
	default:
		return nil, ErrUnknownDummyType
	}
	return seri, nil
}

type Keyer interface {
	GetKey() [aKeyLength]byte
}

type A struct {
	Key [aKeyLength]byte
}

func (a *A) String() string {
	return "A"
}

func (a *A) GetKey() [16]byte {
	return a.Key
}

func (a *A) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (a *A) UnmarshalJSON(i []byte) error {
	panic("implement me")
}

func (a *A) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(a.Key[:], data[:aKeyLength])
	return typeALength, nil
}

func (a *A) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var b [typeALength]byte
	b[0] = TypeA
	copy(b[serializer.SmallTypeDenotationByteSize:], a.Key[:])
	return b[:], nil
}

type As []*A

func (a As) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(a))
	for i, x := range a {
		seris[i] = x
	}
	return seris
}

func (a *As) FromSerializables(seris serializer.Serializables) {
	*a = make(As, len(seris))
	for i, seri := range seris {
		(*a)[i] = seri.(*A)
	}
}

type Keyers []Keyer

func (k Keyers) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(k))
	for i, x := range k {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (k *Keyers) FromSerializables(seris serializer.Serializables) {
	*k = make(Keyers, len(seris))
	for i, seri := range seris {
		(*k)[i] = seri.(Keyer)
	}
}

func randSerializedA() []byte {
	var b [typeALength]byte
	b[0] = TypeA
	keyData := testutil.RandBytes(aKeyLength)
	copy(b[serializer.SmallTypeDenotationByteSize:], keyData)
	return b[:]
}

func randA() *A {
	var k [aKeyLength]byte
	copy(k[:], testutil.RandBytes(aKeyLength))
	return &A{Key: k}
}

type B struct {
	Name [bNameLength]byte
}

func (b *B) String() string {
	return "B"
}

func (b *B) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (b *B) UnmarshalJSON(i []byte) error {
	panic("implement me")
}

func (b *B) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(b.Name[:], data[:bNameLength])
	return typeBLength, nil
}

func (b *B) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var bf [typeBLength]byte
	bf[0] = TypeB
	copy(bf[serializer.SmallTypeDenotationByteSize:], b.Name[:])
	return bf[:], nil
}

func randB() *B {
	var n [bNameLength]byte
	copy(n[:], testutil.RandBytes(bNameLength))
	return &B{Name: n}
}

func TestDeserializeA(t *testing.T) {
	seriA := randSerializedA()
	objA := &A{}
	bytesRead, err := objA.Deserialize(seriA, serializer.DeSeriModePerformValidation)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[serializer.SmallTypeDenotationByteSize:], objA.Key[:])
}

func TestLexicalOrderedByteSlices(t *testing.T) {
	type test struct {
		name   string
		source serializer.LexicalOrderedByteSlices
		target serializer.LexicalOrderedByteSlices
	}
	tests := []test{
		{
			name: "ok - order by first ele",
			source: serializer.LexicalOrderedByteSlices{
				{3, 2, 1},
				{2, 3, 1},
				{1, 2, 3},
			},
			target: serializer.LexicalOrderedByteSlices{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
		},
		{
			name: "ok - order by last ele",
			source: serializer.LexicalOrderedByteSlices{
				{1, 1, 3},
				{1, 1, 2},
				{1, 1, 1},
			},
			target: serializer.LexicalOrderedByteSlices{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.source)
			assert.Equal(t, tt.target, tt.source)
		})
	}
}

func TestRemoveDupsAndSortByLexicalOrderArrayOf32Bytes(t *testing.T) {
	type test struct {
		name   string
		source serializer.LexicalOrdered32ByteArrays
		target serializer.LexicalOrdered32ByteArrays
	}
	tests := []test{
		{
			name: "ok - dups removed and order by first ele",
			source: serializer.LexicalOrdered32ByteArrays{
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			target: serializer.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
		},
		{
			name: "ok - dups removed and order by last ele",
			source: serializer.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			target: serializer.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.source = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(tt.source)
			assert.Equal(t, tt.target, tt.source)
		})
	}
}

func TestSerializationMode_HasMode(t *testing.T) {
	type args struct {
		mode serializer.DeSerializationMode
	}
	tests := []struct {
		name string
		sm   serializer.DeSerializationMode
		args args
		want bool
	}{
		{
			"has no validation",
			serializer.DeSeriModeNoValidation,
			args{mode: serializer.DeSeriModePerformValidation},
			false,
		},
		{
			"has validation",
			serializer.DeSeriModePerformValidation,
			args{mode: serializer.DeSeriModePerformValidation},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sm.HasMode(tt.args.mode); got != tt.want {
				t.Errorf("HasMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayValidationMode_HasMode(t *testing.T) {
	type args struct {
		mode serializer.ArrayValidationMode
	}
	tests := []struct {
		name string
		sm   serializer.ArrayValidationMode
		args args
		want bool
	}{
		{
			"has no validation",
			serializer.ArrayValidationModeNone,
			args{mode: serializer.ArrayValidationModeNoDuplicates},
			false,
		},
		{
			"has mode duplicates",
			serializer.ArrayValidationModeNoDuplicates,
			args{mode: serializer.ArrayValidationModeNoDuplicates},
			true,
		},
		{
			"has mode lexical order",
			serializer.ArrayValidationModeLexicalOrdering,
			args{mode: serializer.ArrayValidationModeLexicalOrdering},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sm.HasMode(tt.args.mode); got != tt.want {
				t.Errorf("HasMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayRules_ElementUniqueValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := serializer.ArrayRules{}

	tests := []test{
		{
			name: "ok - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "not ok - dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
				{1, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.ElementUniqueValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestArrayRules_Bounds(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		min   int
		max   int
		valid bool
	}

	arrayRules := serializer.ArrayRules{}

	tests := []test{
		{
			name: "ok - min",
			args: [][]byte{
				{1},
			},
			min:   1,
			max:   3,
			valid: true,
		},
		{
			name: "ok - max",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   1,
			max:   3,
			valid: true,
		},
		{
			name: "not ok - min",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   4,
			max:   5,
			valid: false,
		},
		{
			name: "not ok - max",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   1,
			max:   2,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayRules.Min = uint(tt.min)
			arrayRules.Max = uint(tt.max)
			err := arrayRules.CheckBounds(uint(len(tt.args)))
			assert.Equal(t, tt.valid, err == nil)
		})
	}
}

func TestArrayRules_LexicalOrderValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := serializer.ArrayRules{}

	tests := []test{
		{
			name: "ok - order by first ele",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "ok - order by last ele",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			valid: true,
		},
		{
			name: "not ok",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.LexicalOrderValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestArrayRules_LexicalOrderWithoutDupsValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := serializer.ArrayRules{}

	tests := []test{
		{
			name: "ok - order by first ele - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "ok - order by last ele - no dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			valid: true,
		},
		{
			name: "not ok - dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
				{1, 1, 3},
			},
			valid: false,
		},
		{
			name: "not ok - order",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.LexicalOrderWithoutDupsValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestArrayRules_AtMostOneOfEachTypeValidatorValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
		ty    serializer.TypeDenotationType
	}

	arrayRules := serializer.ArrayRules{}

	tests := []test{
		{
			name: "ok - types unique - byte",
			args: [][]byte{
				{1, 1, 1},
				{2, 2, 2},
				{3, 3, 3},
			},
			valid: true,
			ty:    serializer.TypeDenotationByte,
		},
		{
			name: "ok - types unique - uint32",
			args: [][]byte{
				{1, 1, 1, 1},
				{2, 2, 2, 2},
				{3, 3, 3, 3},
			},
			valid: true,
			ty:    serializer.TypeDenotationUint32,
		},
		{
			name: "not ok - types not unique - byte",
			args: [][]byte{
				{1, 1, 1},
				{1, 2, 2},
				{3, 3, 3},
			},
			valid: false,
			ty:    serializer.TypeDenotationByte,
		},
		{
			name: "not ok - types not unique - uint32",
			args: [][]byte{
				{1, 1, 1, 1},
				{2, 2, 2, 2},
				{1, 1, 1, 1},
			},
			valid: false,
			ty:    serializer.TypeDenotationUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.AtMostOneOfEachTypeValidator(tt.ty)

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestSerializableSlice(t *testing.T) {
	keyers := make(Keyers, 0)

	seris := make(serializer.Serializables, 5)
	for i := range seris {
		seris[i] = randA()
	}

	keyers.FromSerializables(seris)

	for _, a := range keyers {
		fmt.Println(a.GetKey())
	}
}
