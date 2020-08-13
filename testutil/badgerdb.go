package testutil

import (
	"testing"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/badger"
)

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

	return badger.New(db).WithRealm([]byte(t.Name())), nil
}
