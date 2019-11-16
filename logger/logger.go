package logger

import (
	"fmt"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/syncutils"
	"io"
	"log"
	"os"
)

func init() {
	// auto. route log msgs into the any msg event
	anyRouter := events.NewClosure(func(logLvl LogLevel, prefix string, msg string) {
		Events.AnyMsg.Trigger(logLvl, prefix, msg)
	})
	Events.InfoMsg.Attach(anyRouter)
	Events.NoticeMsg.Attach(anyRouter)
	Events.WarningMsg.Attach(anyRouter)
	Events.ErrorMsg.Attach(anyRouter)
	Events.CriticalMsg.Attach(anyRouter)
	Events.PanicMsg.Attach(anyRouter)
	Events.FatalMsg.Attach(anyRouter)
	Events.DebugMsg.Attach(anyRouter)
}

func LogCaller(handler interface{}, params ...interface{}) {
	handler.(func(LogLevel, string, string))(params[0].(LogLevel), params[1].(string), params[2].(string))
}

var Events = loggerevents{
	InfoMsg:     events.NewEvent(LogCaller),
	NoticeMsg:   events.NewEvent(LogCaller),
	WarningMsg:  events.NewEvent(LogCaller),
	ErrorMsg:    events.NewEvent(LogCaller),
	CriticalMsg: events.NewEvent(LogCaller),
	PanicMsg:    events.NewEvent(LogCaller),
	FatalMsg:    events.NewEvent(LogCaller),
	DebugMsg:    events.NewEvent(LogCaller),
	AnyMsg:      events.NewEvent(LogCaller),
}

type loggerevents struct {
	InfoMsg     *events.Event
	NoticeMsg   *events.Event
	WarningMsg  *events.Event
	ErrorMsg    *events.Event
	CriticalMsg *events.Event
	PanicMsg    *events.Event
	FatalMsg    *events.Event
	DebugMsg    *events.Event
	AnyMsg      *events.Event
}

// every instance of the logger uses the same logger to ensure that
// concurrent prints/writes don't overlap
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// InjectWriter swaps out the logger package's underlying logger's writers.
func InjectWriters(writers ...io.Writer) {
	logger = log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime)
}

func NewLogger(prefix string, logLevel ...LogLevel) *Logger {
	l := &Logger{Prefix: prefix}
	if len(logLevel) > 0 {
		l.logLevel = logLevel[0]
	} else {
		l.logLevel = LevelNormal
	}
	return l
}

type LogLevel byte

const (
	LevelInfo     LogLevel = 1
	LevelNotice            = LevelInfo << 1
	LevelWarning           = LevelNotice << 1
	LevelError             = LevelWarning << 1
	LevelCritical          = LevelError << 1
	LevelPanic             = LevelCritical << 1
	LevelFatal             = LevelPanic << 1
	LevelDebug             = LevelFatal << 1

	LevelNormal = LevelInfo | LevelNotice | LevelWarning | LevelError | LevelCritical | LevelPanic | LevelFatal
)

type Logger struct {
	Prefix     string
	changeMu   syncutils.Mutex
	logLevel   LogLevel
	disabledMu syncutils.Mutex
	disabled   bool
}

func (l *Logger) Enabled() bool {
	return !l.disabled
}

func (l *Logger) Enable() {
	l.disabledMu.Lock()
	l.disabled = false
	l.disabledMu.Unlock()
}

func (l *Logger) Disable() {
	l.disabledMu.Lock()
	l.disabled = true
	l.disabledMu.Unlock()
}

func (l *Logger) ChangeLogLevel(logLevel LogLevel) {
	l.changeMu.Lock()
	l.logLevel = logLevel
	l.changeMu.Unlock()
}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	if l.logLevel&LevelFatal == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.FatalMsg.Trigger(LevelFatal, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ FATAL ] %s:", l.Prefix), msg)
	os.Exit(1)
}

// Fatalf is equivalent to l.Critical followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.logLevel&LevelFatal == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.FatalMsg.Trigger(LevelFatal, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ FATAL ] %s:", l.Prefix), msg)
	os.Exit(1)
}

// Panic is equivalent to l.Critical(fmt.Sprint()) followed by a call to panic().
func (l *Logger) Panic(args ...interface{}) {
	if l.logLevel&LevelPanic == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.PanicMsg.Trigger(LevelPanic, l.Prefix, msg)
	logger.Panicln(fmt.Sprintf("[ PANIC ] %s:", l.Prefix), msg)
}

// Panicf is equivalent to l.Critical followed by a call to panic().
func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.logLevel&LevelPanic == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.PanicMsg.Trigger(LevelPanic, l.Prefix, msg)
	logger.Panicln(fmt.Sprintf("[ PANIC ] %s:", l.Prefix), msg)
}

// Critical logs a message using CRITICAL as log level.
func (l *Logger) Critical(args ...interface{}) {
	if l.logLevel&LevelCritical == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.CriticalMsg.Trigger(LevelCritical, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ CRITICAL ] %s:", l.Prefix), msg)
}

// Criticalf logs a message using CRITICAL as log level.
func (l *Logger) Criticalf(format string, args ...interface{}) {
	if l.logLevel&LevelCritical == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.CriticalMsg.Trigger(LevelCritical, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ CRITICAL ] %s:", l.Prefix), msg)
}

// Error logs a message using ERROR as log level.
func (l *Logger) Error(args ...interface{}) {
	if l.logLevel&LevelError == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.ErrorMsg.Trigger(LevelError, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ ERROR ] %s:", l.Prefix), msg)
}

// Errorf logs a message using ERROR as log level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.logLevel&LevelError == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.ErrorMsg.Trigger(LevelError, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ ERROR ] %s:", l.Prefix), msg)
}

// Warning logs a message using WARNING as log level.
func (l *Logger) Warning(args ...interface{}) {
	if l.logLevel&LevelWarning == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.WarningMsg.Trigger(LevelWarning, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ WARNING ] %s:", l.Prefix), msg)
}

// Warningf logs a message using WARNING as log level.
func (l *Logger) Warningf(format string, args ...interface{}) {
	if l.logLevel&LevelWarning == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.WarningMsg.Trigger(LevelWarning, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ WARNING ] %s:", l.Prefix), msg)
}

// Notice logs a message using NOTICE as log level.
func (l *Logger) Notice(args ...interface{}) {
	if l.logLevel&LevelNotice == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.NoticeMsg.Trigger(LevelNotice, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ NOTICE ] %s:", l.Prefix), msg)
}

// Noticef logs a message using NOTICE as log level.
func (l *Logger) Noticef(format string, args ...interface{}) {
	if l.logLevel&LevelNotice == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.NoticeMsg.Trigger(LevelNotice, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ NOTICE ] %s:", l.Prefix), msg)
}

// Info logs a message using INFO as log level.
func (l *Logger) Info(args ...interface{}) {
	if l.logLevel&LevelInfo == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.InfoMsg.Trigger(LevelInfo, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ INFO ] %s:", l.Prefix), msg)
}

// Infof logs a message using INFO as log level.
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.logLevel&LevelInfo == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.InfoMsg.Trigger(LevelInfo, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ INFO ] %s:", l.Prefix), msg)
}

// Debug logs a message using DEBUG as log level.
func (l *Logger) Debug(args ...interface{}) {
	if l.logLevel&LevelDebug == 0 || l.disabled {
		return
	}
	msg := fmt.Sprint(args...)
	Events.DebugMsg.Trigger(LevelDebug, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ DEBUG ] %s:", l.Prefix), msg)
}

// Debugf logs a message using DEBUG as log level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.logLevel&LevelDebug == 0 || l.disabled {
		return
	}
	msg := fmt.Sprintf(format, args...)
	Events.DebugMsg.Trigger(LevelDebug, l.Prefix, msg)
	logger.Println(fmt.Sprintf("[ DEBUG ] %s:", l.Prefix), msg)
}
