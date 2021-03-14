package objectstorage

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/hive.go/kvstore"
)

type Options struct {
	batchedWriterInstance      *kvstore.BatchedWriter
	cacheTime                  time.Duration
	keyPartitions              []int
	persistenceEnabled         bool
	keysOnly                   bool
	storeOnCreation            bool
	releaseExecutorWorkerCount int
	leakDetectionOptions       *LeakDetectionOptions
	leakDetectionWrapper       func(cachedObject *CachedObjectImpl) LeakDetectionWrapper
	delayedOptions             []func()
	onEvictionCallback         func(cachedObject CachedObject)
}

func newOptions(store kvstore.KVStore, optionalOptions []Option) *Options {
	result := &Options{
		cacheTime:                  0,
		persistenceEnabled:         true,
		releaseExecutorWorkerCount: 1,
		delayedOptions:             make([]func(), 0),
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	if result.leakDetectionOptions != nil && result.leakDetectionWrapper == nil {
		result.leakDetectionWrapper = newLeakDetectionWrapperImpl
	}

	if result.batchedWriterInstance == nil {
		result.batchedWriterInstance = kvstore.NewBatchedWriter(store)
	}

	for _, delayedOption := range result.delayedOptions {
		delayedOption()
	}

	return result
}

func (options *Options) delayed(callback func()) {
	options.delayedOptions = append(options.delayedOptions, callback)
}

type Option func(*Options)

func CacheTime(duration time.Duration) Option {
	return func(args *Options) {
		args.cacheTime = duration
	}
}

// logChannelBufferSize defines the size of the buffer used for the log writer
const logChannelBufferSize = 10240

// logEntry is a container for the
type logEntry struct {
	time       time.Time
	command    kvstore.Command
	parameters [][]byte
}

// String returns a string representation of the log entry
func (l *logEntry) String() string {
	result := l.time.Format("15:04:05") + " " + kvstore.CommandNames[l.command]
	for _, parameter := range l.parameters {
		result += " " + base58.Encode(parameter)
	}

	return result
}

// LogAccess sets up a logger that logs all calls to the underlying store in the given file. It is possible to filter
// the logged commands by providing an optional filter flag.
func LogAccess(fileName string, commandsFilter ...kvstore.Command) Option {
	return func(args *Options) {
		// execute this function after the remaining options have been initialized and a BatchWriter exists
		args.delayed(func() {
			// open log file
			logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			writer := bufio.NewWriter(logFile)
			if err != nil {
				panic(err)
			}

			// open logger channel
			logChannel := make(chan logEntry, logChannelBufferSize)

			// start background worker that writes to the log file
			go func() {
				for {
					switch loggedCommand := <-logChannel; loggedCommand.command {
					case kvstore.ShutdownCommand:
						// write log entry
						if _, err = writer.WriteString(loggedCommand.String() + "\n"); err != nil {
							panic(err)
						}

						// close channel and log file
						err = writer.Flush()
						if err != nil {
							fmt.Println(err)
						}
						close(logChannel)
						err = logFile.Close()
						if err != nil {
							fmt.Println(err)
						}

						return

					default:
						// write log entry
						if _, err := writer.WriteString(loggedCommand.String() + "\n"); err != nil {
							panic(err)
						}
					}
				}
			}()

			// pass through calls to logger channel
			args.batchedWriterInstance.KVStore().AccessCallback(func(command kvstore.Command, parameters ...[]byte) {
				logChannel <- logEntry{time.Now(), command, parameters}
			}, commandsFilter...)
		})
	}
}

func BatchedWriterInstance(batchedWriterInstance *kvstore.BatchedWriter) Option {
	return func(args *Options) {
		args.batchedWriterInstance = batchedWriterInstance
	}
}

func PersistenceEnabled(persistenceEnabled bool) Option {
	return func(args *Options) {
		args.persistenceEnabled = persistenceEnabled
	}
}

func KeysOnly(keysOnly bool) Option {
	return func(args *Options) {
		args.keysOnly = keysOnly
	}
}

// StoreOnCreation writes an object directly to the persistence layer on creation.
func StoreOnCreation(store bool) Option {
	return func(args *Options) {
		args.storeOnCreation = store
	}
}

// ReleaseExecutorWorkerCount sets the number of workers that execute the
// scheduled eviction of the objects in parallel (whenever they become due).
func ReleaseExecutorWorkerCount(releaseExecutorWorkerCount int) Option {
	if releaseExecutorWorkerCount < 1 {
		panic("releaseExecutorWorkerCount must be greater or equal 1")
	}

	return func(args *Options) {
		args.releaseExecutorWorkerCount = releaseExecutorWorkerCount
	}
}

func LeakDetectionEnabled(leakDetectionEnabled bool, options ...LeakDetectionOptions) Option {
	return func(args *Options) {
		if leakDetectionEnabled {
			switch len(options) {
			case 0:
				args.leakDetectionOptions = &LeakDetectionOptions{
					MaxConsumersPerObject: 20,
					MaxConsumerHoldTime:   240 * time.Second,
				}
			case 1:
				args.leakDetectionOptions = &options[0]
			default:
				panic("too many additional arguments in call to LeakDetectionEnabled (only 0 or 1 allowed")
			}
		}
	}
}

func OverrideLeakDetectionWrapper(wrapperFunc func(cachedObject *CachedObjectImpl) LeakDetectionWrapper) Option {
	return func(args *Options) {
		args.leakDetectionWrapper = wrapperFunc
	}
}

func PartitionKey(keyPartitions ...int) Option {
	return func(args *Options) {
		args.keyPartitions = keyPartitions
	}
}

// OnEvictionCallback sets a function that is called on eviction of the object.
func OnEvictionCallback(cb func(cachedObject CachedObject)) Option {
	return func(args *Options) {
		args.onEvictionCallback = cb
	}
}

// the default options used for object storage iteration.
var defaultIteratorOptions = []IteratorOption{
	WithSkipCache(false),
	WithSkipStorage(false),
	WithPrefix(kvstore.EmptyPrefix),
}

// IteratorOption is a function setting an iterator option.
type IteratorOption func(opts *IteratorOptions)

// IteratorOptions define options for iterations in the object storage.
type IteratorOptions struct {
	// whether to skip the elements in the cache.
	skipCache bool
	// whether to skip the elements in the storage.
	skipStorage bool
	// an optional prefix to iterate a subset of elements.
	optionalPrefix []byte
}

// applies the given IteratorOption.
func (o *IteratorOptions) apply(opts ...IteratorOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithSkipCache is used to skip the elements in the cache.
func WithSkipCache(skipCache bool) IteratorOption {
	return func(opts *IteratorOptions) {
		opts.skipCache = skipCache
	}
}

// WithSkipStorage is used to skip the elements in the storage.
func WithSkipStorage(skipStorage bool) IteratorOption {
	return func(opts *IteratorOptions) {
		opts.skipStorage = skipStorage
	}
}

// WithPrefix is used to iterate a subset of elements with a defined prefix.
func WithPrefix(prefix []byte) IteratorOption {
	return func(opts *IteratorOptions) {
		opts.optionalPrefix = prefix
	}
}
