package serix_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/hive.go/serix"
)

const defaultSeriMode = serializer.DeSeriModePerformValidation

var (
	testAPI            = serix.NewAPI()
	ctx                = context.Background()
	defaultArrayRules  = &serializer.ArrayRules{}
	defaultErrProducer = func(err error) error { return err }
	defaultWriteGuard  = func(seri serializer.Serializable) error { return nil }
)

type Bool bool

func (b Bool) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b Bool) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (b Bool) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b Bool) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	ser := serializer.NewSerializer()
	ser.WriteBool(bool(b), defaultErrProducer)
	return ser.Serialize()
}

type Bools []Bool

var boolsLenType = serix.LengthPrefixTypeAsUint16

func (bs Bools) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (bs Bools) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (bs Bools) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (bs Bools) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WriteSliceOfObjects(bs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixType(boolsLenType), defaultArrayRules, defaultErrProducer)
	return s.Serialize()
}

func (bs Bools) ToSerializables() serializer.Serializables {
	serializables := make(serializer.Serializables, len(bs))
	for i, b := range bs {
		serializables[i] = b
	}
	return serializables
}

func (bs Bools) FromSerializables(seris serializer.Serializables) {
	//TODO implement me
	panic("implement me")
}

type SimpleStruct struct {
	Bool       bool      `serix:"0"`
	Uint       uint64    `serix:"1"`
	String     string    `serix:"2,lengthPrefixType=uint16"`
	Bytes      []byte    `serix:"3,lengthPrefixType=uint32"`
	BytesArray [16]byte  `serix:"4"`
	BigInt     *big.Int  `serix:"5"`
	Time       time.Time `serix:"6"`
	Int        uint64    `serix:"7"`
	Float      float64   `serix:"8"`
}

func NewSimpleStruct() SimpleStruct {
	return SimpleStruct{
		Bool:       true,
		Uint:       10,
		String:     "foo",
		Bytes:      []byte{1, 2, 3},
		BytesArray: [16]byte{3, 2, 1},
		BigInt:     big.NewInt(8),
		Time:       time.Unix(1000, 1000),
		Int:        23,
		Float:      4.44,
	}
}

var simpleStructObjectCode = uint32(0)

func (ss SimpleStruct) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (ss SimpleStruct) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (ss SimpleStruct) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (ss SimpleStruct) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WriteNum(simpleStructObjectCode, defaultErrProducer)
	s.WriteBool(ss.Bool, defaultErrProducer)
	s.WriteNum(ss.Uint, defaultErrProducer)
	s.WriteString(ss.String, serializer.SeriLengthPrefixTypeAsUint16, defaultErrProducer)
	s.WriteVariableByteSlice(ss.Bytes, serializer.SeriLengthPrefixTypeAsUint32, defaultErrProducer)
	s.WriteBytes(ss.BytesArray[:], defaultErrProducer)
	s.WriteUint256(ss.BigInt, defaultErrProducer)
	s.WriteTime(ss.Time, defaultErrProducer)
	s.WriteNum(ss.Int, defaultErrProducer)
	s.WriteNum(ss.Float, defaultErrProducer)
	return s.Serialize()
}

type Interface interface {
	Method()

	serix.Serializable
	serix.Deserializable
}

type InterfaceImpl struct {
	interfaceImpl `serix:"0"`
}

type interfaceImpl struct {
	A uint8 `serix:"0"`
	B uint8 `serix:"1"`
}

func (ii *InterfaceImpl) Encode() ([]byte, error) {
	return testAPI.Encode(context.Background(), ii.interfaceImpl, serix.WithValidation())
}

func (ii *InterfaceImpl) Decode(b []byte) (consumedBytes int, err error) {
	return testAPI.Decode(context.Background(), b, &ii.interfaceImpl, serix.WithValidation())
}

var interfaceImplObjectCode = uint32(1)

func (ii *InterfaceImpl) Method() {}

func (ii *InterfaceImpl) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (ii *InterfaceImpl) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (ii *InterfaceImpl) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (ii *InterfaceImpl) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	ser := serializer.NewSerializer()
	ser.WriteNum(interfaceImplObjectCode, defaultErrProducer)
	ser.WriteNum(ii.A, defaultErrProducer)
	ser.WriteNum(ii.B, defaultErrProducer)
	return ser.Serialize()
}

type StructWithInterface struct {
	Interface Interface `serix:"0"`
}

func (si StructWithInterface) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (si StructWithInterface) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (si StructWithInterface) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (si StructWithInterface) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WriteObject(si.Interface.(serializer.Serializable), defaultSeriMode, deSeriCtx, defaultWriteGuard, defaultErrProducer)
	return s.Serialize()
}

type StructWithOptionalField struct {
	Optional *ExportedStruct `serix:"0,optional"`
}

func (so StructWithOptionalField) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (so StructWithOptionalField) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (so StructWithOptionalField) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (so StructWithOptionalField) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WritePayloadLength(0, defaultErrProducer)
	return s.Serialize()
}

type StructWithEmbeddedStructs struct {
	unexportedStruct `serix:"0"`
	ExportedStruct   `serix:"1,nest"`
}

func (se StructWithEmbeddedStructs) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WriteNum(se.unexportedStruct.Foo, defaultErrProducer)
	s.WriteNum(exportedStructObjectCode, defaultErrProducer)
	s.WriteNum(se.ExportedStruct.Bar, defaultErrProducer)
	return s.Serialize()
}

type unexportedStruct struct {
	Foo uint64 `serix:"0"`
}

type ExportedStruct struct {
	Bar uint64 `serix:"0"`
}

var exportedStructObjectCode = uint32(3)

type Map map[uint64]uint64

var mapLenType = serix.LengthPrefixTypeAsUint32

func (m Map) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m Map) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (m Map) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m Map) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	bytes := make([][]byte, len(m))
	var i int
	for k, v := range m {
		s := serializer.NewSerializer()
		s.WriteNum(k, defaultErrProducer)
		s.WriteNum(v, defaultErrProducer)
		b, err := s.Serialize()
		if err != nil {
			return nil, err
		}
		bytes[i] = b
		i++
	}
	s := serializer.NewSerializer()
	mode := defaultSeriMode | serializer.DeSeriModePerformLexicalOrdering
	arrayRules := &serializer.ArrayRules{ValidationMode: serializer.ArrayValidationModeLexicalOrdering}
	s.WriteSliceOfByteSlices(bytes, mode, serializer.SeriLengthPrefixType(mapLenType), arrayRules, defaultErrProducer)
	return s.Serialize()
}

type CustomSerializable int

func (cs CustomSerializable) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (cs CustomSerializable) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (cs CustomSerializable) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (cs CustomSerializable) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return cs.Encode()
}

func (cs CustomSerializable) Encode() ([]byte, error) {
	b := []byte(fmt.Sprintf("int: %d", cs))
	return b, nil
}

func (cs *CustomSerializable) Decode(b []byte) (int, error) {
	_, err := fmt.Sscanf(string(b), "int: %d", cs)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

type ObjectForSyntacticValidation struct{}

var errSyntacticValidation = errors.New("syntactic validation failed")

func SyntacticValidation(ctx context.Context, obj ObjectForSyntacticValidation) error {
	return errSyntacticValidation
}

type ObjectForBytesValidation struct{}

var errBytesValidation = errors.New("bytes validation failed")

func BytesValidation(ctx context.Context, b []byte) error {
	return errBytesValidation
}

func TestMain(m *testing.M) {
	exitCode := func() int {
		if err := testAPI.RegisterTypeSettings(
			SimpleStruct{},
			serix.TypeSettings{}.WithObjectType(simpleStructObjectCode),
		); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterTypeSettings(
			InterfaceImpl{},
			serix.TypeSettings{}.WithObjectType(interfaceImplObjectCode),
		); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterTypeSettings(
			ExportedStruct{},
			serix.TypeSettings{}.WithObjectType(exportedStructObjectCode),
		); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterInterfaceObjects((*Interface)(nil), (*InterfaceImpl)(nil)); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterValidators(ObjectForSyntacticValidation{}, nil, SyntacticValidation); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterValidators(ObjectForBytesValidation{}, BytesValidation, nil); err != nil {
			log.Panic(err)
		}
		return m.Run()
	}()
	os.Exit(exitCode)
}
