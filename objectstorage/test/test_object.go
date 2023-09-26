package test

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

type testObject struct {
	objectstorage.StorableObjectFlags
	sync.Mutex

	id    []byte
	value uint32
}

func newTestObject(id string, value uint32) *testObject {
	return &testObject{
		id:    []byte(id),
		value: value,
	}
}

func (t *testObject) ObjectStorageKey() []byte {
	return t.id
}

func (t *testObject) ObjectStorageValue() []byte {
	result := make([]byte, 4)

	t.Lock()
	defer t.Unlock()

	binary.LittleEndian.PutUint32(result, t.value)

	return result
}

func (t *testObject) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*testObject); !ok {
		panic("invalid testObject passed to testObject.Update()")
	} else {
		t.Lock()
		defer t.Unlock()

		t.value = obj.value
	}
}

func (t *testObject) UnmarshalObjectStorageValue(data []byte) (consumedBytes int, err error) {
	t.Lock()
	defer t.Unlock()

	t.value = binary.LittleEndian.Uint32(data)

	return marshalutil.Uint32Size, nil
}

func (t *testObject) get() uint32 {
	t.Lock()
	defer t.Unlock()

	return t.value
}

func (t *testObject) set(v uint32) {
	t.Lock()
	defer t.Unlock()
	t.value = v
}

// threeLevelObj is an testObject stored on a 3 partition chunked testObject storage.
// ID3 corresponds to threeLevelObj's value.
type threeLevelObj struct {
	objectstorage.StorableObjectFlags
	id  byte
	id2 byte
	id3 byte
}

func newThreeLevelObj(id1 byte, id2 byte, id3Value byte) *threeLevelObj {
	return &threeLevelObj{
		id:  id1,
		id2: id2,
		id3: id3Value,
	}
}

func (t *threeLevelObj) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*threeLevelObj); !ok {
		panic("invalid testObject passed to threeLevelObj.Update()")
	} else {
		t.id3 = obj.id3
	}
}

func (t *threeLevelObj) ObjectStorageKey() []byte {
	var b bytes.Buffer
	b.WriteByte(t.id)
	b.WriteByte(t.id2)
	b.WriteByte(t.id3)

	return b.Bytes()
}

func (t *threeLevelObj) ObjectStorageValue() []byte {
	return []byte{t.id3}
}

func (t *threeLevelObj) UnmarshalObjectStorageValue(data []byte) (int, error) {
	t.id3 = data[0]

	return len(data), nil
}
