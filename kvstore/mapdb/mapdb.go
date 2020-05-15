// Package mapdb provides a map implementation of a key value store.
// It offers a lightweight drop-in replacement of  hive.go/kvstore for tests or in simulations
// where more than one instance is required.
package mapdb

import (
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/kvstore"
)

// MapDB is a simple implementation of KVStore using a map.
type MapDB struct {
	mu    sync.RWMutex
	m     map[string]mapEntry
	realm []byte
}

type mapEntry struct {
	value []byte
}

// NewMapDB creates a database.Database implementation purely based on a go map.
// MapDB does not support TTL.
func NewMapDB() *MapDB {
	return &MapDB{
		m: make(map[string]mapEntry),
	}
}

func (db *MapDB) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &MapDB{
		m:     make(map[string]mapEntry),
		realm: realm,
	}
}

func (db *MapDB) Realm() kvstore.Realm {
	return db.realm
}

func (db *MapDB) Has(key kvstore.Key) (contains bool, err error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, contains = db.m[string(key)]
	return
}

func (db *MapDB) Get(key kvstore.Key) (value kvstore.Value, err error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ent, contains := db.m[string(key)]
	if !contains {
		err = kvstore.ErrKeyNotFound
		return
	}
	value = append([]byte{}, ent.value...)
	return
}

func (db *MapDB) Set(key kvstore.Key, value kvstore.Value) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.m[string(key)] = mapEntry{
		value: append([]byte{}, value...),
	}
	return nil
}

func (db *MapDB) Delete(key kvstore.Key) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.m, string(key))
	return nil
}

func (db *MapDB) DeletePrefix(keyPrefix kvstore.KeyPrefix) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	prefix := string(keyPrefix)
	for key := range db.m {
		if strings.HasPrefix(key, prefix) {
			delete(db.m, key)
		}
	}
	return nil
}

func (db *MapDB) buildKeyPrefix(prefixes []kvstore.KeyPrefix) string {
	var prefix []byte
	for _, optionalPrefix := range prefixes {
		prefix = append(prefix, optionalPrefix...)
	}
	return string(prefix)
}

func (db *MapDB) Iterate(prefixes []kvstore.KeyPrefix, preFetchValues bool, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if len(prefixes) > 0 {
		prefix := db.buildKeyPrefix(prefixes)
		for key, ent := range db.m {
			if strings.HasPrefix(key, prefix) {
				if !kvConsumerFunc([]byte(key), append([]byte{}, ent.value...)) {
					break
				}
			}
		}
		return nil
	}

	for key, ent := range db.m {
		if !kvConsumerFunc([]byte(key), append([]byte{}, ent.value...)) {
			break
		}
	}

	return nil
}

func (db *MapDB) IterateKeys(prefixes []kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if len(prefixes) > 0 {
		prefix := db.buildKeyPrefix(prefixes)
		for key := range db.m {
			if strings.HasPrefix(key, prefix) {
				if !consumerFunc([]byte(key)) {
					break
				}
			}
		}
		return nil
	}

	for key := range db.m {
		if !consumerFunc([]byte(key)) {
			break
		}
	}

	return nil
}

func (db *MapDB) Clear() error {
	db.m = make(map[string]mapEntry)
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

// BatchedMutations is a wrapper to do a batched update on a BoltDB.
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
		b.db.Set(b.sets[i].key, b.sets[i].value)
	}
	for i := 0; i < len(b.deletes); i++ {
		b.db.Delete(b.deletes[i].key)
	}
	return nil
}
