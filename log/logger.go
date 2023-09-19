package log

import (
	"log/slog"
	"os"

	"github.com/iotaledger/hive.go/ds/reactive"
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

	// LogTraceF emits a formatted log message with the TRACE level.
	LogTraceF(fmtString string, args ...any)

	// LogTraceAttrs emits a log message with the TRACE level and the given attributes.
	LogTraceAttrs(msg string, args ...slog.Attr)

	// LogDebug emits a log message with the DEBUG level.
	LogDebug(msg string, args ...any)

	// LogDebugF emits a formatted log message with the DEBUG level.
	LogDebugF(fmtString string, args ...any)

	// LogDebugAttrs emits a log message with the DEBUG level and the given attributes.
	LogDebugAttrs(msg string, args ...slog.Attr)

	// LogInfo emits a log message with the INFO level.
	LogInfo(msg string, args ...any)

	// LogInfoF emits a formatted log message with the INFO level.
	LogInfoF(fmtString string, args ...any)

	// LogInfoAttrs emits a log message with the INFO level and the given attributes.
	LogInfoAttrs(msg string, args ...slog.Attr)

	// LogWarn emits a log message with the WARN level.
	LogWarn(msg string, args ...any)

	// LogWarnF emits a formatted log message with the WARN level.
	LogWarnF(fmtString string, args ...any)

	// LogWarnAttrs emits a log message with the WARN level and the given attributes.
	LogWarnAttrs(msg string, args ...slog.Attr)

	// LogError emits a log message with the ERROR level.
	LogError(msg string, args ...any)

	// LogErrorF emits a formatted log message with the ERROR level.
	LogErrorF(fmtString string, args ...any)

	// LogErrorAttrs emits a log message with the ERROR level and the given attributes.
	LogErrorAttrs(msg string, args ...slog.Attr)

	// Log emits a log message with the given level.
	Log(msg string, level Level, args ...any)

	// LogF emits a formatted log message with the given level.
	LogF(fmtString string, level Level, args ...any)

	// LogAttrs emits a log message with the given level and attributes.
	LogAttrs(msg string, level Level, args ...slog.Attr)

	// NewChildLogger creates a new child logger with the given name.
	NewChildLogger(name string) (childLogger Logger, shutdown func())

	// NewEntityLogger creates a new logger for an entity with the given name. The logger is automatically shut down
	// when the given shutdown event is triggered. The initLogging function is called with the new logger instance and
	// can be used to configure the logger.
	NewEntityLogger(entityName string, shutdownEvent reactive.Event, initLogging func(entityLogger Logger)) Logger
}

// NewLogger creates a new logger with the given name and an optional handler. If no handler is provided, the logger
// uses the built-in text handler that writes to stdout.
func NewLogger(name string, handler ...slog.Handler) Logger {
	if len(handler) == 0 {
		handler = append(handler, NewTextHandler(os.Stdout))
	}

	return newLogger(slog.New(handler[0]), "", name)
}

// EmptyLogger is a logger that does not log anything.
var EmptyLogger Logger = (*logger)(nil)
