package testutil

import (
	"strconv"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"

	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/pebble"
)

// PebbleDB creates a temporary PebbleKVStore that automatically gets cleaned up when the test finishes.
func PebbleDB(t *testing.T) (kvstore.KVStore, error) {
	dir := t.TempDir()

	cache := pebbledb.NewCache(1 << 30)
	defer cache.Unref()

	opts := &pebbledb.Options{
		Cache:                       cache,
		DisableWAL:                  false,
		L0CompactionThreshold:       2,
		L0StopWritesThreshold:       1000,
		LBaseMaxBytes:               64 << 20, // 64 MB
		Levels:                      make([]pebbledb.LevelOptions, 7),
		MaxConcurrentCompactions:    func() int { return 3 },
		MaxOpenFiles:                16384,
		MemTableSize:                64 << 20,
		MemTableStopWritesThreshold: 4,
	}

	for i := 0; i < len(opts.Levels); i++ {
		l := &opts.Levels[i]
		l.BlockSize = 32 << 10       // 32 KB
		l.IndexBlockSize = 256 << 10 // 256 KB
		l.FilterPolicy = bloom.FilterPolicy(10)
		l.FilterType = pebbledb.TableFilter
		if i > 0 {
			l.TargetFileSize = opts.Levels[i-1].TargetFileSize * 2
		}
		l.EnsureDefaults()
	}
	opts.Levels[6].FilterPolicy = nil

	opts.EnsureDefaults()

	db, err := pebble.CreateDB(dir, opts)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Closing database: %v", err)
		}
	})

	databaseCounterMutex.Lock()
	databaseCounter[t.Name()]++
	counter := databaseCounter[t.Name()]
	databaseCounterMutex.Unlock()

	storeWithRealm, err := pebble.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter)))
	if err != nil {
		return nil, err
	}

	return storeWithRealm, nil
}
