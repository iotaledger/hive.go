// Package mapdb provides a map implementation of a key value store.
// It offers a lightweight drop-in replacement of  hive.go/kvstore for tests or in simulations
// where more than one instance is required.
package mapdb

import (
	"sync"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
)

// MapDB is a simple implementation of KVStore using a map.
type MapDB struct {
	m     *syncedKVMap
	realm []byte
}

// NewMapDB creates a kvstore.KVStore implementation purely based on a go map.
func NewMapDB() *MapDB {
	return &MapDB{
		m: &syncedKVMap{m: make(map[string][]byte)},
	}
}

func (db *MapDB) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &MapDB{
		m:     db.m, // use the same underlying map
		realm: realm,
	}
}

func (db *MapDB) Realm() kvstore.Realm {
	return byteutils.ConcatBytes(db.realm)
}

func (db *MapDB) Has(key kvstore.Key) (bool, error) {
	contains := db.m.has(byteutils.ConcatBytes(db.realm, key))
	return contains, nil
}

func (db *MapDB) Get(key kvstore.Key) (kvstore.Value, error) {
	value, contains := db.m.get(byteutils.ConcatBytes(db.realm, key))
	if !contains {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (db *MapDB) Set(key kvstore.Key, value kvstore.Value) error {
	db.m.set(byteutils.ConcatBytes(db.realm, key), value)
	return nil
}

func (db *MapDB) Delete(key kvstore.Key) error {
	db.m.delete(byteutils.ConcatBytes(db.realm, key))
	return nil
}

func (db *MapDB) DeletePrefix(keyPrefix kvstore.KeyPrefix) error {
	db.m.deletePrefix(byteutils.ConcatBytes(db.realm, keyPrefix))
	return nil
}

func (db *MapDB) Iterate(keyPrefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	db.m.iterate(db.realm, keyPrefix, consumerFunc)
	return nil
}

func (db *MapDB) IterateKeys(keyPrefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	db.m.iterateKeys(db.realm, keyPrefix, consumerFunc)
	return nil
}

func (db *MapDB) Clear() error {
	db.m.deletePrefix(db.realm)
	return nil
}

func (db *MapDB) Batched() kvstore.BatchedMutations {
	return &BatchedMutations{
		db: db,
	}
}

type kvtuple struct {
	key   kvstore.Key
	value kvstore.Value
}

// BatchedMutations is a wrapper to do a batched update on a MapDB.
type BatchedMutations struct {
	sync.Mutex
	db      *MapDB
	sets    []kvtuple
	deletes []kvtuple
}

func (b *BatchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	b.Lock()
	defer b.Unlock()
	b.sets = append(b.sets, kvtuple{key, value})
	return nil
}

func (b *BatchedMutations) Delete(key kvstore.Key) error {
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
		if err := b.db.Set(b.sets[i].key, b.sets[i].value); err != nil {
			return err
		}
	}
	for i := 0; i < len(b.deletes); i++ {
		if err := b.db.Delete(b.deletes[i].key); err != nil {
			return err
		}
	}
	return nil
}
