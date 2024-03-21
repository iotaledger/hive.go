package testutil

import (
	"strconv"
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
)

// variables for keeping track of how many databases have been created by the given test.
var databaseCounter = make(map[string]int)
var databaseCounterMutex sync.Mutex

// RocksDB creates a temporary RocksDBKVStore that automatically gets cleaned up when the test finishes.
func RocksDB(t *testing.T) (kvstore.KVStore, error) {
	t.Helper()

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
