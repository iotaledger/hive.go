package bolt

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// Returns whether the given file or directory exists.
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
