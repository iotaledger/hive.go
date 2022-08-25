package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iotaledger/hive.go/core/generics/event"
)

var Events *EventsStruct

// EventsStruct contains all the events that are triggered by the logger.
type EventsStruct struct {
	DebugMsg   *event.Event[*LogEvent]
	InfoMsg    *event.Event[*LogEvent]
	WarningMsg *event.Event[*LogEvent]
	ErrorMsg   *event.Event[*LogEvent]
	PanicMsg   *event.Event[*LogEvent]
	AnyMsg     *event.Event[*LogEvent]
}

func newEventsStruct() (new *EventsStruct) {
	return &EventsStruct{
		DebugMsg:   event.New[*LogEvent](),
		InfoMsg:    event.New[*LogEvent](),
		WarningMsg: event.New[*LogEvent](),
		ErrorMsg:   event.New[*LogEvent](),
		PanicMsg:   event.New[*LogEvent](),
		AnyMsg:     event.New[*LogEvent](),
	}
}

type LogEvent struct {
	Level Level
	Name  string
	Msg   string
}

func init() {
	Events = newEventsStruct()
}

// NewEventCore creates a core that publishes log messages as events.
func NewEventCore(enabler zapcore.LevelEnabler) zapcore.Core {
	enablerFunc := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return enabler.Enabled(lvl)
	})
	enc := zapcore.NewConsoleEncoder(eventCoreEncoderConfig)

	return &eventCore{
		LevelEnabler: enablerFunc,
		enc:          enc,
	}
}

var eventCoreEncoderConfig = zapcore.EncoderConfig{
	MessageKey:     "M", // show encoded message
	LevelKey:       "",  // hide log level
	TimeKey:        "",  // hide timestamp
	NameKey:        "",  // hide logger name
	CallerKey:      "",  // hide log caller
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeTime:     zapcore.RFC3339TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeName:     zapcore.FullNameEncoder,
}

type eventCore struct {
	zapcore.LevelEnabler
	enc zapcore.Encoder
}

func (c *eventCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	for i := range fields {
		fields[i].AddTo(clone.enc)
	}

	return clone
}

func (c *eventCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *eventCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	buf.TrimNewline()
	msg := buf.String()

	switch ent.Level {
	case zapcore.DebugLevel:
		Events.DebugMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})
	case zapcore.InfoLevel:
		Events.InfoMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})
	case zapcore.WarnLevel:
		Events.WarningMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})
	case zapcore.ErrorLevel:
		Events.ErrorMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})
	case zapcore.PanicLevel:
		Events.PanicMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})
	}
	Events.AnyMsg.Trigger(&LogEvent{ent.Level, ent.LoggerName, msg})

	return nil
}

func (c *eventCore) Sync() error {
	return nil
}

func (c *eventCore) clone() *eventCore {
	return &eventCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
	}
}
