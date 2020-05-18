package bolt

import (
	"os"
	
	"github.com/iotaledger/hive.go/kvstore"
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

func buildPrefixedKey(prefixes []kvstore.KeyPrefix) []byte {
	var prefix []byte
	for _, p := range prefixes {
		prefix = append(prefix, p...)
	}
	return prefix
}

func copyBytes(source []byte) []byte {
	cpy := make([]byte, len(source))
	copy(cpy, source)
	return cpy
}
