package reactive

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	logger := NewLogger("node1")
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

	time.Sleep(1 * time.Second) // wait for log message to be printed
}

func TestEntityBasedLogging(t *testing.T) {
	logger := NewLogger("node1")

	testObject1 := NewTestObject(logger)
	testObject1.ImportantValue1.Set(1)      // will produce a log message
	testObject1.LessImportantValue1.Set(10) // will not produce a log message
	testObject1.SetLogLevel(LogLevelDebug)
	testObject1.ImportantValue1.Set(10)     // will produce a log message
	testObject1.LessImportantValue1.Set(20) // will produce a log message
	testObject1.SetLogLevel(LogLevelInfo)
	testObject1.LessImportantValue1.Set(40) // will not produce a log message
	testObject1.ImportantValue1.Set(100)    // will produce a log message
	testObject1.SetLogLevel(LogLevelWarning)
	testObject1.LessImportantValue1.Set(40) // will not produce a log message
	testObject1.ImportantValue1.Set(100)    // will not produce a log message

	testObject2 := NewTestObject(logger)
	testObject2.ImportantValue1.Set(1)      // will produce a log message
	testObject2.LessImportantValue1.Set(10) // will not produce a log message
	testObject2.SetLogLevel(LogLevelDebug)
	testObject2.ImportantValue1.Set(10)     // will produce a log message
	testObject2.LessImportantValue1.Set(20) // will produce a log message
	testObject2.SetLogLevel(LogLevelInfo)
	testObject2.LessImportantValue1.Set(40) // will not produce a log message
	testObject2.ImportantValue1.Set(100)    // will produce a log message
	testObject2.SetLogLevel(LogLevelWarning)
	testObject2.LessImportantValue1.Set(40) // will not produce a log message
	testObject2.ImportantValue1.Set(100)    // will not produce a log message

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
		t.Logger = NewEmbeddedLogger(logger, "TestObject", t.IsEvicted)

		t.ImportantValue1.LogUpdates(t.Logger, LogLevelInfo, "ImportantValue1")
		t.ImportantValue2.LogUpdates(t.Logger, LogLevelInfo, "ImportantValue2")
		t.LessImportantValue1.LogUpdates(t.Logger, LogLevelDebug, "LessImportantValue1")
	}

	return t
}
