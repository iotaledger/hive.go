package test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/badger"
	"github.com/iotaledger/hive.go/kvstore/bolt"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

const (
	DB_BADGER = iota
	DB_BOLT
	DB_MAPDB
)

const (
	usedDatabase = DB_MAPDB
)

func testStorage(t require.TestingT, realm []byte) kvstore.KVStore {

	switch usedDatabase {

	case DB_BADGER:
		dir, err := ioutil.TempDir("", "objectsdb")
		require.NoError(t, err)
		db, err := badger.CreateDB(dir)
		require.NoError(t, err)
		return badger.New(db).WithRealm(realm)

	case DB_BOLT:
		dir, err := ioutil.TempDir("", "bboltdb")
		require.NoError(t, err)
		dirAndFile := fmt.Sprintf("%s/my.db", dir)
		db, err := bbolt.Open(dirAndFile, 0666, nil)
		require.NoError(t, err)
		return bolt.New(db).WithRealm(realm)

	case DB_MAPDB:
		return mapdb.NewMapDB().WithRealm(realm)
	}

	panic("unknown database")
}

func testObjectFactory(key []byte) (objectstorage.StorableObject, int, error) {
	return &TestObject{id: key}, len(key), nil
}

// TestComputeIfAbsentReturningNil tests if ComputeIfAbsent can return nil to simply execute some code if something is
// missing without interfering with consecutive StoreIfAbsent calls and without intersecting parallel ComputeIfAbsent
// calls.
func TestComputeIfAbsentReturningNil(t *testing.T) {
	// define test iterations
	testCount := 50

	// initialize ObjectStorage
	objects := objectstorage.New(testStorage(t, []byte("TestStoreIfAbsentStorage")), testObjectFactory)
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	// repeat test the defined times
	for i := 0; i < testCount; i++ {
		objectStringKey := "missingEntry" + strconv.Itoa(i)

		// define variables to track the execution flow
		firstComputeIfAbsentExecutedOrder := -1
		firstComputeIfAbsentFinished := false
		secondComputeIfAbsentExecutedOrder := -1
		secondComputeIfAbsentFinished := false
		storeExecutedOrder := -1
		orderCounter := 0

		// initialize WaitGroup to wait for the finished goroutines
		var wg sync.WaitGroup

		// start the first ComputeIfAbsent call
		wg.Add(1)
		go func() {
			objects.ComputeIfAbsent([]byte(objectStringKey), func(key []byte) objectstorage.StorableObject {
				firstComputeIfAbsentExecutedOrder = orderCounter
				orderCounter++

				if secondComputeIfAbsentExecutedOrder != -1 {
					assert.Equal(t, secondComputeIfAbsentFinished, true)
				}

				time.Sleep(100 * time.Millisecond)

				firstComputeIfAbsentFinished = true

				return nil
			}).Release()

			wg.Done()
		}()

		// start the second ComputeIfAbsent call
		wg.Add(1)
		go func() {
			objects.ComputeIfAbsent([]byte(objectStringKey), func(key []byte) objectstorage.StorableObject {
				secondComputeIfAbsentExecutedOrder = orderCounter
				orderCounter++

				if firstComputeIfAbsentExecutedOrder != -1 {
					assert.Equal(t, firstComputeIfAbsentFinished, true)
				}

				time.Sleep(100 * time.Millisecond)

				secondComputeIfAbsentFinished = true

				return nil
			}).Release()

			wg.Done()
		}()

		// start the StoreIfAbsent call
		wg.Add(1)
		go func() {
			cachedObject, stored := objects.StoreIfAbsent(NewTestObject(objectStringKey, 33))
			cachedObject.Release()

			if assert.Equal(t, true, stored) {
				storeExecutedOrder = orderCounter
				orderCounter++
			}

			wg.Done()
		}()

		// wait for goroutines to finish
		wg.Wait()

		// make sure the result are as expected
		switch storeExecutedOrder {
		case 0:
			assert.Equal(t, firstComputeIfAbsentExecutedOrder, -1)
			assert.Equal(t, secondComputeIfAbsentExecutedOrder, -1)
			assert.True(t, !firstComputeIfAbsentFinished && !secondComputeIfAbsentFinished)
		case 1:
			assert.True(t, (firstComputeIfAbsentExecutedOrder == 0 && secondComputeIfAbsentExecutedOrder == -1) || (firstComputeIfAbsentExecutedOrder == -1 && secondComputeIfAbsentExecutedOrder == 0))
			assert.True(t, (firstComputeIfAbsentFinished && !secondComputeIfAbsentFinished) || (!firstComputeIfAbsentFinished && secondComputeIfAbsentFinished))
		case 2:
			assert.True(t, (firstComputeIfAbsentExecutedOrder == 0 && secondComputeIfAbsentExecutedOrder == 1) || (firstComputeIfAbsentExecutedOrder == 1 && secondComputeIfAbsentExecutedOrder == 0))
			assert.True(t, firstComputeIfAbsentFinished && secondComputeIfAbsentFinished)
		}
	}

	// shutdown the ObjectStorage
	objects.Shutdown()
}

func TestPrefixIteration(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestStoreIfAbsentStorage")), testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
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
	objects := objectstorage.New(testStorage(t, []byte("Nakamoto")), testObjectFactory,
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

	assert.Equal(t, false, testObject.ShouldPersist())
	testObject.Persist()
	assert.Equal(t, true, testObject.ShouldPersist())
	testObject.Persist(false)
	assert.Equal(t, false, testObject.ShouldPersist())
	testObject.Persist(true)
	assert.Equal(t, true, testObject.ShouldPersist())
}

func BenchmarkStore(b *testing.B) {

	// create our storage
	objects := objectstorage.New(testStorage(b, []byte("TestObjectStorage")), testObjectFactory)
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
	objects := objectstorage.New(testStorage(b, []byte("TestObjectStorage")), testObjectFactory)

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
	objects := objectstorage.New(testStorage(b, []byte("TestObjectStorage")), testObjectFactory, objectstorage.CacheTime(500*time.Millisecond))

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
	objects := objectstorage.New(testStorage(t, []byte("TestStoreIfAbsentStorage")), testObjectFactory)
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

func TestStoreOnCreation(t *testing.T) {
	//
	// without StoreOnCreation
	//
	objects := objectstorage.New(testStorage(t, []byte("TestStoreOnCreation")), testObjectFactory, objectstorage.StoreOnCreation(false), objectstorage.CacheTime(2*time.Second))
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

	// give the batchWriter some time to persist it
	time.Sleep(time.Second)

	if loadedObject := objects.LoadObjectFromStore([]byte("Hans")); !typeutils.IsInterfaceNil(loadedObject) {
		t.Error("the object should NOT be stored in the database yet stored")
	}

	storedObject1.Release(true)

	//
	// with StoreOnCreation
	//
	objects = objectstorage.New(testStorage(t, []byte("TestStoreOnCreation")), testObjectFactory, objectstorage.StoreOnCreation(true), objectstorage.CacheTime(2*time.Second))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	loadedObject = objects.Load([]byte("Hans"))
	loadedObject.Release()

	storedObject1, stored1 = objects.StoreIfAbsent(NewTestObject("Hans", 33))
	assert.Equal(t, true, stored1)

	if typeutils.IsInterfaceNil(storedObject1) {
		t.Error("the object should NOT be nil if it was stored")
	}

	// give the batchWriter some time to persist it
	time.Sleep(time.Second)

	if loadedObject := objects.LoadObjectFromStore([]byte("Hans")); typeutils.IsInterfaceNil(loadedObject) {
		t.Error("the object should NOT be nil if it was stored")
	}

	storedObject1.Release(true)

	objects.Shutdown()
}

func TestDelete(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory)
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
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory)
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

func TestStoreIfAbsentTriggersOnce(t *testing.T) {
	for k := 0; k < 10; k++ {
		// define test parameters
		objectCount := 200
		workerCount := 50

		// initialize object storage
		objectsStorage := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true), objectstorage.LeakDetectionEnabled(true, objectstorage.LeakDetectionOptions{
			MaxConsumersPerObject: 100,
			MaxConsumerHoldTime:   5 * time.Second,
		}))

		// prepare objects to store
		objects := make([]*TestObject, objectCount)
		for i := 0; i < objectCount; i++ {
			objects[i] = NewTestObject(fmt.Sprintf("%v", i), 0)
		}

		// store the same object multiple times in multiple goroutines
		var wg sync.WaitGroup
		var storedObjectsCount int32
		for i := 0; i < objectCount; i++ {
			for j := 0; j < workerCount; j++ {
				wg.Add(1)
				go func(i int) {
					storedObject, stored := objectsStorage.StoreIfAbsent(objects[i])
					if stored {
						atomic.AddInt32(&storedObjectsCount, 1)

						storedObject.Release()
					}

					wg.Done()
				}(i)
			}
		}

		// wait till storing the objects is done
		wg.Wait()
		objectsStorage.Shutdown()

		// evaluate results
		assert.Equal(t, objectCount, int(storedObjectsCount), "StoreIfAbsent should only return true for a single concurrent caller")
	}
}

func TestEvictionBug(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true))

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
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory)

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

func TestForEachWithPrefix(t *testing.T) {

	storage := testStorage(t, []byte("TestForEachWithPrefix"))

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	storedObject1, _ := objects.StoreIfAbsent(NewTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(NewTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(NewTestObject("23", 33))
	storedObject3.Release()

	// Store all to disk
	objects.Shutdown()

	// Setup the storage again with the same database
	objects = objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))

	expectedKeys := make(map[string]types.Empty)

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void

	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Errorf("found an unexpected key: '%v'", string(key))
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()
		return true
	}, []byte("1"))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestForEachKeyOnlyWithPrefix(t *testing.T) {

	storage := testStorage(t, []byte("TestForEachKeyOnlyWithPrefix"))

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	storedObject1, _ := objects.StoreIfAbsent(NewTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(NewTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(NewTestObject("23", 33))
	storedObject3.Release()

	// Store all to disk
	objects.Shutdown()

	// Setup the storage again with the same database
	objects = objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))

	expectedKeys := make(map[string]types.Empty)

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void

	objects.ForEachKeyOnly(func(key []byte) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Errorf("found an unexpected key: '%v'", string(key))
		}

		delete(expectedKeys, string(key))
		return true
	}, false, []byte("1"))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestForEachKeyOnlySkippingCacheWithPrefix(t *testing.T) {

	storage := testStorage(t, []byte("TestPrefixIterationWithPrefixSkippingCache"))

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	storedObject1, _ := objects.StoreIfAbsent(NewTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(NewTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(NewTestObject("23", 33))
	storedObject3.Release()

	// Store all to disk
	objects.Shutdown()

	// Setup the storage again with the same database
	objects = objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))

	expectedKeys := make(map[string]types.Empty)

	expectedKeys["12"] = types.Void
	expectedKeys["13"] = types.Void

	objects.ForEachKeyOnly(func(key []byte) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Errorf("found an unexpected key: '%v'", string(key))
		}

		delete(expectedKeys, string(key))
		return true
	}, true, []byte("1"))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}
