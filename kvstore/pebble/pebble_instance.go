package pebble

import (
	"github.com/cockroachdb/pebble"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/ioutils"
)

func CreateDB(directory string, optionalOptions ...*pebble.Options) (*pebble.DB, error) {

	if err := ioutils.CreateDirectory(directory, 0700); err != nil {
		return nil, ierrors.Wrapf(err, "could not create directory '%s'", directory)
	}

	var opts *pebble.Options

	if len(optionalOptions) > 0 {
		opts = optionalOptions[0]
	} else {
		opts = &pebble.Options{}
	}

	db, err := pebble.Open(directory, opts)
	if err != nil {
		return nil, ierrors.Wrapf(err, "could not open new DB '%s'", directory)
	}

	return db, nil
}
