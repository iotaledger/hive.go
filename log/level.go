package log

import (
	"log/slog"

	"github.com/iotaledger/hive.go/ierrors"
)

// Level is the type of log levels.
type Level = slog.Level

const (
	// LevelTrace is the log level for trace messages.
	LevelTrace = slog.Level(-8)

	// LevelDebug is the log level for debug messages.
	LevelDebug = slog.LevelDebug

	// LevelInfo is the log level for info messages.
	LevelInfo = slog.LevelInfo

	// LevelWarning is the log level for warning messages.
	LevelWarning = slog.LevelWarn

	// LevelError is the log level for error messages.
	LevelError = slog.LevelError

	// LevelFatal is the log level for fatal messages.
	LevelFatal = slog.Level(12)

	// LevelPanic is the log level for panic messages.
	LevelPanic = slog.Level(16)
)

// LevelName returns the name of the given log level.
func LevelName(level Level) string {
	switch level {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelPanic:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// LevelFromString returns the log level for the given string.
func LevelFromString(level string) (Level, error) {
	switch level {
	case "trace", "TRACE":
		return LevelTrace, nil
	case "debug", "DEBUG":
		return LevelDebug, nil
	case "info", "INFO":
		return LevelInfo, nil
	case "warning", "WARNING":
		return LevelWarning, nil
	case "error", "ERROR":
		return LevelError, nil
	case "fatal", "FATAL":
		return LevelFatal, nil
	case "panic", "PANIC":
		return LevelPanic, nil
	default:
		return 0, ierrors.Errorf("unknown log level: %s", level)
	}
}
