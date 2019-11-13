package objectstorage_test

import (
	"encoding/binary"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TestObject struct {
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
