package pebble

import (
	"fmt"

	"github.com/cockroachdb/pebble"

	"github.com/iotaledger/hive.go/v2/kvstore/utils"
)

func CreateDB(directory string, optionalOptions ...*pebble.Options) (*pebble.DB, error) {

	if err := utils.CreateDirectory(directory, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
	}

	var opts *pebble.Options

	if len(optionalOptions) > 0 {
		opts = optionalOptions[0]
	} else {
		opts = &pebble.Options{}
	}

	db, err := pebble.Open(directory, opts)
	if err != nil {
		return nil, fmt.Errorf("could not open new DB: %w", err)
	}

	return db, nil
}
