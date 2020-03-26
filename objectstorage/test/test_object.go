package test

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TestObject struct {
	objectstorage.StorableObjectFlags
	sync.Mutex

	id    []byte
	value uint32
}

func NewTestObject(id string, value uint32) *TestObject {
	return &TestObject{
		id:    []byte(id),
		value: value,
	}
}

func (testObject *TestObject) ObjectStorageKey() []byte {
	return testObject.id
}

func (testObject *TestObject) ObjectStorageValue() []byte {
	result := make([]byte, 4)

	binary.LittleEndian.PutUint32(result, testObject.value)

	return result
}

func (testObject *TestObject) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*TestObject); !ok {
		panic("invalid object passed to testObject.Update()")
	} else {
		testObject.value = obj.value
	}
}

func (testObject *TestObject) UnmarshalObjectStorageValue(data []byte) (err error, consumedBytes int) {
	testObject.value = binary.LittleEndian.Uint32(data)

	return nil, marshalutil.UINT32_SIZE
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

func (t ThreeLevelObj) ObjectStorageKey() []byte {
	var b bytes.Buffer
	b.WriteByte(t.id)
	b.WriteByte(t.id2)
	b.WriteByte(t.id3)
	return b.Bytes()
}

func (t ThreeLevelObj) ObjectStorageValue() []byte {
	return []byte{t.id3}
}

func (t ThreeLevelObj) UnmarshalObjectStorageValue(data []byte) (error, int) {
	t.id3 = data[0]
	return nil, len(data)
}
