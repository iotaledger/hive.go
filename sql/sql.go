package sql

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/options"
)

type DatabaseParameters struct {
	Engine db.Engine

	// SQLite
	Path     string
	Filename string

	// PostgreSQL
	Host     string
	Port     uint
	Database string
	Username string
	Password string
}

type gormDatabaseOptions struct {
	gormConfig       *gorm.Config
	gormLoggerConfig gormLogger.Config
}

// WithGormConfig allows to set the gorm config.
// HINT: The Logger setting will be overwritten by the internal default value or by the value set by WithGormLoggerConfig.
func WithGormConfig(config *gorm.Config) options.Option[gormDatabaseOptions] {
	return func(o *gormDatabaseOptions) {
		o.gormConfig = config
	}
}

// WithGormLoggerConfig allows to set the gorm logger config.
func WithGormLoggerConfig(config gormLogger.Config) options.Option[gormDatabaseOptions] {
	return func(o *gormDatabaseOptions) {
		o.gormLoggerConfig = config
	}
}

// New creates a new gorm database instance with the given parameters and options.
func New(logger log.Logger, dbParams DatabaseParameters, createDatabaseIfNotExists bool, allowedEngines []db.Engine, opts ...options.Option[gormDatabaseOptions]) (*gorm.DB, db.Engine, error) {
	targetEngine, err := db.CheckEngine(dbParams.Path, createDatabaseIfNotExists, dbParams.Engine, allowedEngines)
	if err != nil {
		return nil, db.EngineUnknown, err
	}

	var dbDialector gorm.Dialector

	//nolint:exhaustive // false positive
	switch targetEngine {
	case db.EngineSQLite, db.EngineAuto:
		dbDialector = sqlite.Open(fmt.Sprintf("file:%s?&_journal_mode=WAL&_busy_timeout=60000", filepath.Join(dbParams.Path, dbParams.Filename)))
	case db.EnginePostgreSQL:
		dsn := fmt.Sprintf("host='%s' user='%s' password='%s' dbname='%s' port=%d", dbParams.Host, dbParams.Username, dbParams.Password, dbParams.Database, dbParams.Port)
		dbDialector = postgres.Open(dsn)
	default:
		return nil, db.EngineUnknown, ierrors.Errorf("unknown database engine: %s, supported engines: %s", dbParams.Engine, db.GetSupportedEnginesString(allowedEngines))
	}

	gormDBOptions := options.Apply(&gormDatabaseOptions{
		gormConfig: &gorm.Config{},
		gormLoggerConfig: gormLogger.Config{
			SlowThreshold:             100 * time.Millisecond,
			LogLevel:                  gormLogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	}, opts,
		func(o *gormDatabaseOptions) {
			// overwrite the logger in the gorm config to initialize the logger with the given hive logger from the outside and the given settings.
			o.gormConfig.Logger = gormLogger.New(newLogger(logger), o.gormLoggerConfig)
		},
	)

	database, err := gorm.Open(dbDialector, gormDBOptions.gormConfig)
	if err != nil {
		return nil, db.EngineUnknown, err
	}

	return database, targetEngine, nil
}
