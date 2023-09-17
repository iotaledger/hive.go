package reactive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/lo"
)

// Logger is a reactive logger that can be used to log messages with different log levels.
type Logger struct {
	name          string
	namespace     string
	rootLogger    *slog.Logger
	level         *slog.LevelVar
	reactiveLevel Variable[LogLevel]
}

// NewLogger creates a new logger with the given namespace and an optional handler. The default handler prints log
// records in a human-readable format to stdout.
func NewLogger(name string, handler ...slog.Handler) *Logger {
	return newLogger("", name, slog.New(lo.First(handler, newDefaultLogHandler(os.Stdout))))
}

// NewEmbeddedLogger creates a logger for an entity of a specific type that can be embedded into the entity's struct.
// The logger's name is a combination of the name of the type and an ever-increasing instance counter. The logger is
// automatically closed when the shutdown event is triggered.
func NewEmbeddedLogger(logger *Logger, typeName string, shutdownEvent Event, initLogging func(embeddedLogger *Logger)) *Logger {
	if logger == nil {
		return nil
	}

	embeddedLogger, shutdown := logger.NestedLogger(embeddedInstanceName(typeName))
	shutdownEvent.OnTrigger(shutdown)

	initLogging(embeddedLogger)

	return embeddedLogger
}

// newLogger creates a new logger with the given namespace and root logger instance.
func newLogger(namespace, name string, rootLogger *slog.Logger) *Logger {
	l := &Logger{
		name:          name,
		namespace:     lo.Cond(namespace == "", name, namespace+"."+name),
		rootLogger:    rootLogger,
		level:         new(slog.LevelVar),
		reactiveLevel: NewVariable[LogLevel](),
	}

	l.reactiveLevel.OnUpdate(func(_, newLevel LogLevel) { l.level.Set(newLevel) })

	return l
}

// LoggerName returns the name of the logger (the last part of the namespace).
func (l *Logger) LoggerName() string {
	return l.name
}

// LoggerNamespace returns the full namespace of the logger.
func (l *Logger) LoggerNamespace() string {
	return l.namespace
}

// Log logs a message with the given log level.
func (l *Logger) Log(msg string, level LogLevel, args ...any) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.Log(context.Background(), level, msg, append([]interface{}{namespaceKey, l.namespace}, args...)...)
	}
}

// LogAttrs logs a message with the given log level and typed slog attributes.
func (l *Logger) LogAttrs(msg string, level LogLevel, args ...slog.Attr) {
	if l != nil && l.level.Level() <= level {
		l.rootLogger.LogAttrs(context.Background(), level, msg, append([]slog.Attr{{Key: namespaceKey, Value: slog.StringValue(l.namespace)}}, args...)...)
	}
}

// LogTrace logs a trace message.
func (l *Logger) LogTrace(msg string, args ...any) {
	l.Log(msg, LogLevelTrace, args...)
}

// LogTraceAttrs logs a trace message with typed slog attributes.
func (l *Logger) LogTraceAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LogLevelTrace, args...)
}

// LogDebug logs a debug message.
func (l *Logger) LogDebug(msg string, args ...any) {
	l.Log(msg, LogLevelDebug, args...)
}

// LogDebugAttrs logs a debug message with typed slog attributes.
func (l *Logger) LogDebugAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LogLevelDebug, args...)
}

// LogInfo logs an info message.
func (l *Logger) LogInfo(msg string, args ...any) {
	l.Log(msg, LogLevelInfo, args...)
}

// LogInfoAttrs logs an info message with typed slog attributes.
func (l *Logger) LogInfoAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LogLevelInfo, args...)
}

// LogWarn logs a warning message.
func (l *Logger) LogWarn(msg string, args ...any) {
	l.Log(msg, LogLevelWarning, args...)
}

// LogWarnAttrs logs a warning message with typed slog attributes.
func (l *Logger) LogWarnAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LogLevelWarning, args...)
}

// LogError logs an error message.
func (l *Logger) LogError(msg string, args ...any) {
	l.Log(msg, LogLevelError, args...)
}

// LogErrorAttrs logs an error message with typed slog attributes.
func (l *Logger) LogErrorAttrs(msg string, args ...slog.Attr) {
	l.LogAttrs(msg, LogLevelError, args...)
}

// SetLogLevel sets the log level of the logger.
func (l *Logger) SetLogLevel(level LogLevel) {
	if l != nil {
		l.reactiveLevel.Set(level)
	}
}

// OnLogLevel registers a callback that is called when the log level is set to the given log level or lower. The
// callback has to return a shutdown function that is called when the log level is set to a higher level.
func (l *Logger) OnLogLevel(logLevel LogLevel, setup func() (shutdown func())) (unsubscribe func()) {
	if l == nil {
		return func() {}
	}

	var shutdownEvent Event

	unsubscribeFromLevel := l.reactiveLevel.OnUpdate(func(_, newLevel LogLevel) {
		if newLevel <= logLevel {
			if shutdownEvent == nil {
				shutdownEvent = NewEvent()
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
func (l *Logger) NestedLogger(subNamespace string) (nestedLogger *Logger, shutdown func()) {
	if l == nil {
		return nil, func() {}
	}

	nestedLogger = newLogger(l.namespace, subNamespace, l.rootLogger)

	return nestedLogger, nestedLogger.reactiveLevel.InheritFrom(l.reactiveLevel)
}

func (l *Logger) String() string {
	return strings.TrimRight(fmt.Sprintf("Logger[%s] (LEVEL = %s", l.namespace, LogLevelName(l.level.Level())), " ") + ")"
}

// embeddedInstanceName returns the name of an embedded instance of the given type.
func embeddedInstanceName(typeName string) string {
	embeddedInstanceCounter := func() int64 {
		instanceCounter, loaded := embeddedInstanceCounters.Load(typeName)
		if loaded {
			return instanceCounter.(*atomic.Int64).Add(1) - 1
		}

		instanceCounter, _ = embeddedInstanceCounters.LoadOrStore(typeName, &atomic.Int64{})

		return instanceCounter.(*atomic.Int64).Add(1) - 1
	}

	var nameBuilder strings.Builder

	nameBuilder.WriteString(typeName)
	nameBuilder.WriteString(strconv.FormatInt(embeddedInstanceCounter(), 10))

	return nameBuilder.String()
}

// embeddedInstanceCounters holds the instance counters for embedded loggers.
var embeddedInstanceCounters = sync.Map{}

// namespaceKey is the key of the slog attribute that holds the namespace of the logger.
const namespaceKey = "namespace"

// LogLevel is the type of log levels.
type LogLevel = slog.Level

const (
	// LogLevelTrace is the log level for trace messages.
	LogLevelTrace = slog.Level(-8)

	// LogLevelDebug is the log level for debug messages.
	LogLevelDebug = slog.LevelDebug

	// LogLevelInfo is the log level for info messages.
	LogLevelInfo = slog.LevelInfo

	// LogLevelWarning is the log level for warning messages.
	LogLevelWarning = slog.LevelWarn

	// LogLevelError is the log level for error messages.
	LogLevelError = slog.LevelError
)

// LogLevelName returns the name of the given log level.
func LogLevelName(level LogLevel) string {
	switch level {
	case LogLevelTrace:
		return "TRACE  "
	case LogLevelDebug:
		return "DEBUG  "
	case LogLevelInfo:
		return "INFO   "
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR  "
	default:
		return "UNKNOWN"
	}
}

// newDefaultLogHandler creates a new default handler that writes human-readable log records to the given output.
func newDefaultLogHandler(output io.Writer) slog.Handler {
	return &defaultLogHandler{output: output}
}

// defaultLogHandler is a slog.Handler that writes human-readable log records to an output.
type defaultLogHandler struct {
	output io.Writer
}

// Enabled returns true for all levels as we handle the cutoff ourselves using reactive variables and the ability to
// set loggers to nil.
func (d *defaultLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle writes the log record to the output.
func (d *defaultLogHandler) Handle(_ context.Context, r slog.Record) error {
	var namespace string
	fieldsBuffer := new(bytes.Buffer)

	fieldCount := r.NumAttrs() - 1
	if fieldCount > 0 {
		fieldsBuffer.WriteString("(")
	}

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == namespaceKey {
			namespace = attr.Value.Any().(string)
		} else {
			fieldsBuffer.WriteString(attr.String())
			fieldsBuffer.WriteString(" ")
		}

		return true
	})

	if fieldCount > 0 {
		fieldsBuffer.Truncate(fieldsBuffer.Len() - 1)
		fieldsBuffer.WriteString(")")
	}

	fmt.Fprintf(d.output, "%s\t%s\t%s\t%s \t\t %s\n", r.Time.Format("2006/01/02 15:04:05"), LogLevelName(r.Level), namespace, r.Message, fieldsBuffer.String())

	return nil
}

// WithAttrs is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (d *defaultLogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not supported")
}

// WithGroup is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (d *defaultLogHandler) WithGroup(_ string) slog.Handler {
	panic("not supported")
}