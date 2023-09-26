package testutil

import (
	"strconv"
	"sync"
	"testing"

	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/badger"
)

// variables for keeping track of how many databases have been created by the given test.
var databaseCounter = make(map[string]int)
var databaseCounterMutex sync.Mutex

// BadgerDB creates a temporary BadgerKVStore that automatically gets cleaned up when the test finishes.
func BadgerDB(t *testing.T) (kvstore.KVStore, error) {
	dir := t.TempDir()

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

	storeWithRealm, err := badger.New(db).WithRealm([]byte(t.Name() + strconv.Itoa(counter)))
	if err != nil {
		return nil, err
	}

	return storeWithRealm, nil
}
