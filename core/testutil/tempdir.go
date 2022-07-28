package testutil

import (
	"io/ioutil"
	"os"
	"testing"
)

// TempDir creates a temporary directory, that automatically gets cleaned up when the test finishes.
func TempDir(t *testing.T) (string, error) {
	tempDir, err := ioutil.TempDir("", t.Name())

	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("TempDir RemoveAll cleanup: %v", err)
		}
	})

	return tempDir, nil
}
