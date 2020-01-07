package logger

import (
	"testing"

	"github.com/iotaledger/hive.go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const (
	testName      = "test"
	testMsg       = "123"
	testLoggedMsg = testMsg
)

func TestNewEventCore(t *testing.T) {
	// initialize the mock
	m, teardown := newEventMock(t)

	t.Run("debugLevel", func(t *testing.T) {
		logger := zap.New(NewEventCore(LevelDebug))

		// there should not be any events for debug-level logs.
		logger.Debug(testMsg)
	})

	t.Run("levelDisabled", func(t *testing.T) {
		logger := zap.New(NewEventCore(LevelWarn))

		// there should not be any events, as info is below warning.
		logger.Info(testMsg)
	})

	t.Run("eventsTriggered", func(t *testing.T) {
		logger := zap.New(NewEventCore(LevelInfo)).Named(testName)

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

func (e *eventMock) info(lvl Level, name string, msg string)  { e.Called(lvl, name, msg) }
func (e *eventMock) warn(lvl Level, name string, msg string)  { e.Called(lvl, name, msg) }
func (e *eventMock) error(lvl Level, name string, msg string) { e.Called(lvl, name, msg) }
func (e *eventMock) panic(lvl Level, name string, msg string) { e.Called(lvl, name, msg) }
func (e *eventMock) any(lvl Level, name string, msg string)   { e.Called(lvl, name, msg) }

func newEventMock(t *testing.T) (*eventMock, func()) {
	m := &eventMock{}
	m.Test(t)

	infoC := events.NewClosure(m.info)
	warnC := events.NewClosure(m.warn)
	errorC := events.NewClosure(m.error)
	panicC := events.NewClosure(m.panic)
	anyC := events.NewClosure(m.any)

	Events.InfoMsg.Attach(infoC)
	Events.WarningMsg.Attach(warnC)
	Events.ErrorMsg.Attach(errorC)
	Events.PanicMsg.Attach(panicC)
	Events.AnyMsg.Attach(anyC)

	teardown := func() {
		Events.InfoMsg.Detach(infoC)
		Events.WarningMsg.Detach(warnC)
		Events.ErrorMsg.Detach(errorC)
		Events.PanicMsg.Detach(panicC)
		Events.AnyMsg.Detach(anyC)
	}
	return m, teardown
}
