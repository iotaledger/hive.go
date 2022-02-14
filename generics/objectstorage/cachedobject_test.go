package objectstorage

import (
	"fmt"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/marshalutil"
)

func TestCachedObject_Consume(t *testing.T) {
	objectStorage := New[*testObject](mapdb.NewMapDB(), CacheTime(0))
	defer objectStorage.Shutdown()

	cachedStoredObject1, stored1 := objectStorage.StoreIfAbsent(NewTestObject(1, 3))
	assert.True(t, stored1)
	cachedStoredObject1.Release()

	cachedStoredObject2, stored2 := objectStorage.StoreIfAbsent(NewTestObject(3, 1337))
	assert.True(t, stored2)
	cachedStoredObject2.Release()

	time.Sleep(2 * time.Second)

	objectStorage.Load(marshalutil.New(marshalutil.Uint64Size).WriteUint64(3).Bytes()).Consume(func(t *testObject) {
		fmt.Println(t)
	})
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

func (t *testObject) FromBytes(bytes []byte) (storableObject StorableObject, err error) {
	marshalUtil := marshalutil.New(bytes)

	result := &testObject{}
	if result.key, err = marshalUtil.ReadUint64(); err != nil {
		return nil, errors.Errorf("failed to read key from MarshalUtil: %w", err)
	}
	if result.value, err = marshalUtil.ReadUint64(); err != nil {
		return nil, errors.Errorf("failed to read value from MarshalUtil: %w", err)
	}

	return result, nil
}

func (t *testObject) ObjectStorageKey() []byte {
	return marshalutil.New(marshalutil.Uint64Size).WriteUint64(t.key).Bytes()
}

func (t *testObject) ObjectStorageValue() []byte {
	return marshalutil.New(marshalutil.Uint64Size).WriteUint64(t.value).Bytes()
}

var _ StorableObject = &testObject{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
