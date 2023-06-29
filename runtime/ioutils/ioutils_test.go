package ioutils_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/runtime/ioutils"
)

func TestPathExists(t *testing.T) {
	tmpDirPath := t.TempDir()
	// Test existing directory
	exists, isDir, err := ioutils.PathExists(tmpDirPath)
	require.NoError(t, err, "PathExists returned an error: %v", err)
	require.True(t, exists)
	require.True(t, isDir)

	// Test non-existing file
	exists, isDir, err = ioutils.PathExists(tmpDirPath + "/nonexistent.txt")
	require.NoError(t, err, "PathExists returned an error: %v", err)
	require.False(t, exists)
	require.False(t, isDir)
}

func TestCreateDirectory(t *testing.T) {
	tmpDirPath := t.TempDir()
	empty, err := ioutils.DirectoryEmpty(tmpDirPath)
	require.NoError(t, err, "DirectoryEmpty returned an error: %v", err)
	require.True(t, empty)

	// Test creating a new directory
	dir := tmpDirPath + "/newdir"
	err = ioutils.CreateDirectory(dir, os.ModePerm)
	require.NoError(t, err, "CreateDirectory returned an error: %v", err)

	notEmpty, err := ioutils.DirExistsAndIsNotEmpty(tmpDirPath)
	require.NoError(t, err, "DirExistsAndIsNotEmpty returned an error: %v", err)
	require.True(t, notEmpty)

	// Test creating an existing directory
	err = ioutils.CreateDirectory(dir, os.ModePerm)
	require.NoError(t, err, "CreateDirectory returned an error: %v", err)

	// Test creating an file
	fd, oldFileName, err := ioutils.CreateTempFile(dir + "/tmp")
	require.NoError(t, err, "CreateTempFile returned an error: %v", err)

	// Test creating a directory with existing file name
	err = ioutils.CreateDirectory(oldFileName, os.ModePerm)
	require.Error(t, err, "CreateDirectory returned an error: %v", err)

	// Test rename file
	newFileName := dir + "/abc"
	err = ioutils.CloseFileAndRename(fd, oldFileName, newFileName)
	require.NoError(t, err, "CloseFileAndRename returned an error: %v", err)

	exists, isDir, err := ioutils.PathExists(oldFileName)
	require.NoError(t, err, "PathExists returned an error: %v", err)
	require.False(t, exists)
	require.False(t, isDir)

	exists, isDir, err = ioutils.PathExists(newFileName)
	require.NoError(t, err, "PathExists returned an error: %v", err)
	require.True(t, exists)
	require.False(t, isDir)
}

func TestWriteJSONToFile(t *testing.T) {
	type tmpStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}
	// Test data
	data := &tmpStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "johndoe@example.com",
	}

	tmpDirPath := t.TempDir()
	fileName := tmpDirPath + "/tmp"

	// Write JSON to file
	err := ioutils.WriteJSONToFile(fileName, &data, os.ModePerm)
	require.NoError(t, err, "WriteJSONToFile returned an error: %v", err)

	// Read the written file
	readData := &tmpStruct{}
	err = ioutils.ReadJSONFromFile(fileName, readData)
	require.NoError(t, err, "ReadJSONFromFile returned an error: %v", err)
	require.EqualValues(t, data, readData)

	readDataWrong := [2]tmpStruct{}
	err = ioutils.ReadJSONFromFile(fileName, readDataWrong)
	require.Error(t, err, "ReadJSONFromFile should returned an error, wrong structure is passed")
}
