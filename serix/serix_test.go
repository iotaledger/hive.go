package serix_test

import (
	"context"

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

type SliceLengthType16 []Bool

func (sl SliceLengthType16) LengthPrefixType() serializer.SeriLengthPrefixType {
	return serializer.SeriLengthPrefixTypeAsUint16
}

func (sl SliceLengthType16) ToSerializables() serializer.Serializables {
	serializables := make(serializer.Serializables, len(sl))
	for i, b := range sl {
		serializables[i] = b
	}
	return serializables
}

func (sl SliceLengthType16) FromSerializables(seris serializer.Serializables) {
	//TODO implement me
	panic("implement me")
}

type String16LengthType string

func (s String16LengthType) LengthPrefixType() serializer.SeriLengthPrefixType {
	return serializer.SeriLengthPrefixTypeAsUint16
}

type SimpleStruct struct {
	Bool   bool               `seri:"0"`
	Num    int64              `seri:"1"`
	String String16LengthType `seri:"2"`
}

type Interface interface {
	Method()
}

type InterfaceImpl struct{}

func (ii *InterfaceImpl) Method() {}

func (ii *InterfaceImpl) ObjectCode() interface{} {
	return uint32(1)
}

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
	ser.WriteNum(ii.ObjectCode(), defaultErrProducer)
	return ser.Serialize()
}

type StructWithInterface struct {
	Interface Interface `seri:"0"`
}

type BytesArray16 [16]byte

type BytesSliceLengthType32 []byte

func (b BytesSliceLengthType32) LengthPrefixType() serializer.SeriLengthPrefixType {
	//TODO implement me
	return serializer.SeriLengthPrefixTypeAsUint32
}
func init() {
	testAPI.RegisterObjects((*Interface)(nil), (*InterfaceImpl)(nil))
}
