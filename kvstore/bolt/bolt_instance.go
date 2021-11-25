package bolt

import (
	"fmt"
	"path"

	"go.etcd.io/bbolt"

	"github.com/iotaledger/hive.go/v2/kvstore/utils"
)

func CreateDB(directory string, filename string, optionalOptions ...*bbolt.Options) (*bbolt.DB, error) {

	if err := utils.CreateDirectory(directory, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
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
