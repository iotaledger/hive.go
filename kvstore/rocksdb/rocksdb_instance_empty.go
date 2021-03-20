// +build !rocksdb

package rocksdb

const (
	panicMissingRocksDB = "For RocksDB support please compile with '-tags rocksdb'"
)

type RocksDB struct {
}

// NewRocksDB creates a new RocksDB instance.
func CreateDB(directory string, options ...RocksDBOption) (*RocksDB, error) {

	panic(panicMissingRocksDB)
}

func (r *RocksDB) Flush() error {
	panic(panicMissingRocksDB)
}

func (r *RocksDB) Close() error {
	panic(panicMissingRocksDB)
}
