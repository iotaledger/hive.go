// Package mapdb provides a map implementation of a key value store.
// It offers a lightweight drop-in replacement of  hive.go/kvstore for tests or in simulations
// where more than one instance is required.
package mapdb

import (
	"sync"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/types"
)

// mapDB is a simple implementation of KVStore using a map.
type mapDB struct {
	sync.RWMutex
	m     *syncedKVMap
	realm []byte
}

// NewMapDB creates a kvstore.KVStore implementation purely based on a go map.
func NewMapDB() kvstore.KVStore {
	return &mapDB{
		m: &syncedKVMap{m: make(map[string][]byte)},
	}
}

func (s *mapDB) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &mapDB{
		m:     s.m, // use the same underlying map
		realm: realm,
	}
}

func (s *mapDB) Realm() kvstore.Realm {
	return byteutils.ConcatBytes(s.realm)
}

// Shutdown marks the store as shutdown.
func (s *mapDB) Shutdown() {
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *mapDB) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	s.m.iterate(s.realm, prefix, consumerFunc, iterDirection...)
	return nil
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *mapDB) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	s.m.iterateKeys(s.realm, prefix, consumerFunc, iterDirection...)
	return nil
}

func (s *mapDB) Clear() error {
	s.Lock()
	defer s.Unlock()

	s.m.deletePrefix(s.realm)
	return nil
}

func (s *mapDB) Get(key kvstore.Key) (kvstore.Value, error) {
	s.RLock()
	defer s.RUnlock()

	value, contains := s.m.get(byteutils.ConcatBytes(s.realm, key))
	if !contains {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (s *mapDB) Set(key kvstore.Key, value kvstore.Value) error {
	s.Lock()
	defer s.Unlock()

	return s.set(key, value)
}

func (s *mapDB) set(key kvstore.Key, value kvstore.Value) error {
	s.m.set(byteutils.ConcatBytes(s.realm, key), value)
	return nil
}

func (s *mapDB) Has(key kvstore.Key) (bool, error) {
	s.RLock()
	defer s.RUnlock()

	contains := s.m.has(byteutils.ConcatBytes(s.realm, key))
	return contains, nil
}

func (s *mapDB) Delete(key kvstore.Key) error {
	s.Lock()
	defer s.Unlock()

	return s.delete(key)
}

func (s *mapDB) delete(key kvstore.Key) error {
	s.m.delete(byteutils.ConcatBytes(s.realm, key))
	return nil
}

func (s *mapDB) DeletePrefix(prefix kvstore.KeyPrefix) error {
	s.Lock()
	defer s.Unlock()

	s.m.deletePrefix(byteutils.ConcatBytes(s.realm, prefix))
	return nil
}

func (s *mapDB) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		kvStore:          s,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

func (s *mapDB) Flush() error {
	return nil
}

func (s *mapDB) Close() error {
	return nil
}

type kvtuple struct {
	key   kvstore.Key
	value kvstore.Value
}

// batchedMutations is a wrapper to do a batched update on a mapDB.
type batchedMutations struct {
	sync.Mutex
	kvStore          *mapDB
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty

	sets    []kvtuple
	deletes []kvtuple
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	stringKey := byteutils.ConcatBytesToString(key)

	b.Lock()
	defer b.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	stringKey := byteutils.ConcatBytesToString(key)

	b.Lock()
	defer b.Unlock()

	delete(b.setOperations, stringKey)
	b.deleteOperations[stringKey] = types.Void

	return nil
}

func (b *batchedMutations) Cancel() {
	b.Lock()
	defer b.Unlock()

	b.setOperations = make(map[string]kvstore.Value)
	b.deleteOperations = make(map[string]types.Empty)
}

func (b *batchedMutations) Commit() error {
	b.Lock()
	b.kvStore.Lock()
	defer b.kvStore.Unlock()
	defer b.Unlock()

	for key, value := range b.setOperations {
		err := b.kvStore.set([]byte(key), value)
		if err != nil {
			return err
		}
	}

	for key := range b.deleteOperations {
		err := b.kvStore.delete([]byte(key))
		if err != nil {
			return err
		}
	}

	return nil
}
