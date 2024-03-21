package serix_test

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

type Bool bool

func (b Bool) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (b Bool) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (b Bool) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
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
	// ToDo: implement me
	panic("implement me")
}

func (bs Bools) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (bs Bools) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (bs Bools) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WriteSliceOfObjects(bs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixType(boolsLenType), defaultArrayRules, defaultErrProducer)

	return seri.Serialize()
}

func (bs Bools) ToSerializables() serializer.Serializables {
	serializables := make(serializer.Serializables, len(bs))
	for i, b := range bs {
		serializables[i] = b
	}

	return serializables
}

func (bs Bools) FromSerializables(seris serializer.Serializables) {
	// ToDo: implement me
	panic("implement me")
}

type SimpleStruct struct {
	Bool       bool      `serix:""`
	Uint       uint64    `serix:""`
	String     string    `serix:",lenPrefix=uint16"`
	Bytes      []byte    `serix:",lenPrefix=uint32"`
	BytesArray [16]byte  `serix:""`
	BigInt     *big.Int  `serix:""`
	Time       time.Time `serix:""`
	Int        uint64    `serix:""`
	Float      float64   `serix:""`
}

func NewSimpleStruct() SimpleStruct {
	return SimpleStruct{
		Bool:       true,
		Uint:       10,
		String:     "foo",
		Bytes:      []byte{1, 2, 3},
		BytesArray: [16]byte{3, 2, 1},
		BigInt:     big.NewInt(8),
		Time:       time.Unix(1000, 1000).UTC(),
		Int:        23,
		Float:      4.44,
	}
}

var simpleStructObjectCode = uint32(0)

func (ss SimpleStruct) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (ss SimpleStruct) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (ss SimpleStruct) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (ss SimpleStruct) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WriteNum(simpleStructObjectCode, defaultErrProducer)
	seri.WriteBool(ss.Bool, defaultErrProducer)
	seri.WriteNum(ss.Uint, defaultErrProducer)
	seri.WriteString(ss.String, serializer.SeriLengthPrefixTypeAsUint16, defaultErrProducer, 0, 0)
	seri.WriteVariableByteSlice(ss.Bytes, serializer.SeriLengthPrefixTypeAsUint32, defaultErrProducer, 0, 0)
	seri.WriteBytes(ss.BytesArray[:], defaultErrProducer)
	seri.WriteUint256(ss.BigInt, defaultErrProducer)
	seri.WriteTime(ss.Time, defaultErrProducer)
	seri.WriteNum(ss.Int, defaultErrProducer)
	seri.WriteNum(ss.Float, defaultErrProducer)

	return seri.Serialize()
}

type Interface interface {
	Method()

	serix.Serializable
	serix.Deserializable
}

type InterfaceImpl struct {
	interfaceImpl `serix:""`
}

func (ii *InterfaceImpl) SetDeserializationContext(ctx context.Context) {
	ctx.Value("contextValue").(func())()
}

type interfaceImpl struct {
	A uint8 `serix:""`
	B uint8 `serix:""`
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
	// ToDo: implement me
	panic("implement me")
}

func (ii *InterfaceImpl) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (ii *InterfaceImpl) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (ii *InterfaceImpl) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WriteNum(interfaceImplObjectCode, defaultErrProducer)
	seri.WriteNum(ii.A, defaultErrProducer)
	seri.WriteNum(ii.B, defaultErrProducer)

	return seri.Serialize()
}

type StructWithInterface struct {
	Interface Interface `serix:""`
}

func (si StructWithInterface) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (si StructWithInterface) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (si StructWithInterface) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (si StructWithInterface) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WriteObject(si.Interface.(serializer.Serializable), defaultSeriMode, deSeriCtx, defaultWriteGuard, defaultErrProducer)

	return seri.Serialize()
}

type StructWithOptionalField struct {
	Optional *ExportedStruct `serix:",optional"`
}

func (so StructWithOptionalField) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (so StructWithOptionalField) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (so StructWithOptionalField) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (so StructWithOptionalField) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WritePayloadLength(0, defaultErrProducer)

	return seri.Serialize()
}

type StructWithEmbeddedStructs struct {
	unexportedStruct `serix:""`
	ExportedStruct   `serix:",inlined"`
}

func (se StructWithEmbeddedStructs) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (se StructWithEmbeddedStructs) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	seri := serializer.NewSerializer()
	seri.WriteNum(se.unexportedStruct.Foo, defaultErrProducer)
	seri.WriteNum(exportedStructObjectCode, defaultErrProducer)
	seri.WriteNum(se.ExportedStruct.Bar, defaultErrProducer)

	return seri.Serialize()
}

type unexportedStruct struct {
	Foo uint64 `serix:""`
}

type ExportedStruct struct {
	Bar uint64 `serix:""`
}

func (e ExportedStruct) SetDeserializationContext(ctx context.Context) {
	ctx.Value("contextValue").(func())()
}

var exportedStructObjectCode = uint32(3)

type Map map[uint64]uint64

var mapLenType = serix.LengthPrefixTypeAsUint32

func (m Map) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (m Map) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (m Map) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
	panic("implement me")
}

func (m Map) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	bytes := make([][]byte, len(m))
	var i int
	for k, v := range m {
		seri := serializer.NewSerializer()
		seri.WriteNum(k, defaultErrProducer)
		seri.WriteNum(v, defaultErrProducer)
		b, err := seri.Serialize()
		if err != nil {
			return nil, err
		}
		bytes[i] = b
		i++
	}
	seri := serializer.NewSerializer()
	mode := defaultSeriMode | serializer.DeSeriModePerformLexicalOrdering
	arrayRules := &serializer.ArrayRules{ValidationMode: serializer.ArrayValidationModeLexicalOrdering}
	seri.WriteSliceOfByteSlices(bytes, mode, serializer.SeriLengthPrefixType(mapLenType), arrayRules, defaultErrProducer)

	return seri.Serialize()
}

type CustomSerializable int

func (cs CustomSerializable) SetDeserializationContext(ctx context.Context) {
	ctx.Value("contextValue").(func())()
}

func (cs CustomSerializable) MarshalJSON() ([]byte, error) {
	// ToDo: implement me
	panic("implement me")
}

func (cs CustomSerializable) UnmarshalJSON(bytes []byte) error {
	// ToDo: implement me
	panic("implement me")
}

func (cs CustomSerializable) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	// ToDo: implement me
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

var errSyntacticValidation = ierrors.New("syntactic validation failed")

func SyntacticValidation(ctx context.Context, obj ObjectForSyntacticValidation) error {
	return errSyntacticValidation
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
		if err := testAPI.RegisterValidator(ObjectForSyntacticValidation{}, SyntacticValidation); err != nil {
			log.Panic(err)
		}

		return m.Run()
	}()
	os.Exit(exitCode)
}

func BenchmarkEncode(b *testing.B) {
	simpleStruct := NewSimpleStruct()
	for i := 0; i < b.N; i++ {
		testAPI.Encode(context.Background(), simpleStruct)
	}
}

func BenchmarkDecode(b *testing.B) {
	simpleStruct := NewSimpleStruct()
	encoded, err := testAPI.Encode(context.Background(), simpleStruct)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		testAPI.Decode(context.Background(), encoded, new(SimpleStruct))
	}
}
