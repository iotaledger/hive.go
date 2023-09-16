package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/lo"
)

// Logger is a reactive logger that can be used to log messages with different log levels.
type Logger struct {
	name          string
	namespace     string
	reactiveLevel reactive.Variable[Level]
	rootLogger    *slog.Logger
	level         *slog.LevelVar
}

// New creates a new logger with the given namespace and an optional handler. The default handler prints log records in
// a human-readable format to stdout.
func New(name string, handler ...slog.Handler) *Logger {
	return newLogger("", name, slog.New(lo.First(handler, NewDefaultHandler(os.Stdout))))
}

// newLogger creates a new logger with the given namespace and root logger instance.
func newLogger(namespace, name string, rootLogger *slog.Logger) *Logger {
	l := &Logger{
		reactiveLevel: reactive.NewVariable[Level](),
		name:          name,
		namespace:     lo.Cond(namespace == "", name, namespace+"."+name),
		rootLogger:    rootLogger,
		level:         new(slog.LevelVar),
	}

	l.reactiveLevel.OnUpdate(func(_, newLevel Level) { l.level.Set(newLevel) })

	return l
}

// SetLogLevel sets the log level of the logger.
func (l *Logger) SetLogLevel(level Level) {
	if l != nil {
		l.LogError("log level set to " + LevelName(level))
		l.reactiveLevel.Set(level)
	}
}

// Trace logs a trace message.
func (l *Logger) Trace(msg string, args ...any) {
	l.log(msg, LevelTrace, args...)
}

// TraceAttrs logs a trace message with typed slog attributes.
func (l *Logger) TraceAttrs(msg string, args ...slog.Attr) {
	l.logAttrs(msg, LevelTrace, args...)
}

// OnTraceLevel registers a callback that is called when the log level is set to trace or lower.
func (l *Logger) OnTraceLevel(setup func() (shutdown func())) (unsubscribe func()) {
	return l.onLogLevel(LevelTrace, setup)
}

// LogDebug logs a debug message.
func (l *Logger) LogDebug(msg string, args ...any) {
	l.log(msg, LevelDebug, args...)
}

// DebugAttrs logs a debug message with typed slog attributes.
func (l *Logger) DebugAttrs(msg string, args ...slog.Attr) {
	l.logAttrs(msg, LevelDebug, args...)
}

// OnDebugLevel registers a callback that is called when the log level is set to debug or lower.
func (l *Logger) OnLogLevelDebug(setup func() (shutdown func())) (unsubscribe func()) {
	return l.onLogLevel(LevelDebug, setup)
}

// LogInfo logs an info message.
func (l *Logger) LogInfo(msg string, args ...any) {
	l.log(msg, LevelInfo, args...)
}

// InfoAttrs logs an info message with typed slog attributes.
func (l *Logger) InfoAttrs(msg string, args ...slog.Attr) {
	l.logAttrs(msg, LevelInfo, args...)
}

// OnInfoLevel registers a callback that is called when the log level is set to info or lower.
func (l *Logger) OnLogLevelInfo(setup func() (shutdown func())) (unsubscribe func()) {
	return l.onLogLevel(LevelInfo, setup)
}

// LogWarn logs a warning message.
func (l *Logger) LogWarn(msg string, args ...any) {
	l.log(msg, LevelWarning, args...)
}

// WarnAttrs logs a warning message with typed slog attributes.
func (l *Logger) WarnAttrs(msg string, args ...slog.Attr) {
	l.logAttrs(msg, LevelWarning, args...)
}

// OnWarnLevel registers a callback that is called when the log level is set to warning or lower.
func (l *Logger) OnWarnLevel(setup func() (shutdown func())) (unsubscribe func()) {
	return l.onLogLevel(LevelWarning, setup)
}

// Error logs an error message.
func (l *Logger) LogError(msg string, args ...any) {
	l.log(msg, LevelError, args...)
}

// ErrorAttrs logs an error message with typed slog attributes.
func (l *Logger) ErrorAttrs(msg string, args ...slog.Attr) {
	l.logAttrs(msg, LevelError, args...)
}

// OnErrorLevel registers a callback that is called when the log level is set to error or lower.
func (l *Logger) OnErrorLevel(setup func() (shutdown func())) (unsubscribe func()) {
	return l.onLogLevel(LevelError, setup)
}

// NestedLogger creates a new logger with the given sub-namespace. The new logger inherits the log level from the parent
// logger, but can also be set to its own individual log level.
func (l *Logger) NestedLogger(subNamespace string) (nestedLogger *Logger, shutdown func()) {
	if l == nil {
		return nil, func() {}
	}

	nestedLogger = newLogger(l.namespace, subNamespace, l.rootLogger)
	shutdown = nestedLogger.reactiveLevel.InheritFrom(l.reactiveLevel)

	return nestedLogger, shutdown
}

// log logs a message with the given log level.
func (l *Logger) log(msg string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.Log(context.Background(), level, msg, append([]interface{}{namespaceKey, l.namespace}, args...)...)
	}
}

// logAttrs logs a message with the given log level and typed slog attributes.
func (l *Logger) logAttrs(msg string, level Level, args ...slog.Attr) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, msg, append([]slog.Attr{{Key: namespaceKey, Value: slog.StringValue(l.namespace)}}, args...)...)
	}
}

// onLogLevel registers a callback that is called when the log level is set to the given log level or lower. The
// callback has to return a shutdown function that is called when the log level is set to a higher level.
func (l *Logger) onLogLevel(logLevel Level, setup func() (shutdown func())) (unsubscribe func()) {
	if l == nil {
		return func() {}
	}

	var shutdownEvent reactive.Event

	return l.reactiveLevel.OnUpdate(func(_, newValue Level) {
		if newValue <= logLevel {
			if shutdownEvent == nil {
				shutdownEvent = reactive.NewEvent()
				shutdownEvent.OnTrigger(setup())
			}
		} else {
			if shutdownEvent != nil {
				shutdownEvent.Trigger()
				shutdownEvent = nil
			}
		}
	}, true)
}

const (
	// namespaceKey is the key of the slog attribute that holds the namespace of the logger.
	namespaceKey = "namespace"
)
