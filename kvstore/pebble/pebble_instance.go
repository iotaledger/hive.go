package pebble

import (
	"fmt"

	"github.com/cockroachdb/pebble"
)

func CreateDB(directory string, optionalOptions ...*pebble.Options) (*pebble.DB, error) {
	if err := checkDir(directory); err != nil {
		return nil, fmt.Errorf("could not check directory: %w", err)
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
