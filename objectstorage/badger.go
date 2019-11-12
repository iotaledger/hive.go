package objectstorage

import (
	"github.com/pkg/errors"
	"os"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

var instance *badger.DB

var once sync.Once

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
	directory := *DIRECTORY.Value
	if err := checkDir(directory); err != nil {
		return nil, errors.Wrap(err, "Could not check directory")
	}

	opts := badger.DefaultOptions(directory)
	opts.Logger = &disabledBadgerLogger{}
	opts.Truncate = true
	opts.TableLoadingMode = options.MemoryMap

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open new DB")
	}

	return db, nil
}

type disabledBadgerLogger struct{}

func (this *disabledBadgerLogger) Errorf(string, ...interface{}) {
	// disable logging
}

func (this *disabledBadgerLogger) Infof(string, ...interface{}) {
	// disable logging
}

func (this *disabledBadgerLogger) Warningf(string, ...interface{}) {
	// disable logging
}

func (this *disabledBadgerLogger) Debugf(string, ...interface{}) {
	// disable logging
}
