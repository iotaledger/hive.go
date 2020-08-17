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

	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/badger"
	"github.com/iotaledger/hive.go/kvstore/bolt"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/testutil"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

const (
	dbBadger = iota
	dbBolt
	dbMapDB
)

const (
	usedDatabase = dbMapDB
)

func testStorage(t require.TestingT, realm []byte) kvstore.KVStore {
	switch usedDatabase {
	case dbBadger:
		dir, err := ioutil.TempDir("", "objectsdb")
		require.NoError(t, err)
		db, err := badger.CreateDB(dir)
		require.NoError(t, err)
		return badger.New(db).WithRealm(realm)

	case dbBolt:
		dir, err := ioutil.TempDir("", "bboltdb")
		require.NoError(t, err)
		dirAndFile := fmt.Sprintf("%s/my.db", dir)
		db, err := bbolt.Open(dirAndFile, 0666, nil)
		require.NoError(t, err)
		return bolt.New(db).WithRealm(realm)

	case dbMapDB:
		return mapdb.NewMapDB().WithRealm(realm)
	}

	panic("unknown database")
}

func testObjectFactory(key []byte) (objectstorage.StorableObject, int, error) {
	return &testObject{id: key}, len(key), nil
}

func TestConcurrentCreateDelete(t *testing.T) {
	// test parameters
	objectCount := 100000

	// create badger DB
	badgerDBMissingMessageStorage, err := testutil.BadgerDB(t)
	require.NoError(t, err)
	badgerDBMetadataStorage, err := testutil.BadgerDB(t)
	require.NoError(t, err)

	badgerDBMissingMessageStorage = testStorage(t, []byte("missingMessageStorage"))
	badgerDBMetadataStorage = testStorage(t, []byte("missingMessageStorage"))

	// create ObjectStorage instances
	missingMessageStorage := objectstorage.New(badgerDBMissingMessageStorage, testObjectFactory)
	metadataStorage := objectstorage.New(badgerDBMetadataStorage, testObjectFactory)

	// create sync and async utils
	var wp async.WorkerPool
	var wg sync.WaitGroup
	wp.Tune(1024)

	// result counters
	var eventsCounter int32

	// spam calls with the defined amount of objects
	for i := 0; i < objectCount; i++ {
		// create a copy of the iteration variable (for the closures)
		x := i

		// launch the background worker that removes the missing message entry
		wg.Add(1)
		wp.Submit(func() {
			messageIDString := strconv.Itoa(x)
			messageIDBytes := []byte(messageIDString)

			metadataStorage.ComputeIfAbsent(messageIDBytes, func(key []byte) objectstorage.StorableObject {
				cachedMissingMessage, stored := missingMessageStorage.StoreIfAbsent(newTestObject(messageIDString, uint32(x)))
				if stored {
					cachedMissingMessage.Release()

					atomic.AddInt32(&eventsCounter, 1)
				}

				return nil
			}).Release()

			wg.Done()
		})

		// launch the background worker that creates the missing message entry
		wg.Add(1)
		wp.Submit(func() {
			messageIDString := strconv.Itoa(x)
			messageIDBytes := []byte(messageIDString)

			metadataStorage.Store(newTestObject(messageIDString, uint32(x))).Release()

			if missingMessageStorage.DeleteIfPresent(messageIDBytes) {
				atomic.AddInt32(&eventsCounter, -1)
			}

			wg.Done()
		})
	}

	// wait for a workers to finish
	wg.Wait()

	// count messages still in the store
	messagesInStore := 0
	missingMessageStorage.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		messagesInStore++

		cachedObject.Release()

		return true
	})

	// check test results
	assert.Equal(t, int32(0), eventsCounter, "we should have seen and equal amount of create and delete events")
	assert.Equal(t, 0, messagesInStore, "the store should be empty")

	// shutdown test
	missingMessageStorage.Shutdown()
	metadataStorage.Shutdown()
	wp.Shutdown()
}

func TestPrefixIteration(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestStoreIfAbsentStorage")), testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	storedObject1, _ := objects.StoreIfAbsent(newTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(newTestObject("13", 33))
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

	cachedObj, _ := objects.StoreIfAbsent(newThreeLevelObj(65, 66, 67))
	cachedObj.Release()

	sizeBeforeFlush := objects.GetSize()
	if sizeBeforeFlush != 1 {
		t.Fatalf("expected testObject storage size to be 1 but was %d", sizeBeforeFlush)
	}

	objects.Flush()
	sizeAfterFlush := objects.GetSize()
	if sizeAfterFlush != 0 {
		t.Fatalf("expected testObject storage size to be zero but was %d", sizeAfterFlush)
	}
}

func TestStorableObjectFlags(t *testing.T) {
	testObject := newTestObject("Batman", 44)

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
		objects.Store(newTestObject("Hans"+strconv.Itoa(i), uint32(i))).Release()
	}

	objects.Shutdown()
}

func BenchmarkLoad(b *testing.B) {
	objects := objectstorage.New(testStorage(b, []byte("TestObjectStorage")), testObjectFactory)

	for i := 0; i < b.N; i++ {
		objects.Store(newTestObject("Hans"+strconv.Itoa(i), uint32(i))).Release()
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
		objects.Store(newTestObject("Hans"+strconv.Itoa(0), uint32(i)))
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

	storedObject1, stored1 := objects.StoreIfAbsent(newTestObject("Hans", 33))
	assert.Equal(t, true, stored1)
	if typeutils.IsInterfaceNil(storedObject1) {
		t.Error("the testObject should NOT be nil if it was stored")
	}
	storedObject1.Release()

	storedObject2, stored2 := objects.StoreIfAbsent(newTestObject("Hans", 33))
	assert.Equal(t, false, stored2)
	if !typeutils.IsInterfaceNil(storedObject2) {
		t.Error("the testObject should be nil if it wasn't stored")
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

	storedObject1, stored1 := objects.StoreIfAbsent(newTestObject("Hans", 33))
	assert.Equal(t, true, stored1)

	if typeutils.IsInterfaceNil(storedObject1) {
		t.Error("the testObject should NOT be nil if it was stored")
	}

	// give the batchWriter some time to persist it
	time.Sleep(time.Second)

	if loadedObject := objects.LoadObjectFromStore([]byte("Hans")); !typeutils.IsInterfaceNil(loadedObject) {
		t.Error("the testObject should NOT be stored in the database yet stored")
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

	storedObject1, stored1 = objects.StoreIfAbsent(newTestObject("Hans", 33))
	assert.Equal(t, true, stored1)

	if typeutils.IsInterfaceNil(storedObject1) {
		t.Error("the testObject should NOT be nil if it was stored")
	}

	// give the batchWriter some time to persist it
	time.Sleep(time.Second)

	if loadedObject := objects.LoadObjectFromStore([]byte("Hans")); typeutils.IsInterfaceNil(loadedObject) {
		t.Error("the testObject should NOT be nil if it was stored")
	}

	storedObject1.Release(true)

	objects.Shutdown()
}

func TestDelete(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory)
	objects.Store(newTestObject("Hans", 33)).Release()

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
	objects.Store(newTestObject("Hans", 33)).Release()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// make sure the 2nd goroutine "processes" the testObject first
		time.Sleep(time.Second)

		// check if we "see" the modifications of the 2nd goroutine (using the "consume" method)
		cachedObject.Consume(func(object objectstorage.StorableObject) {
			// test if the changes of the 2nd goroutine are visible
			if object.(*testObject).get() != 3 {
				t.Error(errors.New("the modifications of the 2nd goroutine should be visible"))
			}
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// retrieve, modify and release the testObject manually (without consume)
		cachedObject.Get().(*testObject).set(3)
		cachedObject.Release()
	}()

	wg.Wait()
}

func TestStoreIfAbsentTriggersOnce(t *testing.T) {
	for k := 0; k < 10; k++ {
		// define test parameters
		objectCount := 200
		workerCount := 50

		// initialize testObject storage
		objectsStorage := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true), objectstorage.LeakDetectionEnabled(true, objectstorage.LeakDetectionOptions{
			MaxConsumersPerObject: 100,
			MaxConsumerHoldTime:   5 * time.Second,
		}))

		// prepare objects to store
		objects := make([]*testObject, objectCount)
		for i := 0; i < objectCount; i++ {
			objects[i] = newTestObject(fmt.Sprintf("%v", i), 0)
		}

		// store the same testObject multiple times in multiple goroutines
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
			objects.Store(newTestObject(fmt.Sprintf("%v", i), 0)).Release()
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
				cachedTestObject1 := cachedObject1.Get().(*testObject)
				cachedTestObject1.Lock()
				cachedObject1.Get().(*testObject).value++
				cachedTestObject1.Unlock()
				cachedTestObject1.SetModified(true)
				cachedObject1.Release()

				time.Sleep(time.Duration(1) * time.Millisecond)

				cachedObject2 := objects.Load([]byte(fmt.Sprintf("%v", i)))
				cachedTestObject2 := cachedObject2.Get().(*testObject)
				cachedTestObject2.Lock()
				cachedObject2.Get().(*testObject).value++
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
		if cachedObject.Get().(*testObject).value != count*2 {
			t.Error(fmt.Errorf("Object %d: the modifications should be visible %d!=%d", i, cachedObject.Get().(*testObject).value, count))

			return
		}
		cachedObject.Release()
	}
}

func TestDeleteAndCreate(t *testing.T) {
	objects := objectstorage.New(testStorage(t, []byte("TestObjectStorage")), testObjectFactory)

	for i := 0; i < 5000; i++ {
		objects.Store(newTestObject("Hans", 33)).Release()

		cachedObject := objects.Load([]byte("Hans"))
		if !cachedObject.Exists() {
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
			return newTestObject("Hans", 33)
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

	storedObject1, _ := objects.StoreIfAbsent(newTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(newTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(newTestObject("23", 33))
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

	storedObject1, _ := objects.StoreIfAbsent(newTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(newTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(newTestObject("23", 33))
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

	storedObject1, _ := objects.StoreIfAbsent(newTestObject("12", 33))
	storedObject1.Release()

	storedObject2, _ := objects.StoreIfAbsent(newTestObject("13", 33))
	storedObject2.Release()

	storedObject3, _ := objects.StoreIfAbsent(newTestObject("23", 33))
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
