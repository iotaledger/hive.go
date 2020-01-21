package test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/objectstorage"
)

func testObjectFactory(key []byte) objectstorage.StorableObject { return &TestObject{id: key} }

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
	objects := objectstorage.New([]byte("TestObjectStorage"), testObjectFactory)
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
	objects := objectstorage.New([]byte("TestObjectStorage"), testObjectFactory)

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
	objects := objectstorage.New([]byte("TestObjectStorage"), testObjectFactory, objectstorage.CacheTime(500*time.Millisecond))

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
	objects := objectstorage.New([]byte("TestStoreIfAbsentStorage"), testObjectFactory, objectstorage.CacheTime(1*time.Second))
	if err := objects.Prune(); err != nil {
		t.Error(err)
	}

	loadedObject := objects.Load([]byte("Hans"))
	loadedObject.Release()

	storedObject1, stored1 := objects.StoreIfAbsent([]byte("Hans"), NewTestObject("Hans", 33))
	assert.Equal(t, true, stored1)
	if storedObject1 == nil {
		t.Error("the object should NOT be nil if it was stored")
	}
	storedObject1.Release()

	storedObject2, stored2 := objects.StoreIfAbsent([]byte("Hans"), NewTestObject("Hans", 33))
	assert.Equal(t, false, stored2)
	if storedObject2 != nil {
		t.Error("the object should be nil if it wasn't stored")
	}

	objects.Shutdown()
}

func TestDelete(t *testing.T) {
	objects := objectstorage.New([]byte("TestObjectStorage"), testObjectFactory)
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
	objects := objectstorage.New([]byte("TestObjectStorage"), testObjectFactory)
	objects.Store(NewTestObject("Hans", 33)).Release()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// check if we "see" the modifications of the 2nd goroutine (using the "consume" method)
		cachedObject.Consume(func(object objectstorage.StorableObject) {
			// make sure the 2nd goroutine "processes" the object first
			time.Sleep(100 * time.Millisecond)

			// test if the changes of the 2nd goroutine are visible
			if object.(*TestObject).value != 3 {
				t.Error(errors.New("the modifications of the 2nd goroutine should be visible"))
			}
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cachedObject := objects.Load([]byte("Hans"))

		// retrieve, modify and release the object manually (without consume)
		cachedObject.Get().(*TestObject).value = 3
		cachedObject.Release()
	}()

	wg.Wait()
}
