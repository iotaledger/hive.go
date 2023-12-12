package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/iotaledger/hive.go/runtime/options"
)

// Logger is a reactive logger that can be used to log messages with different log levels.
type Logger interface {
	// LogName returns the name of the logger instance.
	LogName() string

	// LogPath returns the full path of the logger that is formed by a combination of the names of its ancestors and
	// its own name.
	LogPath() string

	// LogLevel returns the current log level of the logger.
	LogLevel() Level

	// SetLogLevel sets the log level of the logger.
	SetLogLevel(level Level)

	// OnLogLevelActive registers a callback that is triggered when the given log level is activated. The shutdown
	// function that is returned by the callback is automatically called when the log level is deactivated.
	OnLogLevelActive(logLevel Level, setup func() (shutdown func())) (unsubscribe func())

	// LogTrace emits a log message with the TRACE level.
	LogTrace(msg string, args ...any)

	// LogTracef emits a formatted log message with the TRACE level.
	LogTracef(fmtString string, args ...any)

	// LogTraceAttrs emits a log message with the TRACE level and the given attributes.
	LogTraceAttrs(msg string, args ...slog.Attr)

	// LogDebug emits a log message with the DEBUG level.
	LogDebug(msg string, args ...any)

	// LogDebugf emits a formatted log message with the DEBUG level.
	LogDebugf(fmtString string, args ...any)

	// LogDebugAttrs emits a log message with the DEBUG level and the given attributes.
	LogDebugAttrs(msg string, args ...slog.Attr)

	// LogInfo emits a log message with the INFO level.
	LogInfo(msg string, args ...any)

	// LogInfof emits a formatted log message with the INFO level.
	LogInfof(fmtString string, args ...any)

	// LogInfoAttrs emits a log message with the INFO level and the given attributes.
	LogInfoAttrs(msg string, args ...slog.Attr)

	// LogWarn emits a log message with the WARN level.
	LogWarn(msg string, args ...any)

	// LogWarnf emits a formatted log message with the WARN level.
	LogWarnf(fmtString string, args ...any)

	// LogWarnAttrs emits a log message with the WARN level and the given attributes.
	LogWarnAttrs(msg string, args ...slog.Attr)

	// LogError emits a log message with the ERROR level.
	LogError(msg string, args ...any)

	// LogErrorf emits a formatted log message with the ERROR level.
	LogErrorf(fmtString string, args ...any)

	// LogErrorAttrs emits a log message with the ERROR level and the given attributes.
	LogErrorAttrs(msg string, args ...slog.Attr)

	// LogFatal emits a log message with the FATAL level, then calls os.Exit(1).
	LogFatal(msg string, args ...any)

	// LogFatalf emits a formatted log message with the FATAL level, then calls os.Exit(1).
	LogFatalf(fmtString string, args ...any)

	// LogFatalAttrs emits a log message with the FATAL level and the given attributes, then calls os.Exit(1).
	LogFatalAttrs(fmtString string, args ...slog.Attr)

	// LogPanic emits a log message with the PANIC level, then panics.
	LogPanic(msg string, args ...any)

	// LogPanicf emits a formatted log message with the PANIC level, then panics.
	LogPanicf(fmtString string, args ...any)

	// LogPanicAttrs emits a log message with the PANIC level and the given attributes, then panics.
	LogPanicAttrs(fmtString string, args ...slog.Attr)

	// Log emits a log message with the given level.
	Log(msg string, level Level, args ...any)

	// Logf emits a formatted log message with the given level.
	Logf(fmtString string, level Level, args ...any)

	// LogAttrs emits a log message with the given level and attributes.
	LogAttrs(msg string, level Level, args ...slog.Attr)

	// NewChildLogger creates a new child logger with the given name.
	NewChildLogger(name string) (childLogger Logger, shutdown func())

	// NewEntityLogger creates a new logger for an entity with the given name. The logger is automatically shut down
	// when the given shutdown event is triggered. The initLogging function is called with the new logger instance and
	// can be used to configure the logger.
	NewEntityLogger(entityName string) (entityLogger Logger, shutdown func())
}

type Options struct {
	// Name is the name of the logger.
	Name string
	// Level is the log level of the logger.
	Level Level
	// TimeFormat is the time format of the logger.
	TimeFormat string
	// Output is the output of the logger.
	Output io.Writer
}

func WithName(name string) options.Option[Options] {
	return func(opts *Options) {
		opts.Name = name
	}
}

func WithLevel(level Level) options.Option[Options] {
	return func(opts *Options) {
		opts.Level = level
	}
}

func WithTimeFormat(timeFormat string) options.Option[Options] {
	return func(opts *Options) {
		opts.TimeFormat = timeFormat
	}
}

func WithOutput(output io.Writer) options.Option[Options] {
	return func(opts *Options) {
		opts.Output = output
	}
}

// NewLogger creates a new logger with the given options.
// If no options are provided, the logger uses the info level and writes to stdout with rfc3339 time format.
func NewLogger(opts ...options.Option[Options]) Logger {
	loggerOptions := options.Apply(&Options{
		Name:       "",
		Level:      LevelInfo,
		TimeFormat: "rfc3339",
		Output:     os.Stdout,
	}, opts)

	logger := newLogger(slog.New(NewTextHandler(loggerOptions)), "", loggerOptions.Name)
	logger.SetLogLevel(loggerOptions.Level)

	return logger
}

// EmptyLogger is a logger that does not log anything.
var EmptyLogger Logger = (*logger)(nil)
