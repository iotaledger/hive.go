package logger

import "log/slog"

// Level is the type of log levels.
type Level = slog.Level

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
	default:
		return "LOGGER"
	}
}

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
)
