package bucketstore

import (
	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
)

// Realm is a KVStore implementation
// that uses an underlying BucketStore with a prefix.
type Realm struct {
	bucketStore *BucketStore
	realm       kvstore.Realm
}

// NewRealm creates a new Realm.
func NewRealm(bucketStore *BucketStore, realm kvstore.Realm) *Realm {
	return &Realm{
		bucketStore: bucketStore,
		realm:       realm,
	}
}

// buildKeyPrefix builds a key using the realm and the given prefix.
func (bsr *Realm) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(bsr.realm, prefix)
}

///////////////////////
// KVStore interface //
///////////////////////

// WithRealm is a factory method for using the same underlying storage with a different realm.
func (bsr *Realm) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	if bsr.bucketStore.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return NewRealm(bsr.bucketStore, bsr.buildKeyPrefix(realm)), nil
}

// Realm returns the configured realm.
func (bsr *Realm) Realm() kvstore.Realm {
	return bsr.realm
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (bsr *Realm) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {

	kvConsumerFuncWrapped := func(key kvstore.Key, value kvstore.Value) bool {
		return kvConsumerFunc(key[len(bsr.realm):], value)
	}

	return bsr.bucketStore.Iterate(bsr.buildKeyPrefix(prefix), kvConsumerFuncWrapped, direction...)
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (bsr *Realm) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {

	consumerFuncWrapped := func(key kvstore.Key) bool {
		return consumerFunc(key[len(bsr.realm):])
	}

	return bsr.bucketStore.IterateKeys(bsr.buildKeyPrefix(prefix), consumerFuncWrapped, direction...)
}

// Clear clears the realm.
func (bsr *Realm) Clear() error {
	return bsr.bucketStore.DeletePrefix(bsr.realm)
}

// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
func (bsr *Realm) Get(key kvstore.Key) (kvstore.Value, error) {
	return bsr.bucketStore.Get(bsr.buildKeyPrefix(key))
}

// Set sets the given key and value.
func (bsr *Realm) Set(key kvstore.Key, value kvstore.Value) error {
	return bsr.bucketStore.Set(bsr.buildKeyPrefix(key), value)
}

// Has checks whether the given key exists.
func (bsr *Realm) Has(key kvstore.Key) (bool, error) {
	return bsr.bucketStore.Has(bsr.buildKeyPrefix(key))
}

// Delete deletes the entry for the given key.
func (bsr *Realm) Delete(key kvstore.Key) error {
	return bsr.bucketStore.Delete(bsr.buildKeyPrefix(key))
}

// DeletePrefix deletes all the entries matching the given key prefix.
func (bsr *Realm) DeletePrefix(prefix kvstore.KeyPrefix) error {
	return bsr.bucketStore.DeletePrefix(bsr.buildKeyPrefix(prefix))
}

// Flush persists all outstanding write operations to disc.
func (bsr *Realm) Flush() error {
	return bsr.bucketStore.Flush()
}

// Close closes the database file handles.
func (bsr *Realm) Close() error {
	return bsr.bucketStore.Close()
}

// Batched returns a BatchedMutations interface to execute batched mutations.
func (bsr *Realm) Batched() (kvstore.BatchedMutations, error) {
	batchedMutation, err := bsr.bucketStore.Batched()
	if err != nil {
		return nil, err
	}

	return &batchedMutationsWithRealm{
		batched: batchedMutation,
		realm:   bsr.realm,
	}, nil
}

// batchedMutationsWithRealm is a wrapper around a kvstore.BatchedMutations with realm.
type batchedMutationsWithRealm struct {
	batched kvstore.BatchedMutations
	realm   kvstore.Realm
}

// Set sets the given key and value.
func (b *batchedMutationsWithRealm) Set(key kvstore.Key, value kvstore.Value) error {
	return b.batched.Set(byteutils.ConcatBytes(b.realm, key), value)
}

// Delete deletes the entry for the given key.
func (b *batchedMutationsWithRealm) Delete(key kvstore.Key) error {
	return b.batched.Delete(byteutils.ConcatBytes(b.realm, key))
}

// Cancel cancels the batched mutations.
func (b *batchedMutationsWithRealm) Cancel() {
	b.batched.Cancel()
}

// Commit commits/flushes the mutations.
func (b *batchedMutationsWithRealm) Commit() error {
	return b.batched.Commit()
}

// code guards
var _ kvstore.KVStore = &Realm{}
var _ kvstore.BatchedMutations = &batchedMutationsWithRealm{}
