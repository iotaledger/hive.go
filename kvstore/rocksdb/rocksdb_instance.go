//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"fmt"

	"github.com/linxGnu/grocksdb"

	"github.com/iotaledger/hive.go/kvstore/utils"
)

// RocksDB holds the underlying grocksdb.DB instance and options
type RocksDB struct {
	db *grocksdb.DB
	ro *grocksdb.ReadOptions
	wo *grocksdb.WriteOptions
	fo *grocksdb.FlushOptions
}

// CreateDB creates a new RocksDB instance.
func CreateDB(directory string, options ...Option) (*RocksDB, error) {

	if err := utils.CreateDirectory(directory, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
	}

	dbOpts := dbOptions(options)

	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCompression(grocksdb.NoCompression)
	if dbOpts.compression {
		opts.SetCompression(grocksdb.ZSTDCompression)
	}

	if dbOpts.parallelism > 0 {
		opts.IncreaseParallelism(dbOpts.parallelism)
	}

	for _, str := range dbOpts.custom {
		var err error
		opts, err = grocksdb.GetOptionsFromString(opts, str)
		if err != nil {
			return nil, err
		}
	}

	ro := grocksdb.NewDefaultReadOptions()
	ro.SetFillCache(dbOpts.fillCache)

	wo := grocksdb.NewDefaultWriteOptions()
	wo.SetSync(dbOpts.sync)
	wo.DisableWAL(dbOpts.disableWAL)

	fo := grocksdb.NewDefaultFlushOptions()

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

func dbOptions(optionalOptions []Option) *Options {
	result := &Options{
		compression: false,
		fillCache:   false,
		sync:        false,
		disableWAL:  true,
		parallelism: 0,
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}
	return result
}

// Flush the database.
func (r *RocksDB) Flush() error {
	return r.db.Flush(r.fo)
}

// Close the database.
func (r *RocksDB) Close() error {
	r.db.Close()
	return nil
}

// GetProperty returns the value of a database property.
func (r *RocksDB) GetProperty(name string) string {
	return r.db.GetProperty(name)
}

// GetIntProperty similar to "GetProperty", but only works for a subset of properties whose
// return value is an integer. Return the value by integer.
func (r *RocksDB) GetIntProperty(name string) (uint64, bool) {
	return r.db.GetIntProperty(name)
}
