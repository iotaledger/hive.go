package testutil

import (
	"strconv"
	"testing"

	"github.com/iotaledger/hive.go/v2/kvstore"
	"github.com/iotaledger/hive.go/v2/kvstore/rocksdb"
)

// RocksDB creates a temporary RocksDBKVStore that automatically gets cleaned up when the test finishes.
func RocksDB(t *testing.T) (kvstore.KVStore, error) {
	dir, err := TempDir(t)
	if err != nil {
		return nil, err
	}

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

	return rocksdb.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter))), nil
}
