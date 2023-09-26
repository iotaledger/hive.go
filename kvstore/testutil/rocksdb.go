package testutil

import (
	"strconv"
	"testing"

	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/rocksdb"
)

// RocksDB creates a temporary RocksDBKVStore that automatically gets cleaned up when the test finishes.
func RocksDB(t *testing.T) (kvstore.KVStore, error) {
	dir := t.TempDir()

	db, err := rocksdb.CreateDB(dir)
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

	storeWithRealm, err := rocksdb.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter)))
	if err != nil {
		return nil, err
	}

	return storeWithRealm, nil
}
