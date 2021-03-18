package rocksdb

const (
	panicMissingRocksDB = "For RocksDB use the hive.go feat/RocksDB branch"
)

type RocksDB struct {
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
	panic(panicMissingRocksDB)
}

func (r *RocksDB) Flush() error {
	panic(panicMissingRocksDB)
}

func (r *RocksDB) Close() error {
	panic(panicMissingRocksDB)
}
