package rocksdb

// Options holds the options used to instantiate the underlying grocksdb.DB.
type Options struct {
	compression    bool
	fillCache      bool
	sync           bool
	disableWAL     bool
	parallelism    int
	blockCacheSize uint64
	custom         []string
}

// Option is one of the Options.
type Option func(*Options)

// UseCompression sets opts.SetCompression(grocksdb.ZSTDCompression).
func UseCompression(compression bool) Option {
	return func(args *Options) {
		args.compression = compression
	}
}

// IncreaseParallelism sets opts.IncreaseParallelism(threadCount).
func IncreaseParallelism(threadCount int) Option {
	return func(args *Options) {
		args.parallelism = threadCount
	}
}

// ReadFillCache sets the opts.SetFillCache ReadOption.
func ReadFillCache(fillCache bool) Option {
	return func(args *Options) {
		args.fillCache = fillCache
	}
}

// WriteSync sets the opts.SetSync WriteOption.
func WriteSync(sync bool) Option {
	return func(args *Options) {
		args.sync = sync
	}
}

// WriteDisableWAL sets the opts.DisableWAL WriteOption.
func WriteDisableWAL(value bool) Option {
	return func(args *Options) {
		args.disableWAL = value
	}
}

// BlockCacheSize sets the size in bytes of the LRU cache for grocksdb blocks.
func BlockCacheSize(size uint64) Option {
	return func(args *Options) {
		args.blockCacheSize = size
	}
}

// Custom passes the given string to GetOptionsFromString.
func Custom(options []string) Option {
	return func(args *Options) {
		args.custom = options
	}
}
