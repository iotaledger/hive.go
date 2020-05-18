package bolt

import (
	"path"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

func CreateDB(directory string, filename string, optionalOptions ...*bbolt.Options) (*bbolt.DB, error) {

	if err := checkDir(directory); err != nil {
		return nil, errors.Wrap(err, "Could not check directory")
	}

	options := bbolt.DefaultOptions
	if len(optionalOptions) > 0 {
		options = optionalOptions[0]
	}

	db, err := bbolt.Open(path.Join(directory, filename), 0666, options)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open new DB")
	}
	return db, nil
}
