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

type Slice16LengthType []Bool

func (sl Slice16LengthType) LengthPrefixType() serializer.SeriLengthPrefixType {
	return serializer.SeriLengthPrefixTypeAsUint16
}

func (sl Slice16LengthType) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	ser := serializer.NewSerializer()
	ser.WriteSliceOfObjects(sl, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, defaultArrayRules, defaultErrProducer)
	return ser.Serialize()
}

func (sl Slice16LengthType) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (sl Slice16LengthType) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (sl Slice16LengthType) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (sl Slice16LengthType) ToSerializables() serializer.Serializables {
	serializables := make(serializer.Serializables, len(sl))
	for i, b := range sl {
		serializables[i] = b
	}
	return serializables
}

func (sl Slice16LengthType) FromSerializables(seris serializer.Serializables) {
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

func (ss *SimpleStruct) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	ser := serializer.NewSerializer()
	ser.WriteBool(ss.Bool, defaultErrProducer)
	ser.WriteNum(ss.Num, defaultErrProducer)
	ser.WriteString(string(ss.String), serializer.SeriLengthPrefixTypeAsUint16, defaultErrProducer)
	return ser.Serialize()
}

func (ss *SimpleStruct) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (ss *SimpleStruct) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (ss *SimpleStruct) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}
