package serix_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serix"
	"github.com/stretchr/testify/require"
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

func (ss SimpleStruct) ObjectCode() interface{} {
	return uint32(0)
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
	Interface Interface `serix:"0"`
}

type BytesArray16 [16]byte

type StructWithEmbeddedStructs struct {
	//unexportedStruct `serix:"0"`
	ExportedStruct `serix:"1,nest"`
}

type unexportedStruct struct {
	Foo int64 `serix:"0"`
}

type ExportedStruct struct {
	Bar int64 `serix:"0"`
}

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

func (es ExportedStruct) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	s := serializer.NewSerializer()
	s.WriteNum(es.ObjectCode(), defaultErrProducer)
	s.WriteNum(es.Bar, defaultErrProducer)
	return s.Serialize()
}

func (es ExportedStruct) ObjectCode() interface{} {
	return uint32(3)
}

func init() {
	testAPI.RegisterInterfaceObjects((*Interface)(nil), (*InterfaceImpl)(nil))
}

type S struct {
	SubS
	SubSs
	Foo string
}

type SubS struct {
}

func (ss SubS) MarshalJSON() ([]byte, error) {
	return json.Marshal("subs")
}

func (ss *SubS) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	return nil
}

type SubSs struct {
}

func (ss SubSs) MarshalJSON() ([]byte, error) {
	return json.Marshal("subss")
}

func (ss *SubSs) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	return nil
}

func TestTMP(t *testing.T) {
	s := S{SubS: SubS{}, Foo: "foo"}
	var i interface{}
	i = s
	m, ok := i.(json.Marshaler)
	t.Log(m, ok)
	b, err := json.Marshal(s)
	require.NoError(t, err)
	t.Log(string(b))
	newS := new(S)
	err = json.Unmarshal(b, newS)
	require.NoError(t, err)
	t.Logf("%+v", newS)
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)
	mm, ok := st.MethodByName("marshalJSON")
	t.Log(mm, ok)
	mmv := sv.MethodByName("marshalJSON")
	t.Log(mmv)

}
