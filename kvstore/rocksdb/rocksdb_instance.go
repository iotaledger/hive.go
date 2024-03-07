//go:build rocksdb

package rocksdb

import (
	"github.com/iotaledger/grocksdb"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/ioutils"
)

// RocksDB holds the underlying grocksdb.DB instance and options.
type RocksDB struct {
	db *grocksdb.DB
	ro *grocksdb.ReadOptions
	wo *grocksdb.WriteOptions
	fo *grocksdb.FlushOptions
}

// CreateDB creates a new RocksDB instance.
func CreateDB(directory string, options ...Option) (*RocksDB, error) {
	if err := ioutils.CreateDirectory(directory, 0o700); err != nil {
		return nil, ierrors.Wrapf(err, "could not create directory '%s'", directory)
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
			return nil, ierrors.Wrapf(err, "could not get options from string '%s'", str)
		}
	}

	ro := grocksdb.NewDefaultReadOptions()
	ro.SetFillCache(dbOpts.fillCache)

	wo := grocksdb.NewDefaultWriteOptions()
	wo.SetSync(dbOpts.sync)
	wo.DisableWAL(dbOpts.disableWAL)

	fo := grocksdb.NewDefaultFlushOptions()

	if dbOpts.blockCacheSize != 0 {
		bbto := grocksdb.NewDefaultBlockBasedTableOptions()
		bbto.SetBlockCache(grocksdb.NewLRUCache(dbOpts.blockCacheSize))
		opts.SetBlockBasedTableFactory(bbto)
	}

	db, err := grocksdb.OpenDb(opts, directory)
	if err != nil {
		return nil, ierrors.Wrapf(err, "could not open new DB '%s'", directory)
	}

	return &RocksDB{
		db: db,
		ro: ro,
		wo: wo,
		fo: fo,
	}, nil
}

// OpenDBReadOnly opens a new RocksDB instance in read-only mode.
func OpenDBReadOnly(directory string, options ...Option) (*RocksDB, error) {
	dbOpts := dbOptions(options)

	opts := grocksdb.NewDefaultOptions()
	opts.SetCompression(grocksdb.NoCompression)
	if dbOpts.compression {
		opts.SetCompression(grocksdb.ZSTDCompression)
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

	db, err := grocksdb.OpenDbForReadOnly(opts, directory, true)
	if err != nil {
		return nil, err
	}

	return &RocksDB{
		db: db,
		ro: ro,
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
