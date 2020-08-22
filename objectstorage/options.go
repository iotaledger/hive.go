package objectstorage

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/mr-tron/base58"
)

type Options struct {
	batchedWriterInstance *BatchedWriter
	cacheTime             time.Duration
	keyPartitions         []int
	persistenceEnabled    bool
	keysOnly              bool
	storeOnCreation       bool
	leakDetectionOptions  *LeakDetectionOptions
	leakDetectionWrapper  func(cachedObject *CachedObjectImpl) LeakDetectionWrapper
	delayedOptions        []func()
}

func newOptions(objectStorage *ObjectStorage, optionalOptions []Option) *Options {
	result := &Options{
		cacheTime:          0,
		persistenceEnabled: true,
		delayedOptions:     make([]func(), 0),
	}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	if result.leakDetectionOptions != nil && result.leakDetectionWrapper == nil {
		result.leakDetectionWrapper = newLeakDetectionWrapperImpl
	}

	if result.batchedWriterInstance == nil {
		result.batchedWriterInstance = NewBatchedWriter(objectStorage.store)
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
	command    kvstore.Command
	parameters [][]byte
}

// String returns a string representation of the log entry
func (l *logEntry) String() string {
	result := kvstore.CommandNames[l.command]
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
			args.batchedWriterInstance.store.AccessCallback(func(command kvstore.Command, parameters ...[]byte) {
				logChannel <- logEntry{command, parameters}
			}, commandsFilter...)
		})
	}
}

func BatchedWriterInstance(batchedWriterInstance *BatchedWriter) Option {
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
