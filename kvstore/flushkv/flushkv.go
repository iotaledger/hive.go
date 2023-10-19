package flushkv

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

// flushKVStore is a wrapper to any KVStore that flushes changes immediately.
type flushKVStore struct {
	store kvstore.KVStore
}

// New creates a kvstore.KVStore implementation that flushes changes immediately.
func New(store kvstore.KVStore) kvstore.KVStore {
	return &flushKVStore{
		store: store,
	}
}

// WithRealm is a factory method for using the same underlying storage with a different realm.
func (s *flushKVStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	store, err := s.store.WithRealm(realm)
	if err != nil {
		return nil, err
	}

	return &flushKVStore{
		store: store,
	}, nil
}

// WithExtendedRealm is a factory method for using the same underlying storage with an realm appended to existing one.
func (s *flushKVStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

// Realm returns the configured realm.
func (s *flushKVStore) Realm() kvstore.Realm {
	return s.store.Realm()
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *flushKVStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	return s.store.Iterate(prefix, consumerFunc, iterDirection...)
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *flushKVStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	return s.store.IterateKeys(prefix, consumerFunc, iterDirection...)
}

// Clear clears the realm.
func (s *flushKVStore) Clear() error {
	if err := s.store.Clear(); err != nil {
		return err
	}

	return s.store.Flush()
}

// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
func (s *flushKVStore) Get(key kvstore.Key) (kvstore.Value, error) {
	return s.store.Get(key)
}

// Set sets the given key and value.
func (s *flushKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	if err := s.store.Set(key, value); err != nil {
		return err
	}

	return s.store.Flush()
}

// Has checks whether the given key exists.
func (s *flushKVStore) Has(key kvstore.Key) (bool, error) {
	return s.store.Has(key)
}

// Delete deletes the entry for the given key.
func (s *flushKVStore) Delete(key kvstore.Key) error {
	if err := s.store.Delete(key); err != nil {
		return err
	}

	return s.store.Flush()
}

// DeletePrefix deletes all the entries matching the given key prefix.
func (s *flushKVStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if err := s.store.DeletePrefix(prefix); err != nil {
		return err
	}

	return s.store.Flush()
}

// Flush persists all outstanding write operations to disc.
func (s *flushKVStore) Flush() error {
	return s.store.Flush()
}

// Close closes the database file handles.
func (s *flushKVStore) Close() error {
	return s.store.Close()
}

// Batched returns a BatchedMutations interface to execute batched mutations.
func (s *flushKVStore) Batched() (kvstore.BatchedMutations, error) {
	batched, err := s.store.Batched()
	if err != nil {
		return nil, err
	}

	return &batchedMutations{
		store:   s.store,
		batched: batched,
	}, nil
}

// batchedMutations is a wrapper around a WriteBatch of a flushKVStore.
type batchedMutations struct {
	store   kvstore.KVStore
	batched kvstore.BatchedMutations
}

// Set sets the given key and value.
func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	return b.batched.Set(key, value)
}

// Delete deletes the entry for the given key.
func (b *batchedMutations) Delete(key kvstore.Key) error {
	return b.batched.Delete(key)
}

// Cancel cancels the batched mutations.
func (b *batchedMutations) Cancel() {
	b.batched.Cancel()
}

// Commit commits/flushes the mutations.
func (b *batchedMutations) Commit() error {
	if err := b.batched.Commit(); err != nil {
		return err
	}

	return b.store.Flush()
}

// code guards.
var _ kvstore.KVStore = &flushKVStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}
