package bucketstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

var (
	testKeyPrefix = kvstore.KeyPrefix("testKey")
)

var testEntries1 = []*struct {
	kvstore.Key
	kvstore.Value
}{
	{Key: []byte(fmt.Sprintf("%s_a", testKeyPrefix)), Value: []byte("testValue_A_1")},
	{Key: []byte(fmt.Sprintf("%s_b", testKeyPrefix)), Value: []byte("testValue_B_1")},
	{Key: []byte(fmt.Sprintf("%s_c", testKeyPrefix)), Value: []byte("testValue_C_1")},
	{Key: []byte(fmt.Sprintf("%s_d", testKeyPrefix)), Value: []byte("testValue_D_1")},
}

var testEntries2 = []*struct {
	kvstore.Key
	kvstore.Value
}{
	{Key: []byte(fmt.Sprintf("%s_a", testKeyPrefix)), Value: []byte("testValue_A_2")},
	{Key: []byte(fmt.Sprintf("%s_b", testKeyPrefix)), Value: []byte("testValue_B_2")},
	{Key: []byte(fmt.Sprintf("%s_c", testKeyPrefix)), Value: []byte("testValue_C_2")},
	{Key: []byte(fmt.Sprintf("%s_d", testKeyPrefix)), Value: []byte("testValue_D_2")},
}

func getTestStore() kvstore.KVStore {
	return mapdb.NewMapDB()
}

func getComparableTestStore(index int) kvstore.KVStore {
	store := getTestStore()

	// needed for interface comparison (without that it would be equal)
	store.Set([]byte("index"), []byte{byte(index)})

	return store
}

func TestBucketStoreNew(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1})
	require.Nil(t, bucketStore)
	require.ErrorIs(t, ErrNotEnoughBuckets, err)

	bucketStore, err = New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()
}

func TestBucketStoreAdd(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()
	testBucket4 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	require.NoError(t, bucketStore.BucketAdd(testBucket4))
}

func TestBucketStoreDelete(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	require.NoError(t, bucketStore.BucketDeleteLast()) // 3

	require.ErrorIs(t, ErrNotEnoughBuckets, bucketStore.BucketDeleteLast()) // 2
}

func TestBucketStoreClose(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	bucketStore.Close()

	err = bucketStore.BucketAdd(testBucket1)
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	err = bucketStore.BucketDeleteLast() // 3
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	_, err = bucketStore.BucketFirst()
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	_, err = bucketStore.BucketLast()
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	_, err = bucketStore.Bucket(0)
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	_, err = bucketStore.BucketCount()
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)

	_, err = bucketStore.BucketForEach(func(index int, bucket kvstore.KVStore) bool { return true })
	require.ErrorIs(t, err, kvstore.ErrStoreClosed)
}

func TestBucketStoreCount(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()
	testBucket4 := getTestStore()
	testBucket5 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	count, err := bucketStore.BucketCount()
	require.NoError(t, err)
	require.Equal(t, 3, count)

	require.NoError(t, bucketStore.BucketAdd(testBucket4))
	count, err = bucketStore.BucketCount()
	require.NoError(t, err)
	require.Equal(t, 4, count)

	require.NoError(t, bucketStore.BucketAdd(testBucket5))
	count, err = bucketStore.BucketCount()
	require.NoError(t, err)
	require.Equal(t, 5, count)

	require.NoError(t, bucketStore.BucketDeleteLast()) // 3
	count, err = bucketStore.BucketCount()
	require.NoError(t, err)
	require.Equal(t, 4, count)
}

func TestBucketStoreIndexing(t *testing.T) {
	testBucket1 := getComparableTestStore(1)
	testBucket2 := getComparableTestStore(2)
	testBucket3 := getComparableTestStore(3)
	testBucket4 := getComparableTestStore(4)
	testBucket5 := getComparableTestStore(5)

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	// 1 is the latest
	bucketFirst, err := bucketStore.BucketFirst()
	require.NotNil(t, bucketFirst)
	require.NoError(t, err)
	require.Equal(t, testBucket1, bucketFirst)

	// 3 is the oldest
	bucketLast, err := bucketStore.BucketLast()
	require.NotNil(t, bucketLast)
	require.NoError(t, err)
	require.Equal(t, testBucket3, bucketLast)

	require.NoError(t, bucketStore.BucketAdd(testBucket4))

	// 4 is the latest now
	bucketFirst, err = bucketStore.BucketFirst()
	require.NotNil(t, bucketFirst)
	require.NoError(t, err)
	require.Equal(t, testBucket4, bucketFirst)

	require.NoError(t, bucketStore.BucketAdd(testBucket5))

	// 5 is the latest now
	bucketFirst, err = bucketStore.BucketFirst()
	require.NotNil(t, bucketFirst)
	require.NoError(t, err)
	require.Equal(t, testBucket5, bucketFirst)

	bucket2, err := bucketStore.Bucket(3)
	require.NotNil(t, bucket2)
	require.NoError(t, err)
	require.Equal(t, testBucket2, bucket2)
	require.NotEqual(t, testBucket3, bucket2)
}

func TestBucketStoreForEach(t *testing.T) {
	testBucket1 := getComparableTestStore(1)
	testBucket2 := getComparableTestStore(2)
	testBucket3 := getComparableTestStore(3)
	testBucket4 := getComparableTestStore(4)
	testBucket5 := getComparableTestStore(5)

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2, testBucket3})
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	require.NoError(t, bucketStore.BucketAdd(testBucket4))
	require.NoError(t, bucketStore.BucketAdd(testBucket5))

	buckets := make(map[int]kvstore.KVStore)
	buckets[0] = testBucket5
	buckets[1] = testBucket4
	buckets[2] = testBucket1
	buckets[3] = testBucket2
	buckets[4] = testBucket3

	_, err = bucketStore.BucketForEach(func(index int, bucket kvstore.KVStore) bool {
		require.Equal(t, buckets[index], bucket)
		delete(buckets, index)
		return true
	})
	require.NoError(t, err)

	require.Equal(t, 0, len(buckets))
}

// KVStore interface
// normal kvstore usage is already tested in kvstore_test.go
// here we have to test the bucket use cases (add/delete bucket, old version of keys in older buckets, pruning etc...)

func TestBucketStoreGet(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()
	testBucket4 := getTestStore()
	testBucket5 := getTestStore()
	testBucket6 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(false))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	// add the entries to the first two buckets
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add buckets
	require.NoError(t, bucketStore.BucketAdd(testBucket3))
	require.NoError(t, bucketStore.BucketAdd(testBucket4))

	// set new values for existing keys
	for _, testEntry := range testEntries2 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if values match the new version
	for _, testEntry := range testEntries2 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	// add buckets and delete the old ones
	require.NoError(t, bucketStore.BucketDeleteLast()) // 2
	require.NoError(t, bucketStore.BucketDeleteLast()) // 1
	require.NoError(t, bucketStore.BucketAdd(testBucket5))
	require.NoError(t, bucketStore.BucketAdd(testBucket6))
	require.NoError(t, bucketStore.BucketDeleteLast()) // 3
	require.NoError(t, bucketStore.BucketDeleteLast()) // 4

	// check if keys are removed
	for _, testEntry := range testEntries1 {
		value, err := bucketStore.Get(testEntry.Key)
		require.Equal(t, kvstore.ErrKeyNotFound, err)
		require.Nil(t, value)
	}
}

func TestBucketStoreDeleteKey(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(false))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	// add the entries to the first two buckets
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add bucket
	require.NoError(t, bucketStore.BucketAdd(testBucket3))

	// set new values for existing keys
	for _, testEntry := range testEntries2 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if values match the new version
	for _, testEntry := range testEntries2 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	// delete keys
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Delete(testEntry.Key))
	}

	// check if keys are deleted in all buckets
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.Equal(t, kvstore.ErrKeyNotFound, err2)
		require.Nil(t, value)
	}
}

func TestBucketStoreDeletePrefix(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(false))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	// add the entries to the first two buckets
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add bucket
	require.NoError(t, bucketStore.BucketAdd(testBucket3))

	// set new values for existing keys
	for _, testEntry := range testEntries2 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
	}

	// check if values match the new version
	for _, testEntry := range testEntries2 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	// delete prefix
	require.NoError(t, bucketStore.DeletePrefix(testKeyPrefix))

	// check if keys are deleted in all buckets
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.Equal(t, kvstore.ErrKeyNotFound, err2)
		require.Nil(t, value)
	}
}

func TestBucketStoreIterateWithOldVersions(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()
	testBucket4 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(false))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	insertedValues := make(map[string]kvstore.Value)

	// add the entries to the first two buckets
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add buckets
	require.NoError(t, bucketStore.BucketAdd(testBucket3))
	require.NoError(t, bucketStore.BucketAdd(testBucket4))

	// set new values for existing keys
	for _, testEntry := range testEntries2 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	unknownValues := make(map[string]kvstore.Value)

	// there will be old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]
		if !found {
			unknownValues[string(key)] = value
			return true
		}

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))

	// the unknownValues should match the old versions of the keys
	for _, testEntry := range testEntries1 {
		value, found := unknownValues[string(testEntry.Key)]
		require.True(t, found)

		// check if values match the old version
		require.Equal(t, testEntry.Value, value)
		delete(unknownValues, string(testEntry.Key))
	}

	// check if we found all old values
	require.Equal(t, 0, len(unknownValues))

	// delete the old buckets to remove the old key versions
	require.NoError(t, bucketStore.BucketDeleteLast()) // 2
	require.NoError(t, bucketStore.BucketDeleteLast()) // 1

	// reinit the expected results
	insertedValues = make(map[string]kvstore.Value)
	for _, testEntry := range testEntries2 {
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	// there should be no old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]

		// check if no old versions of keys exist
		require.True(t, found)

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))
}

func TestBucketStoreIterateWithoutOldVersions(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(true))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	insertedValues := make(map[string]kvstore.Value)

	// add the entries to the first two buckets
	for _, testEntry := range testEntries1 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add bucket
	require.NoError(t, bucketStore.BucketAdd(testBucket3))

	// set new values for existing keys
	for _, testEntry := range testEntries2 {
		require.NoError(t, bucketStore.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	// there should be no old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]

		// check if no old versions of keys exist
		require.True(t, found)

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))
}

func TestBucketStoreBatchedWithOldVersions(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()
	testBucket4 := getTestStore()
	testBucket5 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(false))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	insertedValues := make(map[string]kvstore.Value)

	// add the entries to the first two buckets
	batchedMutation, err := bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries1 {
		require.NoError(t, batchedMutation.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}
	require.NoError(t, batchedMutation.Commit())

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add buckets
	require.NoError(t, bucketStore.BucketAdd(testBucket3))
	require.NoError(t, bucketStore.BucketAdd(testBucket4))

	// set new values for existing keys
	batchedMutation, err = bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries2 {
		require.NoError(t, batchedMutation.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}
	require.NoError(t, batchedMutation.Commit())

	unknownValues := make(map[string]kvstore.Value)

	// there will be old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]
		if !found {
			unknownValues[string(key)] = value
			return true
		}

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))

	// the unknownValues should match the old versions of the keys
	for _, testEntry := range testEntries1 {
		value, found := unknownValues[string(testEntry.Key)]
		require.True(t, found)

		// check if values match the old version
		require.Equal(t, testEntry.Value, value)
		delete(unknownValues, string(testEntry.Key))
	}

	// check if we found all old values
	require.Equal(t, 0, len(unknownValues))

	// delete the old buckets to remove the old key versions
	require.NoError(t, bucketStore.BucketDeleteLast()) // 2
	require.NoError(t, bucketStore.BucketDeleteLast()) // 1

	// reinit the expected results
	insertedValues = make(map[string]kvstore.Value)
	for _, testEntry := range testEntries2 {
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}

	// there should be no old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]

		// check if no old versions of keys exist
		require.True(t, found)

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))

	// add bucket
	require.NoError(t, bucketStore.BucketAdd(testBucket5))

	// delete the keys
	batchedMutation, err = bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries2 {
		require.NoError(t, batchedMutation.Delete(testEntry.Key))
	}
	require.NoError(t, batchedMutation.Commit())

	// there should be no keys
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		require.Fail(t, "no keys expected")
		return true
	}))
}

func TestBucketStoreBatchedWithoutOldVersions(t *testing.T) {
	testBucket1 := getTestStore()
	testBucket2 := getTestStore()
	testBucket3 := getTestStore()

	bucketStore, err := New([]kvstore.KVStore{testBucket1, testBucket2}, WithDeleteKeyInOtherBucketsOnSet(true))
	require.NotNil(t, bucketStore)
	require.NoError(t, err)

	defer bucketStore.Close()

	insertedValues := make(map[string]kvstore.Value)

	// add the entries to the first two buckets
	batchedMutation, err := bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries1 {
		require.NoError(t, batchedMutation.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}
	require.NoError(t, batchedMutation.Commit())

	// check if they exist
	for _, testEntry := range testEntries1 {
		value, err2 := bucketStore.Get(testEntry.Key)
		require.NoError(t, err2)
		require.Equal(t, testEntry.Value, value)
	}

	value, err := bucketStore.Get([]byte("invalid"))
	require.Equal(t, kvstore.ErrKeyNotFound, err)
	require.Nil(t, value)

	// add bucket
	require.NoError(t, bucketStore.BucketAdd(testBucket3))

	// set new values for existing keys
	batchedMutation, err = bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries2 {
		require.NoError(t, batchedMutation.Set(testEntry.Key, testEntry.Value))
		insertedValues[string(testEntry.Key)] = testEntry.Value
	}
	require.NoError(t, batchedMutation.Commit())

	// there should be no old versions of the same keys in old buckets
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		valueExpected, found := insertedValues[string(key)]

		// check if no old versions of keys exist
		require.True(t, found)

		// check if values match the latest version
		require.Equal(t, valueExpected, value)
		delete(insertedValues, string(key))
		return true
	}))

	// check if we found all inserted values
	require.Equal(t, 0, len(insertedValues))

	// delete the keys
	batchedMutation, err = bucketStore.Batched()
	require.NoError(t, err)
	for _, testEntry := range testEntries2 {
		require.NoError(t, batchedMutation.Delete(testEntry.Key))
	}
	require.NoError(t, batchedMutation.Commit())

	// there should be no keys
	require.NoError(t, bucketStore.Iterate(testKeyPrefix, func(key, value kvstore.Value) bool {
		require.Fail(t, "no keys expected")
		return true
	}))
}
