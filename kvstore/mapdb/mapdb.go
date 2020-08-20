// Package mapdb provides a map implementation of a key value store.
// It offers a lightweight drop-in replacement of  hive.go/kvstore for tests or in simulations
// where more than one instance is required.
package mapdb

import (
	"sync"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
)

// mapDB is a simple implementation of KVStore using a map.
type mapDB struct {
	m                            *syncedKVMap
	realm                        []byte
	accessCallback               kvstore.AccessCallback
	accessCallbackCommandsFilter kvstore.Command
}

// NewMapDB creates a kvstore.KVStore implementation purely based on a go map.
func NewMapDB() *mapDB {
	return &mapDB{
		m: &syncedKVMap{m: make(map[string][]byte)},
	}
}

// AccessCallback configures the store to pass all requests to the KVStore to the given callback.
// This can for example be used for debugging and to examine what the KVStore is doing.
func (s *mapDB) AccessCallback(callback kvstore.AccessCallback, commandsFilter ...kvstore.Command) {
	var accessCallbackCommandsFilter kvstore.Command
	if len(commandsFilter) == 0 {
		accessCallbackCommandsFilter = kvstore.AllCommands
	} else {
		for _, filterCommand := range commandsFilter {
			accessCallbackCommandsFilter |= filterCommand
		}
	}

	s.accessCallback = callback
	s.accessCallbackCommandsFilter = accessCallbackCommandsFilter
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

func (s *mapDB) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateCommand) {
		s.accessCallback(kvstore.IterateCommand, prefix)
	}

	s.m.iterate(s.realm, prefix, consumerFunc)
	return nil
}

func (s *mapDB) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateKeysCommand) {
		s.accessCallback(kvstore.IterateKeysCommand, prefix)
	}

	s.m.iterateKeys(s.realm, prefix, consumerFunc)
	return nil
}

func (s *mapDB) Clear() error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.ClearCommand) {
		s.accessCallback(kvstore.ClearCommand)
	}

	s.m.deletePrefix(s.realm)
	return nil
}

func (s *mapDB) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.GetCommand) {
		s.accessCallback(kvstore.GetCommand, key)
	}

	value, contains := s.m.get(byteutils.ConcatBytes(s.realm, key))
	if !contains {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (s *mapDB) Set(key kvstore.Key, value kvstore.Value) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		s.accessCallback(kvstore.SetCommand, key, value)
	}

	s.m.set(byteutils.ConcatBytes(s.realm, key), value)
	return nil
}

func (s *mapDB) Has(key kvstore.Key) (bool, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.HasCommand) {
		s.accessCallback(kvstore.HasCommand, key)
	}

	contains := s.m.has(byteutils.ConcatBytes(s.realm, key))
	return contains, nil
}

func (s *mapDB) Delete(key kvstore.Key) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		s.accessCallback(kvstore.DeleteCommand, key)
	}

	s.m.delete(byteutils.ConcatBytes(s.realm, key))
	return nil
}

func (s *mapDB) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeletePrefixCommand) {
		s.accessCallback(kvstore.DeletePrefixCommand, prefix)
	}

	s.m.deletePrefix(byteutils.ConcatBytes(s.realm, prefix))
	return nil
}

func (s *mapDB) Batched() kvstore.BatchedMutations {
	return &BatchedMutations{
		kvStore: s,
	}
}

type kvtuple struct {
	key   kvstore.Key
	value kvstore.Value
}

// BatchedMutations is a wrapper to do a batched update on a mapDB.
type BatchedMutations struct {
	sync.Mutex
	kvStore *mapDB
	sets    []kvtuple
	deletes []kvtuple
}

func (b *BatchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		b.kvStore.accessCallback(kvstore.SetCommand, key, value)
	}

	b.Lock()
	defer b.Unlock()
	b.sets = append(b.sets, kvtuple{key, value})
	return nil
}

func (b *BatchedMutations) Delete(key kvstore.Key) error {
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		b.kvStore.accessCallback(kvstore.DeleteCommand, key)
	}

	b.Lock()
	defer b.Unlock()
	b.deletes = append(b.deletes, kvtuple{key, nil})
	return nil
}

func (b *BatchedMutations) Cancel() {
	// do nothing
}

func (b *BatchedMutations) Commit() error {
	b.Lock()
	defer b.Unlock()

	for i := 0; i < len(b.sets); i++ {
		if err := b.kvStore.Set(b.sets[i].key, b.sets[i].value); err != nil {
			return err
		}
	}
	for i := 0; i < len(b.deletes); i++ {
		if err := b.kvStore.Delete(b.deletes[i].key); err != nil {
			return err
		}
	}
	return nil
}
