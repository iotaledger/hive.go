package rocksdb

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

type RocksDB struct {
	db *grocksdb.DB
	ro *grocksdb.ReadOptions
	wo *grocksdb.WriteOptions
	fo *grocksdb.FlushOptions
}

// NewRocksDB creates a new RocksDB instance.
func CreateDB(directory string, opts *grocksdb.Options, ro *grocksdb.ReadOptions, wo *grocksdb.WriteOptions, fo *grocksdb.FlushOptions) (*RocksDB, error) {

	if err := checkDir(directory); err != nil {
		return nil, fmt.Errorf("could not check directory: %w", err)
	}

	if opts == nil {
		opts = grocksdb.NewDefaultOptions()
	}
	opts.SetCreateIfMissing(true)

	if ro == nil {
		ro = grocksdb.NewDefaultReadOptions()
	}

	if wo == nil {
		wo = grocksdb.NewDefaultWriteOptions()
	}

	if fo == nil {
		fo = grocksdb.NewDefaultFlushOptions()
	}

	db, err := grocksdb.OpenDb(opts, directory)
	if err != nil {
		return nil, err
	}

	return &RocksDB{
		db: db,
		ro: ro,
		wo: wo,
		fo: fo,
	}, nil
}

func (r *RocksDB) Flush() error {
	r.db.Flush(r.fo)
	return nil
}

func (r *RocksDB) Close() error {
	r.db.Close()
	return nil
}
