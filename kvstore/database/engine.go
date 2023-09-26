package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/runtime/ioutils"
)

type Engine string

const (
	EngineUnknown Engine = "unknown"
	EngineAuto    Engine = "auto"
	EngineBadger  Engine = "badger"
	EngineDebug   Engine = "debug"
	EngineMapDB   Engine = "mapdb"
	EnginePebble  Engine = "pebble"
	EngineRocksDB Engine = "rocksdb"
)

var (
	ErrEngineMismatch = errors.New("database engine mismatch")
)

type databaseInfo struct {
	Engine string `toml:"databaseEngine"`
}

// engineFromString parses an engine from a string.
func engineFromString(engineStr string) Engine {
	if engineStr == "" {
		// no engine specified
		return EngineAuto
	}

	return Engine(strings.ToLower(engineStr))
}

// getSupportedEnginesString returns a string containing all supported engines separated by "/".
func getSupportedEnginesString(supportedEngines []Engine) string {
	supportedEnginesStr := ""
	for i, allowedEngine := range supportedEngines {
		if i != 0 {
			supportedEnginesStr += "/"
		}
		supportedEnginesStr += string(allowedEngine)
	}

	return supportedEnginesStr
}

// EngineAllowed checks if the database engine is allowed.
func EngineAllowed(dbEngine Engine, allowedEngines []Engine) (Engine, error) {

	for _, allowedEngine := range allowedEngines {
		if dbEngine == allowedEngine {
			return dbEngine, nil
		}
	}

	return EngineUnknown, fmt.Errorf("unknown database engine: %s, supported engines: %s", dbEngine, getSupportedEnginesString(allowedEngines))
}

// EngineFromStringAllowed parses an engine from a string and checks if the database engine is allowed.
func EngineFromStringAllowed(dbEngineStr string, allowedEngines []Engine) (Engine, error) {
	return EngineAllowed(engineFromString(dbEngineStr), allowedEngines)
}

// CheckEngine checks if the correct database engine is used.
// This function stores a so called "database info file" in the database folder or
// checks if an existing "database info file" contains the correct engine.
// Otherwise the files in the database folder are not compatible.
func CheckEngine(dbPath string, createDatabaseIfNotExists bool, dbEngine Engine, allowedEngines []Engine) (Engine, error) {

	// check if the given target engine is allowed
	_, err := EngineAllowed(dbEngine, allowedEngines)
	if err != nil {
		return EngineUnknown, err
	}

	switch dbEngine {
	case EngineUnknown:
		return EngineUnknown, errors.New("the database engine must not be EngineUnknown")

	case EngineMapDB:
		// no need to create or access a "database info file" in case of mapdb (in-memory)
		return EngineMapDB, nil
	}

	dbEngineSpecified := dbEngine != EngineAuto

	// check if the database exists and if it should be created
	dbExists, err := ioutils.DirExistsAndIsNotEmpty(dbPath)
	if err != nil {
		return EngineUnknown, err
	}

	if !dbExists {
		if !createDatabaseIfNotExists {
			return EngineUnknown, fmt.Errorf("database not found (%s)", dbPath)
		}

		if createDatabaseIfNotExists && !dbEngineSpecified {
			return EngineUnknown, errors.New("the database engine must be specified if the database should be newly created")
		}
	}

	var targetEngine Engine

	// check if the database info file exists and if it should be created
	dbInfoFilePath := filepath.Join(dbPath, "dbinfo")
	_, err = os.Stat(dbInfoFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return EngineUnknown, fmt.Errorf("unable to check database info file (%s): %w", dbInfoFilePath, err)
		}

		if !dbEngineSpecified {
			return EngineUnknown, fmt.Errorf("database info file not found (%s)", dbInfoFilePath)
		}

		// if the dbInfo file does not exist and the dbEngine is given, create the dbInfo file.
		if err := storeDatabaseInfoToFile(dbInfoFilePath, dbEngine); err != nil {
			return EngineUnknown, err
		}

		targetEngine = dbEngine
	} else {
		dbEngineFromInfoFile, err := LoadEngineFromFile(dbInfoFilePath, allowedEngines)
		if err != nil {
			return EngineUnknown, err
		}

		// if the dbInfo file exists and the dbEngine is given, compare the engines.
		if dbEngineSpecified && dbEngineFromInfoFile != dbEngine {
			return dbEngineFromInfoFile, ErrEngineMismatch
		}

		targetEngine = dbEngineFromInfoFile
	}

	return targetEngine, nil
}

// LoadEngineFromFile returns the engine from the "database info file".
func LoadEngineFromFile(path string, allowedEngines []Engine) (Engine, error) {
	var info databaseInfo

	if err := ioutils.ReadTOMLFromFile(path, &info); err != nil {
		return "", fmt.Errorf("unable to read database info file: %w", err)
	}

	return EngineFromStringAllowed(info.Engine, allowedEngines)
}

// storeDatabaseInfoToFile stores the used engine in a "database info file".
func storeDatabaseInfoToFile(filePath string, engine Engine) error {
	dirPath := filepath.Dir(filePath)

	if err := ioutils.CreateDirectory(dirPath, 0700); err != nil {
		return fmt.Errorf("could not create database dir '%s': %w", dirPath, err)
	}

	info := &databaseInfo{
		Engine: string(engine),
	}

	return ioutils.WriteTOMLToFile(filePath, info, 0660, "# auto-generated\n# !!! do not modify this file !!!")
}
