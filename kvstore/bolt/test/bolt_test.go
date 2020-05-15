package test

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/bolt"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"io/ioutil"
	"strconv"
	"testing"
)

func testStore(t require.TestingT, prefix []byte) kvstore.KVStore {
	dir, err := ioutil.TempDir("", "database.boltdb")
	require.NoError(t, err)
	dirAndFile := fmt.Sprintf("%s/my.db", dir)
	db, err := bbolt.Open(dirAndFile, 0666, nil)
	require.NoError(t, err)
	return bolt.New(db).WithRealm(prefix)
}

func TestSetAndGet(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testStore(t, prefix)

	key := []byte("testKey")
	expectedValue := []byte("testValue")

	err := db.Set(key, expectedValue)
	require.NoError(t, err)

	value, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(expectedValue, value))
}

func TestSetAndGetEmptyValue(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testStore(t, prefix)

	key := []byte("testKey")
	expectedValue := []byte{}

	err := db.Set(key, expectedValue)

	require.NoError(t, err)

	value, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(expectedValue, value))
}

func TestDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testStore(t, prefix)

	key := []byte("testKey")
	expectedValue := []byte("testValue")

	err := db.Set(key, expectedValue)
	require.NoError(t, err)

	value, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(expectedValue, value))

	err = db.Delete(key)
	require.NoError(t, err)

	_, err = db.Get(key)
	require.Error(t, kvstore.ErrKeyNotFound)
}

func TestIterate(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 100

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	err := store.Iterate([]kvstore.KeyPrefix{}, true, func(key kvstore.Key, value kvstore.Value) bool {
		expectedValue, found := insertedValues[string(key)]
		require.True(t, found)
		require.Equal(t, expectedValue, string(value))
		delete(insertedValues, string(key))
		return true
	})
	require.NoError(t, err)

	require.Equal(t, 0, len(insertedValues))
}

func TestIteratePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 100

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := store.Set([]byte("someOtherKey"+str), []byte(str))
		require.NoError(t, err)
	}

	err := store.Iterate([]kvstore.KeyPrefix{[]byte("testKey")}, true, func(key kvstore.Key, value kvstore.Value) bool {
		expectedValue, found := insertedValues[string(key)]
		require.True(t, found)
		require.Equal(t, expectedValue, string(value))
		delete(insertedValues, string(key))
		return true
	})

	require.NoError(t, err)

	require.Equal(t, 0, len(insertedValues))
}

func TestIteratePrefixKeyOnly(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 100

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := store.Set([]byte("someOtherKey"+str), []byte(str))
		require.NoError(t, err)
	}

	err := store.IterateKeys([]kvstore.KeyPrefix{[]byte("testKey")}, func(key kvstore.Key) bool {
		_, found := insertedValues[string(key)]
		require.True(t, found)
		delete(insertedValues, string(key))
		return true
	})
	require.NoError(t, err)

	require.Equal(t, 0, len(insertedValues))
}

func TestDeletePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 1000

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := store.Set([]byte("someOtherKey"+str), []byte(str))
		require.NoError(t, err)
	}

	err := store.DeletePrefix([]byte("someOtherKey"))
	require.NoError(t, err)

	// Verify, that the database only contains the elements without the delete prefix
	err = store.Iterate([]kvstore.KeyPrefix{}, true, func(key kvstore.Key, value kvstore.Value) bool {

		expectedValue, found := insertedValues[string(key)]
		require.True(t, found)
		require.Equal(t, expectedValue, string(value))
		delete(insertedValues, string(key))
		return true
	})
	require.NoError(t, err)

	require.Equal(t, 0, len(insertedValues))
}

func TestDeletePrefixIsEmpty(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 100

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err)
	}

	err := store.DeletePrefix([]byte{})
	require.NoError(t, err)

	// Verify, that the database does not contain any items since we deleted using the prefix
	err = store.Iterate([]kvstore.KeyPrefix{}, true, func(key kvstore.Key, value kvstore.Value) bool {
		t.Fail()
		return true
	})
	require.NoError(t, err)
}

func TestSetAndOverwrite(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)
	count := 100

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		err := store.Set([]byte(testKey), []byte{0})
		require.NoError(t, err)
	}

	verifyCount := 0
	// Verify that all entries are 0
	err := store.Iterate([]kvstore.KeyPrefix{}, true, func(key kvstore.Key, value kvstore.Value) bool {
		require.True(t, bytes.Equal([]byte{0}, value))
		verifyCount = verifyCount + 1
		return true
	})
	require.NoError(t, err)

	// Check that we checked the correct amount of entries
	require.Equal(t, count, verifyCount)

	batch := store.Batched()

	// Batch edit all to value 1
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		batch.Set([]byte(testKey), []byte{1})
	}

	err = batch.Commit()
	require.NoError(t, err)

	verifyCount = 0
	// Verify, that all entries were changed
	err = store.Iterate([]kvstore.KeyPrefix{}, true, func(key kvstore.Key, value kvstore.Value) bool {
		require.True(t, bytes.Equal([]byte{1}, value))
		verifyCount++
		return true
	})
	require.NoError(t, err)

	// Check that we checked the correct amount of entries
	require.Equal(t, count, verifyCount)
}

func TestBatchedWithSetAndDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)

	err := store.Set([]byte("testKey1"), []byte{42})
	require.NoError(t, err)

	err = store.Set([]byte("testKey2"), []byte{13})
	require.NoError(t, err)

	batch := store.Batched()

	batch.Set([]byte("testKey1"), []byte{84})

	batch.Set([]byte("testKey3"), []byte{69})

	batch.Delete([]byte("testKey2"))

	err = batch.Commit()
	require.NoError(t, err)

	err = store.Iterate([]kvstore.KeyPrefix{[]byte("testKey")}, true, func(key kvstore.Key, value kvstore.Value) bool {
		if string(key) == "testKey1" {
			require.True(t, bytes.Equal(value, []byte{84}))
		} else if string(key) == "testKey3" {
			require.True(t, bytes.Equal(value, []byte{69}))
		} else {
			t.Fail()
		}
		return true
	})
	require.NoError(t, err)
}

func TestBatchedWithDuplicateKeys(t *testing.T) {

	prefix := []byte("testPrefix")
	store := testStore(t, prefix)

	batch := store.Batched()

	batch.Set([]byte("testKey1"), []byte{84})
	batch.Set([]byte("testKey1"), []byte{69})

	err := batch.Commit()
	require.NoError(t, err)

	err = store.Iterate([]kvstore.KeyPrefix{[]byte("testKey")}, true, func(key kvstore.Key, value kvstore.Value) bool {
		if string(key) == "testKey1" {
			require.True(t, bytes.Equal(value, []byte{69}))
		} else {
			t.Fail()
		}
		return true
	})
	require.NoError(t, err)
}
