package database

import (
	"os"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/pkg/errors"
)

var (
	instance  *badger.DB
	once      sync.Once
	directory = "mainnetdb"
	light     bool
)

// Returns whether the given file or directory exists.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Settings sets DB dir and light mode
func Settings(dir string, lightMode bool) {
	directory = dir
	light = lightMode
}

func checkDir(dir string) error {
	exists, err := exists(dir)
	if err != nil {
		return err
	}

	if !exists {
		return os.Mkdir(dir, 0700)
	}
	return nil
}

func createDB() (*badger.DB, error) {
	if err := checkDir(directory); err != nil {
		return nil, errors.Wrap(err, "Could not check directory")
	}

	opts := badger.DefaultOptions(directory)
	opts.Logger = nil
	opts.Truncate = true

	if light {
		opts.LevelOneSize = 256 << 18
		opts.LevelSizeMultiplier = 10
		opts.TableLoadingMode = options.FileIO
		opts.ValueLogLoadingMode = options.FileIO

		opts.MaxLevels = 5
		opts.MaxTableSize = 64 << 18
		opts.NumCompactors = 1 // Compactions can be expensive. Only run 2.
		opts.NumLevelZeroTables = 1
		opts.NumLevelZeroTablesStall = 2
		opts.NumMemtables = 1
		opts.SyncWrites = false
		opts.NumVersionsToKeep = 1
		opts.CompactL0OnClose = true

		opts.ValueLogFileSize = 1<<25 - 1

		opts.ValueLogMaxEntries = 250000
		opts.ValueThreshold = 32
		opts.Truncate = false
		opts.LogRotatesToFlush = 2
	} else {
		opts.LevelOneSize = 256 << 20
		opts.LevelSizeMultiplier = 10
		opts.TableLoadingMode = options.MemoryMap
		opts.ValueLogLoadingMode = options.MemoryMap

		opts.MaxLevels = 7
		opts.MaxTableSize = 64 << 20
		opts.NumCompactors = 2 // Compactions can be expensive. Only run 2.
		opts.NumLevelZeroTables = 5
		opts.NumLevelZeroTablesStall = 10
		opts.NumMemtables = 5
		opts.SyncWrites = true
		opts.NumVersionsToKeep = 1
		opts.CompactL0OnClose = true

		opts.ValueLogFileSize = 1<<30 - 1

		opts.ValueLogMaxEntries = 1000000
		opts.ValueThreshold = 32
		opts.Truncate = false
		opts.LogRotatesToFlush = 2
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open new DB")
	}

	return db, nil
}

func GetBadgerInstance() *badger.DB {
	once.Do(func() {
		db, err := createDB()
		if err != nil {
			// errors should cause a panic to avoid singleton deadlocks
			panic(err)
		}
		instance = db
	})
	return instance
}
