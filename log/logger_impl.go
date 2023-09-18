package log

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/lo"
)

// Logger is a reactive logger that can be used to log messages with different log levels.
type logger struct {
	name          string
	namespace     string
	rootLogger    *slog.Logger
	level         *slog.LevelVar
	reactiveLevel reactive.Variable[Level]
}

// newLogger creates a new logger with the given namespace and root logger instance.
func newLogger(namespace, name string, rootLogger *slog.Logger) *logger {
	l := &logger{
		name:          name,
		namespace:     lo.Cond(namespace == "", name, namespace+"."+name),
		rootLogger:    rootLogger,
		level:         new(slog.LevelVar),
		reactiveLevel: reactive.NewVariable[Level](),
	}

	l.reactiveLevel.OnUpdate(func(_, newLevel Level) { l.level.Set(newLevel) })

	return l
}

func (l *logger) LogPath() string {
	return l.namespace
}

func (l *logger) LogLevel() Level {
	return l.level.Level()
}

func (l *logger) LogF(fmtString string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, fmt.Sprintf(fmtString, args...))
	}
}

func (l *logger) LogTraceF(fmt string, args ...any) {
	l.LogF(fmt, LevelTrace, args...)
}

func (l *logger) LogDebugF(fmt string, args ...any) {
	l.LogF(fmt, LevelDebug, args...)
}

func (l *logger) LogInfoF(fmt string, args ...any) {
	l.LogF(fmt, LevelInfo, args...)
}

func (l *logger) LogWarnF(fmt string, args ...any) {
	l.LogF(fmt, LevelWarning, args...)
}

func (l *logger) LogErrorF(fmt string, args ...any) {
	l.LogF(fmt, LevelError, args...)
}

func (l *logger) NewEntityLogger(entityName string, shutdownEvent reactive.Event, initLogging func(entityLogger Logger)) Logger {
	if l == nil {
		return l
	}

	embeddedLogger, shutdown := l.NewChildLogger(uniqueEntityName(entityName))
	shutdownEvent.OnTrigger(shutdown)

	initLogging(embeddedLogger)

	return embeddedLogger
}

// LogName returns the name of the logger (the last part of the namespace).
func (l *logger) LogName() string {
	return l.name
}

// Log logs a message with the given log level.
func (l *logger) Log(msg string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.Log(context.Background(), level, msg, append([]interface{}{namespaceKey, l.namespace}, args...)...)
	}
}

// LogAttrs logs a message with the given log level and typed slog attributes.
func (l *logger) LogAttrs(msg string, level Level, args ...slog.Attr) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, msg, append([]slog.Attr{{Key: namespaceKey, Value: slog.StringValue(l.namespace)}}, args...)...)
	}
}

// LogTrace logs a trace message.
func (l *logger) LogTrace(msg string, args ...any) {
	l.Log(msg, LevelTrace, args...)
}

// LogTraceAttrs logs a trace message with typed slog attributes.
func (l *logger) LogTraceAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelTrace, args...)
}

// LogDebug logs a debug message.
func (l *logger) LogDebug(msg string, args ...any) {
	l.Log(msg, LevelDebug, args...)
}

// LogDebugAttrs logs a debug message with typed slog attributes.
func (l *logger) LogDebugAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelDebug, args...)
}

// LogInfo logs an info message.
func (l *logger) LogInfo(msg string, args ...any) {
	l.Log(msg, LevelInfo, args...)
}

// LogInfoAttrs logs an info message with typed slog attributes.
func (l *logger) LogInfoAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelInfo, args...)
}

// LogWarn logs a warning message.
func (l *logger) LogWarn(msg string, args ...any) {
	l.Log(msg, LevelWarning, args...)
}

// LogWarnAttrs logs a warning message with typed slog attributes.
func (l *logger) LogWarnAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelWarning, args...)
}

// LogError logs an error message.
func (l *logger) LogError(msg string, args ...any) {
	l.Log(msg, LevelError, args...)
}

// LogErrorAttrs logs an error message with typed slog attributes.
func (l *logger) LogErrorAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelError, args...)
}

// SetLogLevel sets the log level of the logger.
func (l *logger) SetLogLevel(level Level) {
	if l != nil {
		l.reactiveLevel.Set(level)
	}
}

// OnLogLevelActive registers a callback that is triggered when the given log level is activated. The shutdown
// function that is returned by the callback is automatically called when the log level is deactivated.
func (l *logger) OnLogLevelActive(logLevel Level, setup func() (shutdown func())) (unsubscribe func()) {
	if l == nil {
		return func() {}
	}

	var shutdownEvent reactive.Event

	unsubscribeFromLevel := l.reactiveLevel.OnUpdate(func(_, newLevel Level) {
		if newLevel <= logLevel {
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

	return func() {
		unsubscribeFromLevel()

		if shutdownEvent != nil {
			shutdownEvent.Trigger()
		}
	}
}

// NestedLogger creates a new logger with the given sub-namespace. The new logger inherits the log level from the parent
// logger, but can also be set to its own individual log level.
func (l *logger) NewChildLogger(subNamespace string) (nestedLogger Logger, shutdown func()) {
	if l == nil {
		return (*logger)(nil), func() {}
	}

	nestedLoggerInstance := newLogger(l.namespace, subNamespace, l.rootLogger)

	return nestedLoggerInstance, nestedLoggerInstance.reactiveLevel.InheritFrom(l.reactiveLevel)
}

func (l *logger) String() string {
	return strings.TrimRight(fmt.Sprintf("Logger[%s] (LEVEL = %s", l.namespace, LevelName(l.level.Level())), " ") + ")"
}

// namespaceKey is the key of the slog attribute that holds the namespace of the logger.
const namespaceKey = "namespace"
