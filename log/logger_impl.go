package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds/reactive"
)

// logger is the default implementation of the Logger interface.
type logger struct {
	// name is the name of the logger instance.
	name string

	// path is the full path of the logger that is formed by a combination of the names of its ancestors and its own
	// name.
	path string

	// rootLogger is the root logger instance.
	rootLogger *slog.Logger

	// parentLogger is the parent logger instance.
	parentLogger *logger

	// unsubscribeFromParent is the function that is used to unsubscribe from the parent logger.
	unsubscribeFromParent func()

	// level is the current log level of the logger.
	level *slog.LevelVar

	// reactiveLevel is the reactive variable that is used to make changes to the log level reactive.
	reactiveLevel reactive.Variable[Level]

	// entityNameCounters holds the instance counters for entity loggers.
	entityNameCounters sync.Map
}

// newLogger creates a new logger instance with the given name and parent logger.
func newLogger(rootLogger *slog.Logger, parentLogger *logger, name string) *logger {
	l := &logger{
		name:          name,
		path:          name,
		rootLogger:    rootLogger,
		parentLogger:  parentLogger,
		level:         new(slog.LevelVar),
		reactiveLevel: reactive.NewVariable[Level](),
	}

	if parentLogger != nil {
		if parentLogger.path != "" {
			l.path = parentLogger.path + "." + l.path
		}

		l.unsubscribeFromParent = l.reactiveLevel.InheritFrom(parentLogger.reactiveLevel)
	}

	l.reactiveLevel.OnUpdate(func(_, newLevel Level) { l.level.Set(newLevel) })

	return l
}

// LogName returns the name of the logger instance.
func (l *logger) LogName() string {
	if l == nil {
		return "<nil>"
	}

	return l.name
}

// LogPath returns the full path of the logger that is formed by a combination of the names of its ancestors and its own
// name.
func (l *logger) LogPath() string {
	if l == nil {
		return "<nil>"
	}

	return l.path
}

// LogLevel returns the current log level of the logger.
func (l *logger) LogLevel() Level {
	if l == nil {
		return LevelInfo
	}

	return l.level.Level()
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

// LogTrace emits a log message with the TRACE level.
func (l *logger) LogTrace(msg string, args ...any) {
	l.Log(msg, LevelTrace, args...)
}

// LogTracef emits a formatted log message with the TRACE level.
func (l *logger) LogTracef(fmtString string, args ...any) {
	l.Logf(fmtString, LevelTrace, args...)
}

// LogTraceAttrs emits a log message with the TRACE level and the given attributes.
func (l *logger) LogTraceAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelTrace, args...)
}

// LogDebug emits a log message with the DEBUG level.
func (l *logger) LogDebug(msg string, args ...any) {
	l.Log(msg, LevelDebug, args...)
}

// LogDebugf emits a formatted log message with the DEBUG level.
func (l *logger) LogDebugf(fmtString string, args ...any) {
	l.Logf(fmtString, LevelDebug, args...)
}

// LogDebugAttrs emits a log message with the DEBUG level and the given attributes.
func (l *logger) LogDebugAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelDebug, args...)
}

// LogInfo emits a log message with the INFO level.
func (l *logger) LogInfo(msg string, args ...any) {
	l.Log(msg, LevelInfo, args...)
}

// LogInfof emits a formatted log message with the INFO level.
func (l *logger) LogInfof(fmtString string, args ...any) {
	l.Logf(fmtString, LevelInfo, args...)
}

// LogInfoAttrs emits a log message with the INFO level and the given attributes.
func (l *logger) LogInfoAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelInfo, args...)
}

// LogWarn emits a log message with the WARN level.
func (l *logger) LogWarn(msg string, args ...any) {
	l.Log(msg, LevelWarning, args...)
}

// LogWarnf emits a formatted log message with the WARN level.
func (l *logger) LogWarnf(fmtString string, args ...any) {
	l.Logf(fmtString, LevelWarning, args...)
}

// LogWarnAttrs emits a log message with the WARN level and the given attributes.
func (l *logger) LogWarnAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelWarning, args...)
}

// LogError emits a log message with the ERROR level.
func (l *logger) LogError(msg string, args ...any) {
	l.Log(msg, LevelError, args...)
}

// LogErrorf emits a formatted log message with the ERROR level.
func (l *logger) LogErrorf(fmtString string, args ...any) {
	l.Logf(fmtString, LevelError, args...)
}

// LogErrorAttrs emits a log message with the ERROR level and the given attributes.
func (l *logger) LogErrorAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelError, args...)
}

// LogFatal emits a log message with the FATAL level, then calls os.Exit(1).
func (l *logger) LogFatal(msg string, args ...any) {
	l.Log(msg, LevelFatal, args...)
	os.Exit(1)
}

// LogFatalf emits a formatted log message with the FATAL level, then calls os.Exit(1).
func (l *logger) LogFatalf(fmtString string, args ...any) {
	l.Logf(fmtString, LevelFatal, args...)
	os.Exit(1)
}

// LogFatalAttrs emits a log message with the FATAL level and the given attributes, then calls os.Exit(1).
func (l *logger) LogFatalAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelFatal, args...)
	os.Exit(1)
}

// LogPanic emits a log message with the PANIC level, then panics.
func (l *logger) LogPanic(msg string, args ...any) {
	l.Log(msg, LevelPanic, args...)
	panic(msg + fmt.Sprint(args...))
}

// LogPanicf emits a formatted log message with the PANIC level, then panics.
func (l *logger) LogPanicf(fmtString string, args ...any) {
	l.Logf(fmtString, LevelPanic, args...)
	panic(fmt.Sprintf(fmtString, args...))
}

// LogPanicAttrs emits a log message with the PANIC level and the given attributes, then panics.
func (l *logger) LogPanicAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelPanic, args...)
	panic(getSlogMessage(msg, args...))
}

// Log emits a log message with the given level.
func (l *logger) Log(msg string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.Log(context.Background(), level, msg, append([]interface{}{namespaceKey, l.path}, args...)...)
	}
}

// Logf emits a formatted log message with the given level.
func (l *logger) Logf(fmtString string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, fmt.Sprintf(fmtString, args...))
	}
}

// LogAttrs emits a log message with the given level and attributes.
func (l *logger) LogAttrs(msg string, level Level, args ...slog.Attr) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, msg, append([]slog.Attr{{Key: namespaceKey, Value: slog.StringValue(l.path)}}, args...)...)
	}
}

// NewChildLogger creates a new child logger with the given name.
func (l *logger) NewChildLogger(name string, enumerateChildren ...bool) (childLogger Logger) {
	if l == nil {
		return l
	}

	if len(enumerateChildren) > 0 && enumerateChildren[0] {
		name = l.uniqueEntityName(name)
	}

	return newLogger(l.rootLogger, l, name)
}

// ParentLogger returns the parent logger of the logger (or nil if it is the root).
func (l *logger) ParentLogger() Logger {
	if l.parentLogger == nil {
		return nil
	}

	return l.parentLogger
}

// UnsubscribeFromParentLogger unsubscribes the logger from its parent logger (e.g. updates about the log level).
// It is important to call this method whenever we remove all references to the logger, otherwise the logger will
// not be garbage collected.
func (l *logger) UnsubscribeFromParentLogger() {
	if l.unsubscribeFromParent != nil {
		l.unsubscribeFromParent()
	}
}

// uniqueEntityName returns the name of an embedded instance of the given type.
func (l *logger) uniqueEntityName(name string) (uniqueName string) {
	entityNameCounter := func() int64 {
		instanceCounter, _ := l.entityNameCounters.LoadOrStore(name, &atomic.Int64{})

		//nolint:forcetypeassert // false positive
		return instanceCounter.(*atomic.Int64).Add(1) - 1
	}

	var nameBuilder strings.Builder
	nameBuilder.WriteString(name)
	nameBuilder.WriteString(strconv.FormatInt(entityNameCounter(), 10))

	return nameBuilder.String()
}

// namespaceKey is the key of the slog attribute that holds the namespace of the logger.
const namespaceKey = "namespace"

func getSlogMessage(msg string, args ...slog.Attr) string {
	if len(args) == 0 {
		return msg
	}

	attributes := ""
	for i, attr := range args {
		attributes += attr.String()
		if i < len(args)-1 {
			attributes += ", "
		}
	}

	return fmt.Sprintf(msg, attributes)
}
