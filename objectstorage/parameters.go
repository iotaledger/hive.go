package objectstorage

import flag "github.com/spf13/pflag"

func init() {
	flag.String("objectstorage.directory", "objectsdb", "path to the database folder")
}
