package test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/debug"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
)

var (
	testEntries = []*struct {
		kvstore.Key
		kvstore.Value
	}{
		{Key: []byte("a"), Value: []byte("valueA")},
		{Key: []byte("b"), Value: []byte("valueB")},
		{Key: []byte("c"), Value: []byte("valueC")},
		{Key: []byte("d"), Value: []byte("valueD")},
	}
)

func testStore(t *testing.T, dbImplementation string, realm []byte) (kvstore.KVStore, error) {
	switch dbImplementation {

	case "mapDB":
		return mapdb.NewMapDB().WithRealm(realm)

	case "rocksdb":
		dir := t.TempDir()
		db, err := rocksdb.CreateDB(dir)
		require.NoError(t, err, "used db: %s", dbImplementation)

		return rocksdb.New(db).WithRealm(realm)

	case "debug":
		return debug.New(mapdb.NewMapDB(), func(command debug.Command, parameters ...[]byte) {
			s := []string{
				debug.CommandNames[command],
			}
			for _, p := range parameters {
				s = append(s, string(p))
			}
			fmt.Println(s)
		}).WithRealm(realm)
	}

	panic("unknown database")
}

func countKeys(t *testing.T, store kvstore.KVStore) int {
	count := 0
	err := store.IterateKeys(kvstore.EmptyPrefix, func(k kvstore.Key) bool {
		count++

		return true
	})
	require.NoError(t, err)

	return count
}

func TestSetAndGet(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		for _, entry := range testEntries {
			err := store.Set(entry.Key, entry.Value)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		for _, entry := range testEntries {
			value, err := store.Get(entry.Key)
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.True(t, bytes.Equal(entry.Value, value), "used db: %s", dbImplementation)
		}

		value, err := store.Get([]byte("invalid"))
		require.Equal(t, kvstore.ErrKeyNotFound, err, "used db: %s", dbImplementation)
		require.Nil(t, value, "used db: %s", dbImplementation)
	}
}

func TestSetAndGetEmptyValue(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		expectedValue := []byte{}

		for _, entry := range testEntries {
			err := store.Set(entry.Key, expectedValue)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		for _, entry := range testEntries {
			value, err := store.Get(entry.Key)
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.True(t, bytes.Equal(expectedValue, value), "used db: %s", dbImplementation)
		}

		value, err := store.Get([]byte("invalid"))
		require.Equal(t, kvstore.ErrKeyNotFound, err, "used db: %s", dbImplementation)
		require.Nil(t, value, "used db: %s", dbImplementation)
	}
}

func TestDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		for _, entry := range testEntries {
			err := store.Set(entry.Key, entry.Value)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		for _, entry := range testEntries {
			value, err := store.Get(entry.Key)
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.True(t, bytes.Equal(entry.Value, value), "used db: %s", dbImplementation)
		}

		for _, entry := range testEntries {
			err := store.Delete(entry.Key)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		for _, entry := range testEntries {
			value, err := store.Get(entry.Key)
			require.Equal(t, kvstore.ErrKeyNotFound, err, "used db: %s", dbImplementation)
			require.Nil(t, value, "used db: %s", dbImplementation)
		}
	}
}

func TestRealm(t *testing.T) {

	prefix := kvstore.EmptyPrefix
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		realm := kvstore.Realm("realm")
		realmStore, err := store.WithRealm(realm)
		require.NoError(t, err)

		tmpStore, err := store.WithRealm(kvstore.Realm("tmp"))
		require.NoError(t, err)

		for _, entry := range testEntries {
			err := realmStore.Set(entry.Key, entry.Value)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}
		require.Equal(t, len(testEntries), countKeys(t, realmStore), "used db: %s", dbImplementation)

		for _, entry := range testEntries {
			err := tmpStore.Set(append(entry.Key, []byte("_2")...), entry.Value)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}
		require.Equal(t, len(testEntries), countKeys(t, tmpStore), "used db: %s", dbImplementation)

		realmStore2, err := store.WithRealm(realm)
		require.NoError(t, err)

		for _, entry := range testEntries {
			has, err := realmStore2.Has(entry.Key)
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.True(t, has, "used db: %s", dbImplementation)

			has, err = realmStore2.Has(append(entry.Key, []byte("_2")...))
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.False(t, has, "used db: %s", dbImplementation)
		}

		// when clearing "realm" the keys in "tmp" should still be there
		err = realmStore.Clear()
		require.NoError(t, err, "used db: %s", dbImplementation)

		require.Equal(t, 0, countKeys(t, realmStore), "used db: %s", dbImplementation)
		require.Equal(t, len(testEntries), countKeys(t, tmpStore), "used db: %s", dbImplementation)

		for _, entry := range testEntries {
			has, err := tmpStore.Has(append(entry.Key, []byte("_2")...))
			require.NoError(t, err, "used db: %s", dbImplementation)
			require.True(t, has, "used db: %s", dbImplementation)
		}
	}
}

func TestClear(t *testing.T) {

	prefix := kvstore.EmptyPrefix
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		require.EqualValues(t, 0, countKeys(t, store), "used db: %s", dbImplementation)

		for _, entry := range testEntries {
			err := store.Set(entry.Key, entry.Value)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}
		require.Equal(t, len(testEntries), countKeys(t, store), "used db: %s", dbImplementation)

		// check that Clear removes all the keys
		err = store.Clear()
		require.NoError(t, err, "used db: %s", dbImplementation)
		require.EqualValues(t, 0, countKeys(t, store), "used db: %s", dbImplementation)
	}
}

func TestIterate(t *testing.T) {

	prefix := kvstore.EmptyPrefix
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			expectedValue, found := insertedValues[string(key)]
			require.True(t, found, "used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "used db: %s", dbImplementation)
			delete(insertedValues, string(key))

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		require.Equal(t, 0, len(insertedValues), "used db: %s", dbImplementation)
	}
}

func TestIterateDirection(t *testing.T) {

	prefix := kvstore.EmptyPrefix

	for _, dbImplementation := range dbImplementations {

		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 9

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		insertedValuesWithTestPrefix := len(insertedValues)

		// forward iteration
		i := 0
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			str := strconv.FormatInt(int64(i), 10)
			expectedKey := "testKey" + str
			expectedValue := "testValue" + str

			require.Equal(t, expectedKey, string(key), "direction forward, used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "direction forward, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionForward)
		require.NoError(t, err, "direction forward, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction forward, used db: %s", dbImplementation)

		// backward iteration
		i = 0
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			str := strconv.FormatInt(int64(insertedValuesWithTestPrefix-1-i), 10)
			expectedKey := "testKey" + str
			expectedValue := "testValue" + str

			require.Equal(t, expectedKey, string(key), "direction backward, used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "direction backward, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionBackward)
		require.NoError(t, err, "direction backward, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction backward, used db: %s", dbImplementation)

		// insert other keys to check prefix filtering
		for i = 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "exampleKey" + str
			testValue := "exampleValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		// insert "upperBound" key for backwards prefix scan edge case
		testKey := "tesu"
		testValue := ""
		err = store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err, "used db: %s", dbImplementation)
		insertedValues[testKey] = testValue

		// forward iteration with prefix
		i = 0
		err = store.Iterate([]byte("test"), func(key kvstore.Key, value kvstore.Value) bool {
			str := strconv.FormatInt(int64(i), 10)
			expectedKey := "testKey" + str
			expectedValue := "testValue" + str

			require.Equal(t, expectedKey, string(key), "direction forward with prefix, used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "direction forward with prefix, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionForward)
		require.NoError(t, err, "direction forward with prefix, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction forward with prefix, used db: %s", dbImplementation)

		// backward iteration with prefix
		i = 0
		err = store.Iterate([]byte("test"), func(key kvstore.Key, value kvstore.Value) bool {
			str := strconv.FormatInt(int64(insertedValuesWithTestPrefix-1-i), 10)
			expectedKey := "testKey" + str
			expectedValue := "testValue" + str

			require.Equal(t, expectedKey, string(key), "direction backward with prefix, used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "direction backward with prefix, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionBackward)
		require.NoError(t, err, "direction backward with prefix, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction backward with prefix, used db: %s", dbImplementation)
	}
}

func TestIterateDirectionKeyOnly(t *testing.T) {

	prefix := kvstore.EmptyPrefix

	for _, dbImplementation := range dbImplementations {

		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 9

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		insertedValuesWithTestPrefix := len(insertedValues)

		// forward iteration
		i := 0
		err = store.IterateKeys(kvstore.EmptyPrefix, func(key kvstore.Key) bool {
			str := strconv.FormatInt(int64(i), 10)
			expectedKey := "testKey" + str

			require.Equal(t, expectedKey, string(key), "direction forward, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionForward)
		require.NoError(t, err, "direction forward, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction forward, used db: %s", dbImplementation)

		// backward iteration
		i = 0
		err = store.IterateKeys(kvstore.EmptyPrefix, func(key kvstore.Key) bool {
			str := strconv.FormatInt(int64(insertedValuesWithTestPrefix-1-i), 10)
			expectedKey := "testKey" + str

			require.Equal(t, expectedKey, string(key), "direction backward, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionBackward)
		require.NoError(t, err, "direction backward, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction backward, used db: %s", dbImplementation)

		// insert other keys to check prefix filtering
		for i = 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "exampleKey" + str
			testValue := "exampleValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		// insert "upperBound" key for backwards prefix scan edge case
		testKey := "tesu"
		testValue := ""
		err = store.Set([]byte(testKey), []byte(testValue))
		require.NoError(t, err, "used db: %s", dbImplementation)
		insertedValues[testKey] = testValue

		// forward iteration with prefix
		i = 0
		err = store.IterateKeys([]byte("test"), func(key kvstore.Key) bool {
			str := strconv.FormatInt(int64(i), 10)
			expectedKey := "testKey" + str

			require.Equal(t, expectedKey, string(key), "direction forward with prefix, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionForward)
		require.NoError(t, err, "direction forward with prefix, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction forward with prefix, used db: %s", dbImplementation)

		// backward iteration with prefix
		i = 0
		err = store.IterateKeys([]byte("test"), func(key kvstore.Key) bool {
			str := strconv.FormatInt(int64(insertedValuesWithTestPrefix-1-i), 10)
			expectedKey := "testKey" + str

			require.Equal(t, expectedKey, string(key), "direction backward with prefix, used db: %s", dbImplementation)

			i++

			return true
		}, kvstore.IterDirectionBackward)
		require.NoError(t, err, "direction backward with prefix, used db: %s", dbImplementation)
		require.Equal(t, insertedValuesWithTestPrefix, i, "direction backward with prefix, used db: %s", dbImplementation)
	}
}

func TestIteratePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		// Insert some more values with a different prefix
		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			err = store.Set([]byte("someOtherKey"+str), []byte(str))
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
			expectedValue, found := insertedValues[string(key)]
			require.True(t, found, "used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "used db: %s", dbImplementation)
			delete(insertedValues, string(key))

			return true
		})

		require.NoError(t, err, "used db: %s", dbImplementation)

		require.Equal(t, 0, len(insertedValues), "used db: %s", dbImplementation)
	}
}

func TestIteratePrefixKeyOnly(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		// Insert some more values with a different prefix
		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			err = store.Set([]byte("someOtherKey"+str), []byte(str))
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = store.IterateKeys(kvstore.KeyPrefix("testKey"), func(key kvstore.Key) bool {
			_, found := insertedValues[string(key)]
			require.True(t, found, "used db: %s", dbImplementation)
			delete(insertedValues, string(key))

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		require.Equal(t, 0, len(insertedValues), "used db: %s", dbImplementation)
	}
}

func TestDeletePrefix(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 1000

		insertedValues := make(map[string]string)

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
			insertedValues[testKey] = testValue
		}

		// Insert some more values with a different prefix
		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			err = store.Set([]byte("someOtherKey"+str), []byte(str))
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = store.DeletePrefix([]byte("someOtherKey"))
		require.NoError(t, err, "used db: %s", dbImplementation)

		// Verify, that the database only contains the elements without the delete prefix
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {

			expectedValue, found := insertedValues[string(key)]
			require.True(t, found, "used db: %s", dbImplementation)
			require.Equal(t, expectedValue, string(value), "used db: %s", dbImplementation)
			delete(insertedValues, string(key))

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		require.Equal(t, 0, len(insertedValues), "used db: %s", dbImplementation)
	}
}

func TestDeletePrefixIsEmpty(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			testValue := "testValue" + str
			err = store.Set([]byte(testKey), []byte(testValue))
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = store.DeletePrefix(kvstore.EmptyPrefix)
		require.NoError(t, err, "used db: %s", dbImplementation)

		// Verify, that the database does not contain any items since we deleted using the prefix
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			t.Fail()

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)
	}
}

func TestSetAndOverwrite(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100

		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			err = store.Set([]byte(testKey), []byte{0})
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		verifyCount := 0
		// Verify that all entries are 0
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			require.True(t, bytes.Equal([]byte{0}, value), "used db: %s", dbImplementation)
			verifyCount++

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		// Check that we checked the correct amount of entries
		require.Equal(t, count, verifyCount, "used db: %s", dbImplementation)

		batch, err := store.Batched()
		require.NoError(t, err)

		// Batch edit all to value 1
		for i := 0; i < count; i++ {
			str := strconv.FormatInt(int64(i), 10)
			testKey := "testKey" + str
			err = batch.Set([]byte(testKey), []byte{1})
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = batch.Commit()
		require.NoError(t, err, "used db: %s", dbImplementation)

		verifyCount = 0
		// Verify, that all entries were changed
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			require.True(t, bytes.Equal([]byte{1}, value), "used db: %s", dbImplementation)
			verifyCount++

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		// Check that we checked the correct amount of entries
		require.Equal(t, count, verifyCount, "used db: %s", dbImplementation)
	}
}

func TestBatchedWithSetAndDelete(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		err = store.Set([]byte("testKey1"), []byte{42})
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = store.Set([]byte("testKey2"), []byte{13})
		require.NoError(t, err, "used db: %s", dbImplementation)

		batch, err := store.Batched()
		require.NoError(t, err)

		err = batch.Set([]byte("testKey1"), []byte{84})
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = batch.Set([]byte("testKey3"), []byte{69})
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = batch.Delete([]byte("testKey2"))
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = batch.Commit()
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
			switch {
			case string(key) == "testKey1":
				require.True(t, bytes.Equal(value, []byte{84}), "used db: %s", dbImplementation)

			case string(key) == "testKey3":
				require.True(t, bytes.Equal(value, []byte{69}), "used db: %s", dbImplementation)

			default:
				t.Fail()
			}

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)
	}
}

func TestBatchedWithDuplicateKeys(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		batch, err := store.Batched()
		require.NoError(t, err)

		err = batch.Set([]byte("testKey1"), []byte{84})
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = batch.Set([]byte("testKey1"), []byte{69})
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = batch.Commit()
		require.NoError(t, err, "used db: %s", dbImplementation)

		err = store.Iterate(kvstore.KeyPrefix("testKey"), func(key kvstore.Key, value kvstore.Value) bool {
			if string(key) == "testKey1" {
				require.True(t, bytes.Equal(value, []byte{69}), "used db: %s", dbImplementation)
			} else {
				t.Fail()
			}

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)
	}
}

func TestBatchedWithLotsOfKeys(t *testing.T) {
	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err)

		count := 100_000

		batch, err := store.Batched()
		require.NoError(t, err)

		for i := 0; i < count; i++ {
			testKey := make([]byte, 49)
			testValue := make([]byte, 156)
			_, err = rand.Read(testKey)
			require.NoError(t, err)
			_, err = rand.Read(testValue)
			require.NoError(t, err)
			err = batch.Set(testKey, testValue)
			require.NoError(t, err, "used db: %s", dbImplementation)
		}

		err = batch.Commit()
		require.NoError(t, err, "used db: %s", dbImplementation)

		verifyCount := 0
		// Verify, that all entries were changed
		err = store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
			verifyCount++

			return true
		})
		require.NoError(t, err, "used db: %s", dbImplementation)

		// Check that we checked the correct amount of entries
		require.Equal(t, count, verifyCount, "used db: %s", dbImplementation)

		require.NoError(t, store.Close())
	}
}

func TestStoreClosed(t *testing.T) {

	prefix := []byte("testPrefix")
	for _, dbImplementation := range dbImplementations {
		store, err := testStore(t, dbImplementation, prefix)
		require.NoError(t, err, "used db: %s", dbImplementation)

		batchedMutation, err := store.Batched()
		require.NoError(t, err, "used db: %s", dbImplementation)

		require.NoError(t, store.Close(), "used db: %s", dbImplementation)

		err = batchedMutation.Commit()
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		_, err = store.WithRealm(kvstore.EmptyPrefix)
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Iterate(kvstore.EmptyPrefix, func(key, value kvstore.Value) bool { return true })
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.IterateKeys(kvstore.EmptyPrefix, func(key kvstore.Key) bool { return true })
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Clear()
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		_, err = store.Get(kvstore.Key{0})
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Set(kvstore.Key{0}, []byte{1})
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		_, err = store.Has(kvstore.Key{0})
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Delete(kvstore.Key{0})
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.DeletePrefix(kvstore.EmptyPrefix)
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Flush()
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)

		err = store.Close()
		require.ErrorIs(t, err, nil, "used db: %s", dbImplementation)

		_, err = store.Batched()
		require.ErrorIs(t, err, kvstore.ErrStoreClosed, "used db: %s", dbImplementation)
	}
}
