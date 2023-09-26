package test

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/badger"
	"github.com/izuc/zipp.foundation/kvstore/mapdb"
	"github.com/izuc/zipp.foundation/kvstore/pebble"
	"github.com/izuc/zipp.foundation/kvstore/testutil"
	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/objectstorage/typeutils"
	"github.com/izuc/zipp.foundation/runtime/workerpool"
)

const (
	dbBadger = iota
	dbMapDB
	dbPebble
)

const (
	usedDatabase = dbMapDB
)

func testStorage(t testing.TB, realm []byte) (kvstore.KVStore, error) {
	switch usedDatabase {
	case dbBadger:
		dir := t.TempDir()
		db, err := badger.CreateDB(dir)
		require.NoError(t, err)

		return badger.New(db).WithRealm(realm)

	case dbMapDB:
		return mapdb.NewMapDB().WithRealm(realm)

	case dbPebble:
		dir := t.TempDir()
		db, err := pebble.CreateDB(dir)
		require.NoError(t, err)

		return pebble.New(db).WithRealm(realm)
	}

	panic("unknown database")
}

func testObjectFactory(key []byte, data []byte) (objectstorage.StorableObject, error) {
	return &testObject{id: key, value: binary.LittleEndian.Uint32(data)}, nil
}

// TestConcurrentCreateDelete tests if ConsumeIfAbsent and Store can be used in parallel without breaking the
// ObjectStorage.
func TestConcurrentCreateDelete(t *testing.T) {
	// test parameters
	objectCount := 50000

	// create badger DB
	badgerDBMissingMessageStorage, err := testutil.BadgerDB(t)
	require.NoError(t, err)
	badgerDBMetadataStorage, err := testutil.BadgerDB(t)
	require.NoError(t, err)

	// create ObjectStorage instances
	missingMessageStorage := objectstorage.New(badgerDBMissingMessageStorage, testObjectFactory)
	metadataStorage := objectstorage.New(badgerDBMetadataStorage, testObjectFactory)

	// create workerpool
	wp := workerpool.New(t.Name(), 1024).Start()

	// result counters
	var eventsCounter int32

	var deletedMap sync.Map
	var createdMap sync.Map

	// spam calls with the defined amount of objects
	for i := 0; i < objectCount; i++ {
		// create a copy of the iteration variable (for the closures)
		x := uint32(i)
		messageIDString := strconv.Itoa(i)
		messageIDBytes := []byte(messageIDString)

		// launch the background worker that removes the missing message entry
		wp.Submit(func() {
			metadataStorage.ComputeIfAbsent(messageIDBytes, func(key []byte) objectstorage.StorableObject {
				cachedMissingMessage, stored := missingMessageStorage.StoreIfAbsent(newTestObject(messageIDString, x))
				if stored {
					createdMap.Store(string(key), "CREATED")

					cachedMissingMessage.Release()

					atomic.AddInt32(&eventsCounter, 1)
				}

				return nil
			}).Release()
		})

		// launch the background worker that creates the missing message entry
		wp.Submit(func() {
			metadataStorage.Store(newTestObject(messageIDString, x)).Release()

			if missingMessageStorage.DeleteIfPresent(messageIDBytes) {
				atomic.AddInt32(&eventsCounter, -1)

				deletedMap.Store(messageIDString, true)
			}
		})
	}

	// wait for workers to finish
	wp.Shutdown(false)
	wp.PendingTasksCounter.WaitIsZero()

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
}

// TestTransaction tests if Transactions with the same identifier can not run in parallel and that Transactions and
// RTransactions wait for each other.
func TestTransaction(t *testing.T) {
	storage, err := testStorage(t, []byte("TestStoreIfAbsentStorage"))
	require.NoError(t, err)

	// initialize ObjectStorage
	objects := objectstorage.New(storage, testObjectFactory)
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	// retrieve a CachedObject
	cachedObject := objects.Load([]byte("someObject"))

	// initialize variables to keep track of the execution order
	firstTransactionFinished := false
	rTransactionFinished := false

	// initialize WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// execute first Transaction with identifier 1
	wg.Add(1)
	go func() {
		cachedObject.Retain().Transaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)

			time.Sleep(200 * time.Millisecond)

			firstTransactionFinished = true
		}, 1)

		wg.Done()
	}()

	// make sure the second Transaction with identifier 1 executes after the first one
	wg.Add(1)
	go func() {
		// make the Transaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().Transaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, firstTransactionFinished, true)
		}, 1)

		wg.Done()
	}()

	// make sure the third Transaction with identifier 1 and 2 also waits for 1
	wg.Add(1)
	go func() {
		// make the Transaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().Transaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, firstTransactionFinished, true)
		}, 1, 2)

		wg.Done()
	}()

	// make sure the fourth Transaction with identifier 2 runs in parallel to number 1
	wg.Add(1)
	go func() {
		// make the Transaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().Transaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, firstTransactionFinished, false)
		}, 2)

		wg.Done()
	}()

	// make sure that RTransactions wait for Transactions to finish
	wg.Add(1)
	go func() {
		// make the RTransaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().RTransaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, firstTransactionFinished, true)
		}, 1)

		wg.Done()
	}()

	// run RTransaction with a new identifier and keep track of its execution order
	wg.Add(1)
	go func() {
		cachedObject.Retain().RTransaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)

			time.Sleep(200 * time.Millisecond)

			rTransactionFinished = true
		}, 4)

		wg.Done()
	}()

	// make sure that RTransactions can run simultaneously
	wg.Add(1)
	go func() {
		// make the RTransaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().RTransaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, rTransactionFinished, false)
		}, 4)

		wg.Done()
	}()

	// make sure that Transactions wait for RTransactions to finish
	wg.Add(1)
	go func() {
		// make the RTransaction start slightly later but while the first one is still running
		time.Sleep(100 * time.Millisecond)

		cachedObject.Retain().Transaction(func(object objectstorage.StorableObject) {
			assert.Equal(t, object, nil)
			assert.Equal(t, rTransactionFinished, true)
		}, 4)

		wg.Done()
	}()

	// wait for goroutines to finish
	wg.Wait()

	// release object and shutdown ObjectStorage
	cachedObject.Release()
	objects.Shutdown()
}

// TestComputeIfAbsentReturningNil tests if ComputeIfAbsent can return nil to simply execute some code if something is
// missing without interfering with consecutive StoreIfAbsent calls and without intersecting parallel ComputeIfAbsent
// calls.
func TestComputeIfAbsentReturningNil(t *testing.T) {
	// define test iterations
	testCount := 50

	// initialize ObjectStorage
	storage, err := testStorage(t, []byte("TestStoreIfAbsentStorage"))
	require.NoError(t, err)

	// initialize ObjectStorage
	objects := objectstorage.New(storage, testObjectFactory)
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
			cachedObject, stored := objects.StoreIfAbsent(newTestObject(objectStringKey, 33))
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
	storage, err := testStorage(t, []byte("TestStoreIfAbsentStorage"))
	require.NoError(t, err)

	// initialize ObjectStorage
	objects := objectstorage.New(storage, testObjectFactory, objectstorage.PartitionKey(1, 1), objectstorage.LeakDetectionEnabled(true))
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
	})

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
	}, objectstorage.WithIteratorPrefix([]byte("")))

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
	}, objectstorage.WithIteratorPrefix([]byte("1")))

	assert.Equal(t, 0, len(expectedKeys))

	expectedKeys["12"] = types.Void
	objects.ForEach(func(key []byte, cachedObject objectstorage.CachedObject) bool {
		if _, elementExists := expectedKeys[string(key)]; !elementExists {
			t.Error("found an unexpected key")
		}

		delete(expectedKeys, string(key))
		cachedObject.Release()

		return true
	}, objectstorage.WithIteratorPrefix([]byte("12")))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestDeletionWithMoreThanTwoPartitions(t *testing.T) {
	storage, err := testStorage(t, []byte("Nakamoto"))
	require.NoError(t, err)

	// initialize ObjectStorage
	objects := objectstorage.New(storage, testObjectFactory,
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
	o := newTestObject("Batman", 44)

	assert.Equal(t, false, o.IsModified())
	o.SetModified(true)
	assert.Equal(t, true, o.IsModified())
	o.SetModified(false)
	assert.Equal(t, false, o.IsModified())
	o.SetModified(true)
	assert.Equal(t, true, o.IsModified())

	assert.Equal(t, false, o.IsDeleted())
	o.Delete(true)
	assert.Equal(t, true, o.IsDeleted())
	o.Delete(false)
	assert.Equal(t, false, o.IsDeleted())
	o.Delete(true)
	assert.Equal(t, true, o.IsDeleted())

	assert.Equal(t, false, o.ShouldPersist())
	o.Persist(true)
	assert.Equal(t, true, o.ShouldPersist())
	o.Persist(false)
	assert.Equal(t, false, o.ShouldPersist())
	o.Persist(true)
	assert.Equal(t, true, o.ShouldPersist())
}

func BenchmarkStore(b *testing.B) {
	storage, err := testStorage(b, []byte("TestObjectStorage"))
	require.NoError(b, err)

	// create our storage
	objects := objectstorage.New(storage, testObjectFactory)
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
	storage, err := testStorage(b, []byte("TestObjectStorage"))
	require.NoError(b, err)

	objects := objectstorage.New(storage, testObjectFactory)

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
	storage, err := testStorage(b, []byte("TestObjectStorage"))
	require.NoError(b, err)

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.CacheTime(500*time.Millisecond))

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
	storage, err := testStorage(t, []byte("TestStoreIfAbsentStorage"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory)
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
	storage, err := testStorage(t, []byte("TestStoreOnCreation"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.StoreOnCreation(false), objectstorage.CacheTime(2*time.Second))
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
	storage, err = testStorage(t, []byte("TestStoreOnCreation"))
	require.NoError(t, err)

	objects = objectstorage.New(storage, testObjectFactory, objectstorage.StoreOnCreation(true), objectstorage.CacheTime(2*time.Second))
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
	storage, err := testStorage(t, []byte("TestObjectStorage"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory)
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
	storage, err := testStorage(t, []byte("TestObjectStorage"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory)
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
		storage, err := testStorage(t, []byte("TestObjectStorage"))
		require.NoError(t, err)

		objectsStorage := objectstorage.New(storage, testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true), objectstorage.LeakDetectionEnabled(true, objectstorage.LeakDetectionOptions{
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
	storage, err := testStorage(t, []byte("TestObjectStorage"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory, objectstorage.CacheTime(0), objectstorage.PersistenceEnabled(true))

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
			go func(i int) {
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
			}(i)
		}
	}
	wait.Wait()

	for i := testCount - 1; i >= 0; i-- {
		cachedObject := objects.Load([]byte(fmt.Sprintf("%v", i)))
		if cachedObject.Get().(*testObject).value != count*2 {
			t.Error(fmt.Errorf("Object %d: the modifications should be visible %d!=%d", i, cachedObject.Get().(*testObject).value, count*2))

			return
		}
		cachedObject.Release()
	}
}

func TestDeleteAndCreate(t *testing.T) {
	storage, err := testStorage(t, []byte("TestObjectStorage"))
	require.NoError(t, err)

	objects := objectstorage.New(storage, testObjectFactory)

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
	storage, err := testStorage(t, []byte("TestForEachWithPrefix"))
	require.NoError(t, err)

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
	}, objectstorage.WithIteratorPrefix([]byte("1")))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestForEachKeyOnlyWithPrefix(t *testing.T) {
	storage, err := testStorage(t, []byte("TestForEachKeyOnlyWithPrefix"))
	require.NoError(t, err)

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
	}, objectstorage.WithIteratorPrefix([]byte("1")))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}

func TestForEachKeyOnlySkippingCacheWithPrefix(t *testing.T) {
	storage, err := testStorage(t, []byte("TestPrefixIterationWithPrefixSkippingCache"))
	require.NoError(t, err)

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
	}, objectstorage.WithIteratorPrefix([]byte("1")), objectstorage.WithIteratorSkipCache(true))

	assert.Equal(t, 0, len(expectedKeys))

	objects.Shutdown()
}
