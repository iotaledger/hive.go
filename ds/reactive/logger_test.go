package reactive

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/lo"
)

func TestLogger(t *testing.T) {
	logger := New("node1")
	logger.LogDebug("some log (invisible due to log level)")
	logger.SetLogLevel(LogLevelTrace)
	logger.LogTrace("created chain", "id", "chain1")

	networkLogger, shutdownNetworkLogger := logger.NestedLogger("network")
	defer shutdownNetworkLogger()
	networkLogger.SetLogLevel(LogLevelInfo)
	networkLogger.LogDebug("instantiated chain (invisible)", "id", 1)

	chainLogger, shutdownChainLogger := logger.NestedLogger("chain1")
	defer shutdownChainLogger()
	chainLogger.SetLogLevel(LogLevelDebug)
	chainLogger.LogDebug("attested weight updated (visible)", "oldWeight", 7, "newWeight", 10)
}

func TestEntityBasedLogging(t *testing.T) {
	logger := New("node1")

	testObject := NewTestObject(logger)

	testObject.ImportantValue1.Set(1)      // will produce a log message
	testObject.LessImportantValue1.Set(10) // will not produce a log message
	testObject.SetLogLevel(LogLevelDebug)
	testObject.ImportantValue1.Set(10)     // will produce a log message
	testObject.LessImportantValue1.Set(20) // will produce a log message
	testObject.SetLogLevel(LogLevelInfo)
	testObject.LessImportantValue1.Set(40) // will not produce a log message
	testObject.ImportantValue1.Set(100)    // will produce a log message
	testObject.SetLogLevel(LogLevelWarning)
	testObject.LessImportantValue1.Set(40) // will not produce a log message
	testObject.ImportantValue1.Set(100)    // will not produce a log message

	time.Sleep(1 * time.Second) // wait for log message to be printed
}

type TestObject struct {
	ImportantValue1     Variable[uint64]
	ImportantValue2     Variable[uint64]
	LessImportantValue1 Variable[uint64]
	IsEvicted           Event

	*Logger
}

func NewTestObject(logger *Logger) *TestObject {
	t := &TestObject{
		ImportantValue1:     NewVariable[uint64](),
		ImportantValue2:     NewVariable[uint64](),
		LessImportantValue1: NewVariable[uint64](),
		IsEvicted:           NewEvent(),
	}

	if logger != nil {
		t.Logger = NewEntityLogger(logger, "TestObject", t.IsEvicted)

		//t.ImportantValue1.LogUpdates(t.Logger, LogLevelInfo, "ImportantValue1")

		t.initLogging()
	}

	return t
}

func (t *TestObject) initLogging() {
	if t.Logger != nil {
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
}
