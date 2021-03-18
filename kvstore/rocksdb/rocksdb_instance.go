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

type RocksDBOptions struct {
	compression bool
	fillCache   bool
	sync        bool
}

type RocksDBOption func(*RocksDBOptions)

// UseCompression sets opts.SetCompression(grocksdb.ZSTDCompression)
func UseCompression(compression bool) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.compression = compression
	}
}

// ReadFillCache sets the opts.SetFillCache ReadOption
func ReadFillCache(fillCache bool) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.fillCache = fillCache
	}
}

// WriteSync sets the opts.SetSync WriteOption
func WriteSync(sync bool) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.sync = sync
	}
}

func dbOptions(optionalOptions []RocksDBOption) *RocksDBOptions {
	result := &RocksDBOptions{}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}
	return result
}

// NewRocksDB creates a new RocksDB instance.
func CreateDB(directory string, options ...RocksDBOption) (*RocksDB, error) {

	if err := checkDir(directory); err != nil {
		return nil, fmt.Errorf("could not check directory: %w", err)
	}

	dbOpts := dbOptions(options)

	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCompression(grocksdb.NoCompression)
	if dbOpts.compression {
		opts.SetCompression(grocksdb.ZSTDCompression)
	}

	ro := grocksdb.NewDefaultReadOptions()
	ro.SetFillCache(dbOpts.fillCache)

	wo := grocksdb.NewDefaultWriteOptions()
	wo.SetSync(dbOpts.sync)

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

func (r *RocksDB) Flush() error {
	return r.db.Flush(r.fo)
}

func (r *RocksDB) Close() error {
	r.db.Close()
	return nil
}
