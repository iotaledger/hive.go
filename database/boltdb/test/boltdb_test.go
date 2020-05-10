package test

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/hive.go/database"
	"github.com/iotaledger/hive.go/database/boltdb"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"io/ioutil"
	"strconv"
	"testing"
)

func testDatabase(t require.TestingT, prefix []byte) *boltdb.BoltDB {
	dir, err := ioutil.TempDir("", "database.boltdb")
	require.NoError(t, err)
	dirAndFile := fmt.Sprintf("%s/my.db", dir)
	db, err := bbolt.Open(dirAndFile, 0666, nil)
	require.NoError(t, err)
	return boltdb.NewDBWithPrefix(prefix, db)
}

func TestSetAndGet(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)

	key := []byte("testKey")
	value := []byte("testValue")

	err := db.Set(database.Entry{
		Key:   key,
		Value: value,
	})
	require.NoError(t, err)

	entry, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(key, entry.Key) && bytes.Equal(value, entry.Value))
}

func TestSetAndGetEmptyValue(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)

	key := []byte("testKey")
	value := []byte{}

	err := db.Set(database.Entry{
		Key:   key,
		Value: value,
	})
	require.NoError(t, err)

	entry, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(key, entry.Key) && bytes.Equal(value, entry.Value))
}

func TestDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)

	key := []byte("testKey")
	value := []byte("testValue")

	err := db.Set(database.Entry{
		Key:   key,
		Value: value,
	})
	require.NoError(t, err)

	entry, err := db.Get(key)
	require.NoError(t, err)

	require.True(t, bytes.Equal(key, entry.Key) && bytes.Equal(value, entry.Value))

	err = db.Delete(key)
	require.NoError(t, err)

	_, err = db.Get(key)
	require.Error(t, boltdb.ErrKeyNotFound)
}

func TestForEach(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)
	count := 1000

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := db.Set(database.Entry{
			Key:   []byte(testKey),
			Value: []byte(testValue),
		})
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	db.ForEach(func(entry database.Entry) (stop bool) {

		value, found := insertedValues[string(entry.Key)]
		require.True(t, found)
		require.Equal(t, value, string(entry.Value))
		delete(insertedValues, string(entry.Key))
		return false
	})

	require.Equal(t, 0, len(insertedValues))
}

func TestForEachPrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)
	count := 1000

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := db.Set(database.Entry{
			Key:   []byte(testKey),
			Value: []byte(testValue),
		})
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := db.Set(database.Entry{
			Key:   []byte("someOtherKey" + str),
			Value: []byte(str),
		})
		require.NoError(t, err)
	}

	db.ForEachPrefix([]byte("testKey"), func(entry database.Entry) (stop bool) {

		value, found := insertedValues[string(entry.Key)]
		require.True(t, found)
		require.Equal(t, value, string(entry.Value))
		delete(insertedValues, string(entry.Key))
		return false
	})

	require.Equal(t, 0, len(insertedValues))
}

func TestForEachPrefixKeyOnly(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)
	count := 1000

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := db.Set(database.Entry{
			Key:   []byte(testKey),
			Value: []byte(testValue),
		})
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := db.Set(database.Entry{
			Key:   []byte("someOtherKey" + str),
			Value: []byte(str),
		})
		require.NoError(t, err)
	}

	db.ForEachPrefixKeyOnly([]byte("testKey"), func(key database.Key) (stop bool) {

		_, found := insertedValues[string(key)]
		require.True(t, found)
		delete(insertedValues, string(key))
		return false
	})

	require.Equal(t, 0, len(insertedValues))
}

func TestDeletePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	db := testDatabase(t, prefix)
	count := 1000

	insertedValues := make(map[string]string)

	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		testKey := "testKey" + str
		testValue := "testValue" + str
		err := db.Set(database.Entry{
			Key:   []byte(testKey),
			Value: []byte(testValue),
		})
		require.NoError(t, err)
		insertedValues[testKey] = testValue
	}

	// Insert some more values with a different prefix
	for i := 0; i < count; i++ {
		str := strconv.FormatInt(int64(i), 10)
		err := db.Set(database.Entry{
			Key:   []byte("someOtherKey" + str),
			Value: []byte(str),
		})
		require.NoError(t, err)
	}

	err := db.DeletePrefix([]byte("someOtherKey"))
	require.NoError(t, err)

	// Verify, that the database only contains the elements without the delete prefix
	db.ForEach(func(entry database.Entry) (stop bool) {

		value, found := insertedValues[string(entry.Key)]
		require.True(t, found)
		require.Equal(t, value, string(entry.Value))
		delete(insertedValues, string(entry.Key))
		return false
	})

	require.Equal(t, 0, len(insertedValues))
}
