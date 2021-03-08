package badger

import (
	"fmt"
	"runtime"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
)

func CreateDB(directory string, optionalOptions ...badger.Options) (*badger.DB, error) {
	if err := checkDir(directory); err != nil {
		return nil, fmt.Errorf("could not check directory: %w", err)
	}

	var opts badger.Options

	if len(optionalOptions) > 0 {
		opts = optionalOptions[0]
	} else {
		opts = badger.DefaultOptions(directory)
		opts.Logger = nil

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

		if runtime.GOOS == "windows" {
			opts = opts.WithTruncate(true)
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("could not open new DB: %w", err)
	}

	return db, nil
}
