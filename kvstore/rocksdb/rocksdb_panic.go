//go:build !rocksdb

package rocksdb

import "github.com/iotaledger/hive.go/kvstore"

const (
	panicMissingRocksDB = "For RocksDB support please compile with '-tags rocksdb'"
)

// RocksDB holds the underlying grocksdb.DB instance and options.
type RocksDB struct {
}

// CreateDB creates a new RocksDB instance.
func CreateDB(_ string, _ ...Option) (*RocksDB, error) {
	panic(panicMissingRocksDB)
}

// OpenDBReadOnly opens a new RocksDB instance in read-only mode.
func OpenDBReadOnly(_ string, _ ...Option) (*RocksDB, error) {
	panic(panicMissingRocksDB)
}

// New creates a new KVStore with the underlying RocksDB.
func New(_ *RocksDB) kvstore.KVStore {
	panic(panicMissingRocksDB)
}

// Flush the database.
func (r *RocksDB) Flush() error {
	panic(panicMissingRocksDB)
}

// Close the database.
func (r *RocksDB) Close() error {
	panic(panicMissingRocksDB)
}

// GetProperty returns the value of a database property.
func (r *RocksDB) GetProperty(_ string) string {
	panic(panicMissingRocksDB)
}

// GetIntProperty similar to "GetProperty", but only works for a subset of properties whose
// return value is an integer. Return the value by integer.
func (r *RocksDB) GetIntProperty(_ string) (uint64, bool) {
	panic(panicMissingRocksDB)
}
