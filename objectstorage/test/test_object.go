package test

import (
	"bytes"
	"encoding/binary"

	"github.com/iotaledger/hive.go/objectstorage"
)

type TestObject struct {
	objectstorage.StorableObjectFlags

	id    []byte
	value uint32
}

func NewTestObject(id string, value uint32) *TestObject {
	return &TestObject{
		id:    []byte(id),
		value: value,
	}
}

func (testObject *TestObject) GetStorageKey() []byte {
	return testObject.id
}

func (testObject *TestObject) MarshalBinary() ([]byte, error) {
	result := make([]byte, 4)

	binary.LittleEndian.PutUint32(result, testObject.value)

	return result, nil
}

func (testObject *TestObject) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*TestObject); !ok {
		panic("invalid object passed to testObject.Update()")
	} else {
		testObject.value = obj.value
	}
}

func (testObject *TestObject) UnmarshalBinary(data []byte) error {
	testObject.value = binary.LittleEndian.Uint32(data)

	return nil
}

// ThreeLevelObj is an object stored on a 3 partition chunked object storage.
// ID3 corresponds to ThreeLevelObj's value.
type ThreeLevelObj struct {
	objectstorage.StorableObjectFlags
	id  byte
	id2 byte
	id3 byte
}

func NewThreeLevelObj(id1 byte, id2 byte, id3Value byte) *ThreeLevelObj {
	return &ThreeLevelObj{
		id:  id1,
		id2: id2,
		id3: id3Value,
	}
}

func (t ThreeLevelObj) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*ThreeLevelObj); !ok {
		panic("invalid object passed to ThreeLevelObj.Update()")
	} else {
		t.id3 = obj.id3
	}
}

func (t ThreeLevelObj) GetStorageKey() []byte {
	var b bytes.Buffer
	b.WriteByte(t.id)
	b.WriteByte(t.id2)
	b.WriteByte(t.id3)
	return b.Bytes()
}

func (t ThreeLevelObj) MarshalBinary() (data []byte, err error) {
	return []byte{t.id3}, nil
}

func (t ThreeLevelObj) UnmarshalBinary(data []byte) error {
	t.id3 = data[0]
	return nil
}
