package bolt

import (
	"fmt"
	"path"

	"go.etcd.io/bbolt"
)

func CreateDB(directory string, filename string, optionalOptions ...*bbolt.Options) (*bbolt.DB, error) {

	if err := checkDir(directory); err != nil {
		return nil, fmt.Errorf("could not check directory: %w", err)
	}

	options := bbolt.DefaultOptions
	if len(optionalOptions) > 0 {
		options = optionalOptions[0]
	}

	db, err := bbolt.Open(path.Join(directory, filename), 0666, options)
	if err != nil {
		return nil, fmt.Errorf("could not open new DB: %w", err)
	}
	return db, nil
}
