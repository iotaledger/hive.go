package serix_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
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

func (sl Bools) ToSerializables() serializer.Serializables {
	serializables := make(serializer.Serializables, len(sl))
	for i, b := range sl {
		serializables[i] = b
	}
	return serializables
}

func (sl Bools) FromSerializables(seris serializer.Serializables) {
	//TODO implement me
	panic("implement me")
}

type SimpleStruct struct {
	Bool   bool   `serix:"0"`
	Num    int64  `serix:"1"`
	String string `serix:"2,lengthPrefixType=uint16"`
}

var simpleStructObjectCode = uint32(0)

type Interface interface {
	Method()
}

type InterfaceImpl struct{}

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
	//TODO implement me
	ser := serializer.NewSerializer()
	ser.WriteNum(interfaceImplObjectCode, defaultErrProducer)
	return ser.Serialize()
}

type StructWithOptionalField struct {
	Optional *ExportedStruct `serix:"0,optional"`
}

type StructWithInterface struct {
	Interface Interface `serix:"0"`
}

type BytesArray16 [16]byte

type StructWithEmbeddedStructs struct {
	unexportedStruct `serix:"0"`
	ExportedStruct   `serix:"1,nest"`
}

type unexportedStruct struct {
	Foo int64 `serix:"0"`
}

type ExportedStruct struct {
	Bar int64 `serix:"0"`
}

var exportedStructObjectCode = uint32(3)

func (es ExportedStruct) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (es ExportedStruct) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (es ExportedStruct) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

type CustomSerializable int

func (cs CustomSerializable) Encode() ([]byte, error) {
	b := []byte(fmt.Sprintf("int: %d", cs))
	return b, nil
}

type ObjectForSyntacticValidation struct{}

var errSyntacticValidation = errors.New("syntactic validation failed")

func SyntacticValidation(obj ObjectForSyntacticValidation) error {
	return errSyntacticValidation
}

type ObjectForBytesValidation struct{}

var errBytesValidation = errors.New("bytes validation failed")

func BytesValidation([]byte) error {
	return errBytesValidation
}

func TestMain(m *testing.M) {
	exitCode := func() int {
		if err := testAPI.RegisterTypeSettings(
			SimpleStruct{},
			serix.TypeSettings{}.WithObjectCode(simpleStructObjectCode),
		); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterTypeSettings(
			InterfaceImpl{},
			serix.TypeSettings{}.WithObjectCode(interfaceImplObjectCode),
		); err != nil {
			log.Panic(err)
		}
		if err := testAPI.RegisterTypeSettings(
			ExportedStruct{},
			serix.TypeSettings{}.WithObjectCode(exportedStructObjectCode),
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
