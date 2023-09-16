package logger

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/lo"
)

func TestLogger(t *testing.T) {
	logger := New("node1")
	logger.LogDebug("some log (invisible due to log level)")
	logger.SetLogLevel(LevelTrace)
	logger.Trace("created chain", "id", "chain1")

	networkLogger, shutdownNetworkLogger := logger.NestedLogger("network")
	defer shutdownNetworkLogger()
	networkLogger.SetLogLevel(LevelInfo)
	networkLogger.LogDebug("instantiated chain (invisible)", "id", 1)

	chainLogger, shutdownChainLogger := logger.NestedLogger("chain1")
	defer shutdownChainLogger()
	chainLogger.SetLogLevel(LevelDebug)
	chainLogger.LogDebug("attested weight updated (visible)", "oldWeight", 7, "newWeight", 10)
}

func TestEntityBasedLogging(t *testing.T) {
	logger := New("node1")

	testObject := NewTestObject(logger)
	logger.LogInfo("created " + testObject.name)

	testObject.ImportantValue1.Set(1)      // will produce a log message
	testObject.LessImportantValue1.Set(10) // will not produce a log message
	testObject.SetLogLevel(LevelDebug)
	testObject.ImportantValue1.Set(10)     // will produce a log message
	testObject.LessImportantValue1.Set(20) // will produce a log message
	testObject.SetLogLevel(LevelInfo)
	testObject.LessImportantValue1.Set(40) // will not produce a log message
	testObject.ImportantValue1.Set(100)    // will produce a log message
	testObject.SetLogLevel(LevelWarning)
	testObject.LessImportantValue1.Set(40) // will not produce a log message
	testObject.ImportantValue1.Set(100)    // will not produce a log message

	time.Sleep(1 * time.Second) // wait for log message to be printed
}

type TestObject struct {
	ImportantValue1     reactive.Variable[uint64]
	ImportantValue2     reactive.Variable[uint64]
	LessImportantValue1 reactive.Variable[uint64]
	IsEvicted           reactive.Event

	*Logger
}

func NewTestObject(logger *Logger) *TestObject {
	t := &TestObject{
		ImportantValue1:     reactive.NewVariable[uint64](),
		ImportantValue2:     reactive.NewVariable[uint64](),
		LessImportantValue1: reactive.NewVariable[uint64](),
		IsEvicted:           reactive.NewEvent(),
	}

	t.initLogging(logger)

	return t
}

func (t *TestObject) initLogging(logger *Logger) {
	t.Logger = NewEntityLogger(logger, "TestObject", t.IsEvicted)

	t.Logger.OnLogLevelInfo(func() (shutdown func()) {
		return lo.Batch(
			t.ImportantValue1.OnUpdate(LogReactiveVariableUpdate[uint64](t.Logger.LogInfo, "ImportantValue1")),
			t.ImportantValue2.OnUpdate(LogReactiveVariableUpdate[uint64](t.Logger.LogInfo, "ImportantValue2")),
		)
	})

	t.Logger.OnLogLevelDebug(func() (shutdown func()) {
		return t.LessImportantValue1.OnUpdate(LogReactiveVariableUpdate[uint64](t.Logger.LogDebug, "LessImportantValue1"))
	})
}
