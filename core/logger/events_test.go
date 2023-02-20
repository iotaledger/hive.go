package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/iotaledger/hive.go/lo"
)

const (
	testName      = "test"
	testMsg       = "123"
	testLoggedMsg = testMsg
)

func TestNewEventCore(t *testing.T) {
	// initialize the mock
	m, teardown := newEventMock(t)

	t.Run("levelDisabled", func(t *testing.T) {
		logger := zap.New(NewEventCore(LevelWarn))

		// there should not be any events, as info is below warning.
		logger.Info(testMsg)
	})

	t.Run("eventsTriggered", func(t *testing.T) {
		logger := zap.New(NewEventCore(LevelDebug)).Named(testName)

		m.On("debug", LevelDebug, testName, testLoggedMsg).Once()
		m.On("any", LevelDebug, testName, testLoggedMsg).Once()
		logger.Debug(testMsg)

		m.On("info", LevelInfo, testName, testLoggedMsg).Once()
		m.On("any", LevelInfo, testName, testLoggedMsg).Once()
		logger.Info(testMsg)

		m.On("warn", LevelWarn, testName, testLoggedMsg).Once()
		m.On("any", LevelWarn, testName, testLoggedMsg).Once()
		logger.Warn(testMsg)

		m.On("error", LevelError, testName, testLoggedMsg).Once()
		m.On("any", LevelError, testName, testLoggedMsg).Once()
		logger.Error(testMsg)

		m.On("panic", LevelPanic, testName, testLoggedMsg).Once()
		m.On("any", LevelPanic, testName, testLoggedMsg).Once()
		assert.Panics(t, func() { logger.Panic(testMsg) }, testMsg)

		m.AssertExpectations(t)
	})

	// remove the mock
	teardown()
}

type eventMock struct{ mock.Mock }

func (e *eventMock) debug(lvl Level, name string, msg string) { e.Called(lvl, name, msg) }
func (e *eventMock) info(lvl Level, name string, msg string)  { e.Called(lvl, name, msg) }
func (e *eventMock) warn(lvl Level, name string, msg string)  { e.Called(lvl, name, msg) }
func (e *eventMock) error(lvl Level, name string, msg string) { e.Called(lvl, name, msg) }
func (e *eventMock) panic(lvl Level, name string, msg string) { e.Called(lvl, name, msg) }
func (e *eventMock) any(lvl Level, name string, msg string)   { e.Called(lvl, name, msg) }

func newEventMock(t *testing.T) (*eventMock, func()) {
	m := &eventMock{}
	m.Test(t)

	return m, lo.Batch(
		Events.DebugMsg.Hook(func(event *LogEvent) { m.debug(event.Level, event.Name, event.Msg) }).Unhook,
		Events.InfoMsg.Hook(func(event *LogEvent) { m.info(event.Level, event.Name, event.Msg) }).Unhook,
		Events.WarningMsg.Hook(func(event *LogEvent) { m.warn(event.Level, event.Name, event.Msg) }).Unhook,
		Events.ErrorMsg.Hook(func(event *LogEvent) { m.error(event.Level, event.Name, event.Msg) }).Unhook,
		Events.PanicMsg.Hook(func(event *LogEvent) { m.panic(event.Level, event.Name, event.Msg) }).Unhook,
		Events.AnyMsg.Hook(func(event *LogEvent) { m.any(event.Level, event.Name, event.Msg) }).Unhook,
	)
}
