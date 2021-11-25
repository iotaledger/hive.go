package testutil

import (
	"strconv"
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/v2/kvstore"
	"github.com/iotaledger/hive.go/v2/kvstore/badger"
)

// variables for keeping track of how many databases have been created by the given test
var databaseCounter = make(map[string]int)
var databaseCounterMutex sync.Mutex

// BadgerDB creates a temporary BadgerKVStore that automatically gets cleaned up when the test finishes.
func BadgerDB(t *testing.T) (kvstore.KVStore, error) {
	dir, err := TempDir(t)
	if err != nil {
		return nil, err
	}

	db, err := badger.CreateDB(dir)
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

	return badger.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter))), nil
}
