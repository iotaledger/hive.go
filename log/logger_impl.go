package log

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/lo"
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

	// level is the current log level of the logger.
	level *slog.LevelVar

	// reactiveLevel is the reactive variable that is used to make changes to the log level reactive.
	reactiveLevel reactive.Variable[Level]

	// entityNameCounters holds the instance counters for entity loggers.
	entityNameCounters sync.Map
}

// newLogger creates a new logger instance with the given name and parent logger.
func newLogger(rootLogger *slog.Logger, parentPath, name string) *logger {
	l := &logger{
		name:          name,
		path:          lo.Cond(parentPath == "", name, parentPath+"."+name),
		rootLogger:    rootLogger,
		level:         new(slog.LevelVar),
		reactiveLevel: reactive.NewVariable[Level](),
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

// LogTraceF emits a formatted log message with the TRACE level.
func (l *logger) LogTraceF(fmtString string, args ...any) {
	l.LogF(fmtString, LevelTrace, args...)
}

// LogTraceAttrs emits a log message with the TRACE level and the given attributes.
func (l *logger) LogTraceAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelTrace, args...)
}

// LogDebug emits a log message with the DEBUG level.
func (l *logger) LogDebug(msg string, args ...any) {
	l.Log(msg, LevelDebug, args...)
}

// LogDebugF emits a formatted log message with the DEBUG level.
func (l *logger) LogDebugF(fmtString string, args ...any) {
	l.LogF(fmtString, LevelDebug, args...)
}

// LogDebugAttrs emits a log message with the DEBUG level and the given attributes.
func (l *logger) LogDebugAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelDebug, args...)
}

// LogInfo emits a log message with the INFO level.
func (l *logger) LogInfo(msg string, args ...any) {
	l.Log(msg, LevelInfo, args...)
}

// LogInfoF emits a formatted log message with the INFO level.
func (l *logger) LogInfoF(fmtString string, args ...any) {
	l.LogF(fmtString, LevelInfo, args...)
}

// LogInfoAttrs emits a log message with the INFO level and the given attributes.
func (l *logger) LogInfoAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelInfo, args...)
}

// LogWarn emits a log message with the WARN level.
func (l *logger) LogWarn(msg string, args ...any) {
	l.Log(msg, LevelWarning, args...)
}

// LogWarnF emits a formatted log message with the WARN level.
func (l *logger) LogWarnF(fmtString string, args ...any) {
	l.LogF(fmtString, LevelWarning, args...)
}

// LogWarnAttrs emits a log message with the WARN level and the given attributes.
func (l *logger) LogWarnAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelWarning, args...)
}

// LogError emits a log message with the ERROR level.
func (l *logger) LogError(msg string, args ...any) {
	l.Log(msg, LevelError, args...)
}

// LogErrorF emits a formatted log message with the ERROR level.
func (l *logger) LogErrorF(fmtString string, args ...any) {
	l.LogF(fmtString, LevelError, args...)
}

// LogErrorAttrs emits a log message with the ERROR level and the given attributes.
func (l *logger) LogErrorAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LevelError, args...)
}

// Log emits a log message with the given level.
func (l *logger) Log(msg string, level Level, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.Log(context.Background(), level, msg, append([]interface{}{namespaceKey, l.path}, args...)...)
	}
}

// LogF emits a formatted log message with the given level.
func (l *logger) LogF(fmtString string, level Level, args ...any) {
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
func (l *logger) NewChildLogger(name string) (childLogger Logger, shutdown func()) {
	if l == nil {
		return l, func() {}
	}

	nestedLoggerInstance := newLogger(l.rootLogger, l.path, name)

	return nestedLoggerInstance, nestedLoggerInstance.reactiveLevel.InheritFrom(l.reactiveLevel)
}

// NewEntityLogger creates a new logger for an entity with the given name. The logger is automatically shut down when
// the given shutdown event is triggered. The initLogging function is called with the new logger instance and can be
// used to configure the logger.
func (l *logger) NewEntityLogger(entityName string, shutdownEvent reactive.Event, initLogging func(entityLogger Logger)) Logger {
	if l == nil {
		return l
	}

	embeddedLogger, shutdown := l.NewChildLogger(l.uniqueEntityName(entityName))
	shutdownEvent.OnTrigger(shutdown)

	initLogging(embeddedLogger)

	return embeddedLogger
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
