// +build !rocksdb

package rocksdb

import (
	"github.com/iotaledger/hive.go/kvstore"
)

type rocksDBStore struct {
}

// New creates a new KVStore with the underlying RocksDB.
func New(db *RocksDB) kvstore.KVStore {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Realm() []byte {
	panic(panicMissingRocksDB)
}

// Shutdown marks the store as shutdown.
func (s *rocksDBStore) Shutdown() {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Clear() error {
	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *rocksDBStore) Get(key kvstore.Key) (kvstore.Value, error) {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Set(key kvstore.Key, value kvstore.Value) error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Has(key kvstore.Key) (bool, error) {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Delete(key kvstore.Key) error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Batched() kvstore.BatchedMutations {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Flush() error {
	panic(panicMissingRocksDB)
}

func (s *rocksDBStore) Close() error {
	panic(panicMissingRocksDB)
}
