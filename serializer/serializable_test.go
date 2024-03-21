//nolint:gosec // we don't care about these linters in test cases
package serializer_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
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
	ErrUnknownDummyType = ierrors.New("unknown example type")

	dummyTypeArrayRules = &serializer.ArrayRules{
		Guards: serializer.SerializableGuard{
			ReadGuard: DummyTypeSelector,
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *A:
				case *B:
					return ErrUnknownDummyType
				}

				return nil
			},
		},
	}
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

func (a *A) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(a.Key[:], data[:aKeyLength])

	return typeALength, nil
}

func (a *A) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
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

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}

	return b
}

func randSerializedA() []byte {
	var b [typeALength]byte
	b[0] = TypeA
	keyData := RandBytes(aKeyLength)
	copy(b[serializer.SmallTypeDenotationByteSize:], keyData)

	return b[:]
}

func randA() *A {
	var k [aKeyLength]byte
	copy(k[:], RandBytes(aKeyLength))

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

func (b *B) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(b.Name[:], data[:bNameLength])

	return typeBLength, nil
}

func (b *B) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var bf [typeBLength]byte
	bf[0] = TypeB
	copy(bf[serializer.SmallTypeDenotationByteSize:], b.Name[:])

	return bf[:], nil
}

func randB() *B {
	var n [bNameLength]byte
	copy(n[:], RandBytes(bNameLength))

	return &B{Name: n}
}

func TestDeserializeA(t *testing.T) {
	seriA := randSerializedA()
	objA := &A{}
	bytesRead, err := objA.Deserialize(seriA, serializer.DeSeriModePerformValidation, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[serializer.SmallTypeDenotationByteSize:], objA.Key[:])
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
		ar    *serializer.ArrayRules
	}

	tests := []test{
		{
			name: "ok - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			ar:    &serializer.ArrayRules{},
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
			ar:    &serializer.ArrayRules{},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := tt.ar.ElementUniqueValidator()

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
		ar    *serializer.ArrayRules
	}

	tests := []test{
		{
			name: "ok - order by first ele",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			ar:    &serializer.ArrayRules{},
			valid: true,
		},
		{
			name: "ok - order by last ele",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			ar:    &serializer.ArrayRules{},
			valid: true,
		},
		{
			name: "not ok",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			ar:    &serializer.ArrayRules{},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := tt.ar.LexicalOrderValidator()

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
		ar    *serializer.ArrayRules
	}

	tests := []test{
		{
			name: "ok - order by first ele - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			ar:    &serializer.ArrayRules{},
			valid: true,
		},
		{
			name: "ok - order by last ele - no dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			ar:    &serializer.ArrayRules{},
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
			ar:    &serializer.ArrayRules{},
			valid: false,
		},
		{
			name: "not ok - order",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			ar:    &serializer.ArrayRules{},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := tt.ar.LexicalOrderWithoutDupsValidator()

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
		ar    *serializer.ArrayRules
		ty    serializer.TypeDenotationType
	}

	tests := []test{
		{
			name: "ok - types unique - byte",
			args: [][]byte{
				{1, 1, 1},
				{2, 2, 2},
				{3, 3, 3},
			},
			valid: true,
			ar:    &serializer.ArrayRules{},
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
			ar:    &serializer.ArrayRules{},
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
			ar:    &serializer.ArrayRules{},
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
			ar:    &serializer.ArrayRules{},
			ty:    serializer.TypeDenotationUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := tt.ar.AtMostOneOfEachTypeValidator(tt.ty)

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
