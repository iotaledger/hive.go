package logger

import (
	"github.com/iotaledger/hive.go/events"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Events contains all the events that are triggered by the logger.
var Events = struct {
	DebugMsg   *events.Event
	InfoMsg    *events.Event
	WarningMsg *events.Event
	ErrorMsg   *events.Event
	PanicMsg   *events.Event
	AnyMsg     *events.Event
}{
	DebugMsg:   events.NewEvent(logCaller),
	InfoMsg:    events.NewEvent(logCaller),
	WarningMsg: events.NewEvent(logCaller),
	ErrorMsg:   events.NewEvent(logCaller),
	PanicMsg:   events.NewEvent(logCaller),
	AnyMsg:     events.NewEvent(logCaller),
}

func logCaller(handler interface{}, params ...interface{}) {
	handler.(func(Level, string, string))(params[0].(Level), params[1].(string), params[2].(string))
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
		Events.DebugMsg.Trigger(ent.Level, ent.LoggerName, msg)
	case zapcore.InfoLevel:
		Events.InfoMsg.Trigger(ent.Level, ent.LoggerName, msg)
	case zapcore.WarnLevel:
		Events.WarningMsg.Trigger(ent.Level, ent.LoggerName, msg)
	case zapcore.ErrorLevel:
		Events.ErrorMsg.Trigger(ent.Level, ent.LoggerName, msg)
	case zapcore.PanicLevel:
		Events.PanicMsg.Trigger(ent.Level, ent.LoggerName, msg)
	}
	Events.AnyMsg.Trigger(ent.Level, ent.LoggerName, msg)

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
