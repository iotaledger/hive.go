package rocksdb

// RocksDBOptions holds the options used to instantiate the underlying grocksdb.DB
type RocksDBOptions struct {
	compression bool
	fillCache   bool
	sync        bool
	custom      []string
	parallelism int
}

// RocksDBOption is one of the RocksDBOptions
type RocksDBOption func(*RocksDBOptions)

// UseCompression sets opts.SetCompression(grocksdb.ZSTDCompression)
func UseCompression(compression bool) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.compression = compression
	}
}

// IncreaseParallelism sets opts.IncreaseParallelism(thread_count)
func IncreaseParallelism(thread_count int) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.parallelism = thread_count
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

// Custom passes the given string to GetOptionsFromString
func Custom(options []string) RocksDBOption {
	return func(args *RocksDBOptions) {
		args.custom = options
	}
}

func dbOptions(optionalOptions []RocksDBOption) *RocksDBOptions {
	result := &RocksDBOptions{}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}
	return result
}
