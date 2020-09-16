package test

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/badger"
	"github.com/iotaledger/hive.go/kvstore/bolt"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/kvstore/pebble"
	"github.com/stretchr/testify/require"
)

const (
	dbBadger = iota
	dbBolt
	dbMapDB
	dbPebble
)

var (
	dbImplementations = []int{dbBadger, dbBolt, dbMapDB, dbPebble}
)

func testStore(t require.TestingT, dbImplementation int, realm []byte) kvstore.KVStore {
	switch dbImplementation {

	case dbBadger:
		dir, err := ioutil.TempDir("", "database.badger")
		require.NoError(t, err)
		db, err := badger.CreateDB(dir)
		require.NoError(t, err)
		return badger.New(db).WithRealm(realm)

	case dbBolt:
		dir, err := ioutil.TempDir("", "database.bolt")
		require.NoError(t, err)
		db, err := bolt.CreateDB(dir, "my.db", nil)
		require.NoError(t, err)
		return bolt.New(db).WithRealm(realm)

	case dbMapDB:
		return mapdb.NewMapDB().WithRealm(realm)

	case dbPebble:
		dir, err := ioutil.TempDir("", "database.pebble")
		require.NoError(t, err)
		db, err := pebble.CreateDB(dir)
		require.NoError(t, err)
		return pebble.New(db).WithRealm(realm)
	}

	panic("unknown database")
}

func TestSetAndGet(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)

		key := []byte("testKey")
		expectedValue := []byte("testValue")

		err := store.Set(key, expectedValue)
		require.NoError(t, err)

		value, err := store.Get(key)
		require.NoError(t, err)

		require.True(t, bytes.Equal(expectedValue, value))
	}
}

func TestSetAndGetEmptyValue(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)

		key := []byte("testKey")
		expectedValue := []byte{}

		err := store.Set(key, expectedValue)

		require.NoError(t, err)

		value, err := store.Get(key)
		require.NoError(t, err)

		require.True(t, bytes.Equal(expectedValue, value))
	}
}

func TestDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)

		key := []byte("testKey")
		expectedValue := []byte("testValue")

		err := store.Set(key, expectedValue)
		require.NoError(t, err)

		value, err := store.Get(key)
		require.NoError(t, err)

		require.True(t, bytes.Equal(expectedValue, value))

		err = store.Delete(key)
		require.NoError(t, err)

		_, err = store.Get(key)
		require.Error(t, kvstore.ErrKeyNotFound)
	}
}

func TestIterate(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
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

		err := store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			expectedValue, found := insertedValues[string(key)]
			require.True(t, found)
			require.Equal(t, expectedValue, string(value))
			delete(insertedValues, string(key))
			return true
		})
		require.NoError(t, err)

		require.Equal(t, 0, len(insertedValues))
	}
}

func TestIteratePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
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

		err := store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
			expectedValue, found := insertedValues[string(key)]
			require.True(t, found)
			require.Equal(t, expectedValue, string(value))
			delete(insertedValues, string(key))
			return true
		})

		require.NoError(t, err)

		require.Equal(t, 0, len(insertedValues))
	}
}

func TestIteratePrefixKeyOnly(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
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

		err := store.IterateKeys(kvstore.KeyPrefix("testKey"), func(key kvstore.Key) bool {
			_, found := insertedValues[string(key)]
			require.True(t, found)
			delete(insertedValues, string(key))
			return true
		})
		require.NoError(t, err)

		require.Equal(t, 0, len(insertedValues))
	}
}

func TestDeletePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
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
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {

			expectedValue, found := insertedValues[string(key)]
			require.True(t, found)
			require.Equal(t, expectedValue, string(value))
			delete(insertedValues, string(key))
			return true
		})
		require.NoError(t, err)

		require.Equal(t, 0, len(insertedValues))
	}
}

func TestDeletePrefixIsEmpty(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
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
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			t.Fail()
			return true
		})
		require.NoError(t, err)
	}
}

func TestSetAndOverwrite(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
		count := 100

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			err := store.Set([]byte(testKey), []byte{0})
			require.NoError(t, err)
		}

		verifyCount := 0
		// Verify that all entries are 0
		err := store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
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
			err := batch.Set([]byte(testKey), []byte{1})
			require.NoError(t, err)
		}

		err = batch.Commit()
		require.NoError(t, err)

		verifyCount = 0
		// Verify, that all entries were changed
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			require.True(t, bytes.Equal([]byte{1}, value))
			verifyCount++
			return true
		})
		require.NoError(t, err)

		// Check that we checked the correct amount of entries
		require.Equal(t, count, verifyCount)
	}
}

func TestBatchedWithSetAndDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)

		err := store.Set([]byte("testKey1"), []byte{42})
		require.NoError(t, err)

		err = store.Set([]byte("testKey2"), []byte{13})
		require.NoError(t, err)

		batch := store.Batched()

		err = batch.Set([]byte("testKey1"), []byte{84})
		require.NoError(t, err)

		err = batch.Set([]byte("testKey3"), []byte{69})
		require.NoError(t, err)

		err = batch.Delete([]byte("testKey2"))
		require.NoError(t, err)

		err = batch.Commit()
		require.NoError(t, err)

		err = store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
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
}

func TestBatchedWithDuplicateKeys(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)

		batch := store.Batched()

		err := batch.Set([]byte("testKey1"), []byte{84})
		require.NoError(t, err)

		err = batch.Set([]byte("testKey1"), []byte{69})
		require.NoError(t, err)

		err = batch.Commit()
		require.NoError(t, err)

		err = store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
			if string(key) == "testKey1" {
				require.True(t, bytes.Equal(value, []byte{69}))
			} else {
				t.Fail()
			}
			return true
		})
		require.NoError(t, err)
	}
}

func TestBatchedWithLotsOfKeys(t *testing.T) {

	prefix := []byte("testPrefix")
	for dbImplementation := range dbImplementations {
		store := testStore(t, dbImplementation, prefix)
		count := 500_000

		batch := store.Batched()

		for i := 0; i < count; i++ {
			testKey := make([]byte, 49)
			testValue := make([]byte, 156)
			rand.Read(testKey)
			rand.Read(testValue)
			err := batch.Set(testKey, testValue)
			require.NoError(t, err)
		}

		err := batch.Commit()
		require.NoError(t, err)

		verifyCount := 0
		// Verify, that all entries were changed
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			verifyCount++
			return true
		})
		require.NoError(t, err)

		// Check that we checked the correct amount of entries
		require.Equal(t, count, verifyCount)
	}
}
