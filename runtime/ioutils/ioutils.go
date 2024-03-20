package ioutils

import (
	"encoding/binary"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/iotaledger/hive.go/ierrors"
)

// PathExists returns whether the given file or directory exists.
func PathExists(path string) (exists bool, isDirectory bool, err error) {
	fileInfo, err := os.Stat(path)
	if err == nil {
		return true, fileInfo.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, false, nil
	}

	return false, false, err
}

// CreateDirectory checks if the directory exists,
// otherwise it creates it with given permissions.
func CreateDirectory(dir string, perm os.FileMode) error {
	exists, isDir, err := PathExists(dir)
	if err != nil {
		return err
	}

	if exists {
		if !isDir {
			return ierrors.Errorf("given path is a file instead of a directory %s", dir)
		}

		return nil
	}

	return os.MkdirAll(dir, perm)
}

// FolderSize returns the size of a folder.
func FolderSize(target string) (int64, error) {
	var size int64

	err := filepath.Walk(target, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	return size, err
}

// ReadFromFile reads structured binary data from the file named by filename to data.
// A successful call returns err == nil, not err == EOF.
// ReadFromFile uses binary.Read to decode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func ReadFromFile(filename string, data interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return binary.Read(f, binary.LittleEndian, data)
}

// WriteToFile writes the binary representation of data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm
// (before umask); otherwise WriteFile truncates it before writing, without changing permissions.
// WriteToFile uses binary.Write to encode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func WriteToFile(filename string, data interface{}, perm os.FileMode) (err error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	return binary.Write(f, binary.LittleEndian, data)
}

// ReadJSONFromFile reads JSON data from the file named by filename to data.
// ReadJSONFromFile uses json.Unmarshal to decode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func ReadJSONFromFile(filename string, data interface{}) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, data)
}

// WriteJSONToFile writes the JSON representation of data to a file named by filename.
// If the file does not exist, WriteJSONToFile creates it with permissions perm
// (before umask); otherwise WriteJSONToFile truncates it before writing, without changing permissions.
// WriteJSONToFile uses json.MarshalIndent to encode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func WriteJSONToFile(filename string, data interface{}, perm os.FileMode) (err error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ierrors.Wrap(err, "unable to marshal data to JSON")
	}

	if _, err := f.Write(jsonData); err != nil {
		return ierrors.Wrapf(err, "unable to write JSON data to %s", filename)
	}

	if err := f.Sync(); err != nil {
		return ierrors.Wrapf(err, "unable to fsync file content to %s", filename)
	}

	return nil
}

// ReadYAMLFromFile reads YAML data from the file named by filename to data.
// ReadYAMLFromFile uses yaml.Unmarshal to decode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func ReadYAMLFromFile(filename string, data interface{}) error {
	yamlData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlData, data)
}

// WriteYAMLToFile writes the YAML representation of data to a file named by filename.
// If the file does not exist, WriteYAMLToFile creates it with permissions perm
// (before umask); otherwise WriteYAMLToFile truncates it before writing, without changing permissions.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func WriteYAMLToFile(filename string, data interface{}, perm os.FileMode, indent int) (err error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(indent)

	if err := encoder.Encode(data); err != nil {
		return ierrors.Wrapf(err, "unable to marshal YAML data to %s", filename)
	}

	if err := encoder.Close(); err != nil {
		return ierrors.Wrapf(err, "unable to close YAML encoder for %s", filename)
	}

	if err := f.Sync(); err != nil {
		return ierrors.Wrapf(err, "unable to fsync file content to %s", filename)
	}

	return nil
}

// ReadTOMLFromFile reads TOML data from the file named by filename to data.
// ReadTOMLFromFile uses toml.Unmarshal to decode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func ReadTOMLFromFile(filename string, data interface{}) error {
	tomlData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return toml.Unmarshal(tomlData, data)
}

// WriteTOMLToFile writes the TOML representation of data to a file named by filename.
// If the file does not exist, WriteTOMLToFile creates it with permissions perm
// (before umask); otherwise WriteTOMLToFile truncates it before writing, without changing permissions.
// WriteTOMLToFile uses toml.Marshal to encode data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values. An additional header can be passed.
func WriteTOMLToFile(filename string, data interface{}, perm os.FileMode, header ...string) (err error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	tomlData, err := toml.Marshal(data)
	if err != nil {
		return ierrors.Wrap(err, "unable to marshal data to TOML")
	}

	if len(header) > 0 {
		if _, err := f.WriteString(header[0] + "\n"); err != nil {
			return ierrors.Wrapf(err, "unable to write header to %s", filename)
		}
	}

	if _, err := f.Write(tomlData); err != nil {
		return ierrors.Wrapf(err, "unable to write TOML data to %s", filename)
	}

	if err := f.Sync(); err != nil {
		return ierrors.Wrapf(err, "unable to fsync file content to %s", filename)
	}

	return nil
}

// CreateTempFile creates a file descriptor with _tmp as file extension.
func CreateTempFile(filePath string) (*os.File, string, error) {
	filePathTmp := filePath + "_tmp"

	// we don't need to check the error, maybe the file doesn't exist
	_ = os.Remove(filePathTmp)

	fileDescriptor, err := os.OpenFile(filePathTmp, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, "", err
	}

	return fileDescriptor, filePathTmp, nil
}

// CloseFileAndRename closes the file descriptor and renames the file.
func CloseFileAndRename(fileDescriptor *os.File, sourceFilePath string, targetFilePath string) error {
	if err := fileDescriptor.Close(); err != nil {
		return ierrors.Wrap(err, "unable to close file")
	}
	if err := os.Rename(sourceFilePath, targetFilePath); err != nil {
		return ierrors.Wrap(err, "unable to rename file")
	}

	return nil
}

// DirectoryEmpty returns whether the given directory is empty.
func DirectoryEmpty(dirPath string) (bool, error) {
	// check if the directory exists
	if _, err := os.Stat(dirPath); err != nil {
		return false, ierrors.Wrapf(err, "unable to check directory (%s)", dirPath)
	}

	// check if the directory is empty
	if err := filepath.WalkDir(dirPath, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirPath == path {
			// skip the root folder itself
			return nil
		}

		return os.ErrExist
	}); err != nil {
		if !os.IsExist(err) {
			return false, ierrors.Wrapf(err, "unable to check directory (%s)", dirPath)
		}

		// directory is not empty
		return false, nil
	}

	// directory is empty
	return true, nil
}

// DirExistsAndIsNotEmpty checks if the folder exists and is not empty.
func DirExistsAndIsNotEmpty(path string) (bool, error) {
	dirExists, isDir, err := PathExists(path)
	if err != nil {
		return false, ierrors.Wrapf(err, "unable to check dir path (%s)", path)
	}
	if !dirExists {
		return false, nil
	}
	if !isDir {
		return false, ierrors.Errorf("given path is a file instead of a directory %s", path)
	}

	// check if the directory is empty (needed for example in docker environments)
	dirEmpty, err := DirectoryEmpty(path)
	if err != nil {
		return false, ierrors.Wrapf(err, "unable to check dir (%s)", path)
	}

	return !dirEmpty, nil
}
