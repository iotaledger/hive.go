package rocksdb

type RocksDBOptions struct {
	compression bool
	fillCache   bool
	sync        bool
	custom      []string
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
