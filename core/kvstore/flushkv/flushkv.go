package flushkv

import (
	"github.com/iotaledger/hive.go/core/byteutils"
	"github.com/iotaledger/hive.go/core/kvstore"
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

func (s *flushKVStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	store, err := s.store.WithRealm(realm)
	if err != nil {
		return nil, err
	}

	return &flushKVStore{
		store: store,
	}, nil
}

func (s *flushKVStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

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

func (s *flushKVStore) Clear() error {
	if err := s.store.Clear(); err != nil {
		return err
	}

	return s.store.Flush()
}

func (s *flushKVStore) Get(key kvstore.Key) (kvstore.Value, error) {
	return s.store.Get(key)
}

func (s *flushKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	if err := s.store.Set(key, value); err != nil {
		return err
	}

	return s.store.Flush()
}

func (s *flushKVStore) Has(key kvstore.Key) (bool, error) {
	return s.store.Has(key)
}

func (s *flushKVStore) Delete(key kvstore.Key) error {
	if err := s.store.Delete(key); err != nil {
		return err
	}

	return s.store.Flush()
}

func (s *flushKVStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if err := s.store.DeletePrefix(prefix); err != nil {
		return err
	}

	return s.store.Flush()
}

func (s *flushKVStore) Flush() error {
	return s.store.Flush()
}

func (s *flushKVStore) Close() error {
	return s.store.Close()
}

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

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	return b.batched.Set(key, value)
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	return b.batched.Delete(key)
}

func (b *batchedMutations) Cancel() {
	b.batched.Cancel()
}

func (b *batchedMutations) Commit() error {
	if err := b.batched.Commit(); err != nil {
		return err
	}

	return b.store.Flush()
}

var _ kvstore.KVStore = &flushKVStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}
