//go:build !rocksdb
// +build !rocksdb

package rocksdb

import "github.com/iotaledger/hive.go/kvstore"

const (
	panicMissingRocksDB = "For RocksDB support please compile with '-tags rocksdb'"
)

// RocksDB holds the underlying grocksdb.DB instance and options
type RocksDB struct {
}

// CreateDB creates a new RocksDB instance.
func CreateDB(directory string, options ...Option) (*RocksDB, error) {
	panic(panicMissingRocksDB)
}

func (r *RocksDB) Flush() error {
	panic(panicMissingRocksDB)
}

func (r *RocksDB) Close() error {
	panic(panicMissingRocksDB)
}

// New creates a new KVStore with the underlying RocksDB.
func New(db *RocksDB) kvstore.KVStore {
	panic(panicMissingRocksDB)
}

// GetProperty returns the value of a database property.
func (r *RocksDB) GetProperty(name string) string {
	panic(panicMissingRocksDB)
}

// GetIntProperty similar to "GetProperty", but only works for a subset of properties whose
// return value is an integer. Return the value by integer.
func (r *RocksDB) GetIntProperty(name string) (uint64, bool) {
	panic(panicMissingRocksDB)
}
