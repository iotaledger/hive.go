//nolint:revive // we don't care about these linters in test cases
package generic

import (
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/kvstore/mapdb"
	"github.com/izuc/zipp.foundation/serializer/v2/byteutils"
	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

func TestNewStructStorage(t *testing.T) {
	objectStorage := NewStructStorage[testObject](mapdb.NewMapDB(), CacheTime(0))
	defer objectStorage.Shutdown()

	cachedStoredObject1, stored1 := objectStorage.StoreIfAbsent(NewTestObject(1, 3))
	assert.True(t, stored1)
	cachedStoredObject1.Release()

	cachedStoredObject2, stored2 := objectStorage.StoreIfAbsent(NewTestObject(3, 1337))
	assert.True(t, stored2)
	cachedStoredObject2.Release()

	time.Sleep(2 * time.Second)

	objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(3).Bytes()).Consume(func(testObject *testObject) {
		assert.Equal(t, uint64(1337), testObject.value)
	})
	load := objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(4).Bytes())

	_, exists1 := load.Unwrap()
	assert.False(t, exists1)
	load.Release()

	load = objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(3).Bytes())
	value, exists2 := load.Unwrap()
	assert.True(t, exists2)
	assert.Equal(t, uint64(1337), value.value)
	load.Release()

	load = objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(1).Bytes())
	objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(1).Bytes()).Consume(func(testObject *testObject) {
		testObject.Delete()
	})
	_, exists3 := load.Unwrap()
	assert.False(t, exists3)
	load.Release()
}

func TestNewInterfaceStorage(t *testing.T) {
	objectStorage := NewInterfaceStorage[testInterface](mapdb.NewMapDB(), func(key []byte, data []byte) (result StorableObject, err error) {
		r := new(testObject)
		if err = r.FromObjectStorage(key, data); err != nil {
			return nil, err
		}

		return r, nil
	}, CacheTime(0))
	defer objectStorage.Shutdown()

	cachedStoredObject1, stored1 := objectStorage.StoreIfAbsent(NewTestObject(1, 3))
	assert.True(t, stored1)
	cachedStoredObject1.Release()

	cachedStoredObject2, stored2 := objectStorage.StoreIfAbsent(NewTestObject(3, 1337))
	assert.True(t, stored2)
	cachedStoredObject2.Release()

	time.Sleep(2 * time.Second)

	objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(3).Bytes()).Consume(func(i testInterface) {
		assert.Equal(t, uint64(1337), i.(*testObject).value)
	})
	load := objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(4).Bytes())

	_, exists1 := load.Unwrap()
	assert.False(t, exists1)
	load.Release()

	load = objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(3).Bytes())
	value, exists2 := load.Unwrap()
	assert.True(t, exists2)
	assert.Equal(t, uint64(1337), value.(*testObject).value)
	load.Release()

	load = objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(1).Bytes())
	objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(1).Bytes()).Consume(func(testObject testInterface) {
		testObject.Delete()
	})
	_, exists3 := load.Unwrap()
	assert.False(t, exists3)
	load.Release()
}

// region testObject ///////////////////////////////////////////////////////////////////////////////////////////////////

type testObject struct {
	key   uint64
	value uint64

	StorableObjectFlags
}

func NewTestObject(key, value uint64) *testObject {
	return &testObject{
		key:   key,
		value: value,
	}
}

func (t *testObject) FromObjectStorage(key, value []byte) (err error) {
	return t.FromBytes(byteutils.ConcatBytes(key, value))

}
func (t *testObject) FromBytes(bytes []byte) (err error) {
	marshalUtil := marshalutil.New(bytes)
	if t.key, err = marshalUtil.ReadUint64(); err != nil {
		return errors.Errorf("failed to read key from MarshalUtil: %w", err)
	}
	if t.value, err = marshalUtil.ReadUint64(); err != nil {
		return errors.Errorf("failed to read value from MarshalUtil: %w", err)
	}

	return nil
}

func (t *testObject) ObjectStorageKey() []byte {
	return marshalutil.New(marshalutil.Uint64Size).WriteUint64(t.key).Bytes()
}

func (t *testObject) ObjectStorageValue() []byte {
	return marshalutil.New(marshalutil.Uint64Size).WriteUint64(t.value).Bytes()
}

var _ StorableObject = &testObject{}

type testInterface interface {
	StorableObject
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
