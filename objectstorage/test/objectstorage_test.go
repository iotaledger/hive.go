package test

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/objectstorage/boltdb"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func testDatabase(t require.TestingT) objectstorage.Storage {
	dir, err := ioutil.TempDir("", "bboltdb")
	require.NoError(t, err)
	dirAndFile := fmt.Sprintf("%s/my.db", dir)
	db, err := bbolt.Open(dirAndFile, 0666, nil)
	require.NoError(t, err)
	return boltdb.New(db)
	/*
		dir, err := ioutil.TempDir("", "objectsdb")
		require.NoError(t, err)
		db, err := database.CreateDB(dir)
		require.NoError(t, err)
		return badgerstorage.New(db)
	*/
}

func testObjectFactory(key []byte) (objectstorage.StorableObject, int, error) {
	return &TestObject{id: key}, len(key), nil
}

func TestPrefixIteration(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestStoreIfAbsentStorage"), testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	storedObject1, _ := objects.StoreIfAbsent(NewTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(NewTestObject("13", 33))
	storedObject2.Release()

	storedObject3 := objects.Load([]byte("12"))
	storedObject3.Release()

	expectedKeys := make(map[string]types.Empty)

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void
	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()
		return true
	})

	assert.Equal(t, 0, len(expectedKeys))

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void
	objects.ForEachKeyOnly(func(key []byte) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		return true
	}, false)

	assert.Equal(t, 0, len(expectedKeys))

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void
	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()
		return true
	}, []byte(""))

	assert.Equal(t, 0, len(expectedKeys))

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void
	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()
		return true
	}, []byte("1"))

	assert.Equal(t, 0, len(expectedKeys))

	expectedKeys["12"] = types.Void
	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()
		return true
	}, []byte("12"))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestDeletionWithMoreThanTwoPartitions(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("Nakamoto"), testObjectFactory,
		objectstorage.PartitionKey(1, 1, 1),
		objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	cachedObj, _ := objects.StoreIfAbsent(NewThreeLevelObj(65, 66, 67))
	cachedObj.Release()

	sizeBeforeFlush := objects.GetSize()
	if sizeBeforeFlush != 1 {
		t.Fatalf("expected object storage size to be 1 but was %d", sizeBeforeFlush)
	}

	objects.Flush()
	sizeAfterFlush := objects.GetSize()
	if sizeAfterFlush != 0 {
		t.Fatalf("expected object storage size to be zero but was %d", sizeAfterFlush)
	}
}

func TestStorableObjectFlags(t *testing.T) {
	testObject := NewTestObject("Batman", 44)

	assert.Equal(t, false, testObject.IsModified())
	testObject.SetModified()
	assert.Equal(t, true, testObject.IsModified())
	testObject.SetModified(false)
	assert.Equal(t, false, testObject.IsModified())
	testObject.SetModified(true)
	assert.Equal(t, true, testObject.IsModified())

	assert.Equal(t, false, testObject.IsDeleted())
	testObject.Delete()
	assert.Equal(t, true, testObject.IsDeleted())
	testObject.Delete(false)
	assert.Equal(t, false, testObject.IsDeleted())
	testObject.Delete(true)
	assert.Equal(t, true, testObject.IsDeleted())

	assert.Equal(t, false, testObject.PersistenceEnabled())
	testObject.Persist()
	assert.Equal(t, true, testObject.PersistenceEnabled())
	testObject.Persist(false)
	assert.Equal(t, false, testObject.PersistenceEnabled())
	testObject.Persist(true)
	assert.Equal(t, true, testObject.PersistenceEnabled())
}

func BenchmarkStore(b *testing.B) {

	// create our storage
	objects := objectstorage.New(testDatabase(b), []byte("TestObjectStorage"), testObjectFactory)
	if err := objects.Prune(); err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		objects.Store(NewTestObject("Hans"+strconv.Itoa(i), uint32(i))).Release()
	}

	objects.Shutdown()
}

func BenchmarkLoad(b *testing.B) {
	objects := objectstorage.New(testDatabase(b), []byte("TestObjectStorage"), testObjectFactory)

	for i := 0; i < b.N; i++ {
		objects.Store(NewTestObject("Hans"+strconv.Itoa(i), uint32(i))).Release()
	}

	time.Sleep(2 * time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cachedObject := objects.Load([]byte("Hans" + strconv.Itoa(i)))

		cachedObject.Release()
	}
}

func BenchmarkLoadCachingEnabled(b *testing.B) {
	objects := objectstorage.New(testDatabase(b), []byte("TestObjectStorage"), testObjectFactory, objectstorage.CacheTime(500*time.Millisecond))

	for i := 0; i < b.N; i++ {
		objects.Store(NewTestObject("Hans"+strconv.Itoa(0), uint32(i)))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cachedObject := objects.Load([]byte("Hans" + strconv.Itoa(0)))

		cachedObject.Release()
	}
}

func TestStoreIfAbsent(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestStoreIfAbsentStorage"), testObjectFactory)
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	loadedObject := objects.Load([]byte("Hans"))
	loadedObject.Release()

	storedObject1, stored1 := objects.StoreIfAbsent(NewTestObject("Hans", 33))
	assert.Equal(t, true, stored1)
	if typeutils.IsInterfaceNil(storedObject1) {
		t.Error("the object should NOT be nil if it was stored")
	}
	storedObject1.Release()

	storedObject2, stored2 := objects.StoreIfAbsent(NewTestObject("Hans", 33))
	assert.Equal(t, false, stored2)
	if !typeutils.IsInterfaceNil(storedObject2) {
		t.Error("the object should be nil if it wasn't stored")
	}

	objects.Shutdown()
}

func TestDelete(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestObjectStorage"), testObjectFactory)
	objects.Store(NewTestObject("Hans", 33)).Release()

	cachedObject := objects.Load([]byte("Hans"))
	if !cachedObject.Exists() {
		t.Error("the item should exist")
	}
	cachedObject.Release()

	objects.Delete([]byte("Hans"))
	objects.Delete([]byte("Huns"))

	cachedObject = objects.Load([]byte("Hans"))
	if cachedObject.Exists() {
		t.Error("the item should not exist exist")
	}
	cachedObject.Release()

	objects.Shutdown()
}

func TestConcurrency(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestObjectStorage"), testObjectFactory)
	objects.Store(NewTestObject("Hans", 33)).Release()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// make sure the 2nd goroutine "processes" the object first
		time.Sleep(time.Second)

		// check if we "see" the modifications of the 2nd goroutine (using the "consume" method)
		cachedObject.Consume(func(object objectstorage.StorableObject) {
			// test if the changes of the 2nd goroutine are visible
			if object.(*TestObject).get() != 3 {
				t.Error(errors.New("the modifications of the 2nd goroutine should be visible"))
			}
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// retrieve, modify and release the object manually (without consume)
		cachedObject.Get().(*TestObject).set(3)
		cachedObject.Release()
	}()

	wg.Wait()
}

func TestEvictionBug(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestObjectStorage"), testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true))

	testCount := 12001 // fails (if not, make the number bigger)

	// create the test objects
	wait := sync.WaitGroup{}
	wait.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func(i int) {
			objects.Store(NewTestObject(fmt.Sprintf("%v", i), 0)).Release()
			wait.Done()
		}(i)
	}
	wait.Wait()

	count := uint32(10)

	wait.Add(testCount * int(count))
	for i := 0; i < testCount; i++ {
		for j := 0; j < int(count); j++ {
			go func(i, j int) {
				cachedObject1 := objects.Load([]byte(fmt.Sprintf("%v", i)))
				cachedTestObject1 := cachedObject1.Get().(*TestObject)
				cachedTestObject1.Lock()
				cachedObject1.Get().(*TestObject).value++
				cachedTestObject1.Unlock()
				cachedTestObject1.SetModified(true)
				cachedObject1.Release()

				time.Sleep(time.Duration(1) * time.Millisecond)

				cachedObject2 := objects.Load([]byte(fmt.Sprintf("%v", i)))
				cachedTestObject2 := cachedObject2.Get().(*TestObject)
				cachedTestObject2.Lock()
				cachedObject2.Get().(*TestObject).value++
				cachedTestObject2.Unlock()
				cachedTestObject2.SetModified(true)
				cachedObject2.Release()
				wait.Done()
			}(i, j)
		}
	}
	wait.Wait()

	for i := testCount - 1; i >= 0; i-- {
		//time.Sleep(time.Duration(10) * time.Microsecond)
		cachedObject := objects.Load([]byte(fmt.Sprintf("%v", i)))
		if cachedObject.Get().(*TestObject).value != count*2 {
			t.Error(fmt.Errorf("Object %d: the modifications should be visible %d!=%d", i, cachedObject.Get().(*TestObject).value, count))

			return
		}
		cachedObject.Release()
	}
}

func TestDeleteAndCreate(t *testing.T) {
	objects := objectstorage.New(testDatabase(t), []byte("TestObjectStorage"), testObjectFactory)

	for i := 0; i < 5000; i++ {
		objects.Store(NewTestObject("Hans", 33)).Release()

		cachedObject := objects.Load([]byte("Hans"))
		if !cachedObject.Exists() {
			fmt.Println(cachedObject.Exists())
			t.Errorf("the item should exist: %d", i)
		}
		cachedObject.Release()

		objects.Delete([]byte("Hans"))
		objects.Delete([]byte("Huns"))

		cachedObject = objects.Load([]byte("Hans"))
		if cachedObject.Exists() {
			t.Errorf("the item should not exist: %d", i)
		}
		cachedObject.Release()

		newlyAdded := false
		cachedObject = objects.ComputeIfAbsent([]byte("Hans"), func(key []byte) objectstorage.StorableObject {
			newlyAdded = true
			return NewTestObject("Hans", 33)
		})
		cachedObject.Release()

		if !newlyAdded {
			t.Errorf("the item should not exist: %d", i)
		}
		objects.Delete([]byte("Hans"))
	}

	objects.Shutdown()
}
