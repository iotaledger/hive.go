package reactive

import (
	"testing"
	"time"
)

// TestLogger tests the logger by using the traditional logging methods that align with the slog interface and the
// ability to create nested loggers with individual log levels.
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

// TestEntityBasedLogging tests the entity based logging.
func TestEntityBasedLogging(t *testing.T) {
	logger := NewLogger("node1")

	testObject0 := NewTestObject(logger)
	testObject0.ImportantValue1.Set(1)      // will produce a log message
	testObject0.LessImportantValue1.Set(10) // will not produce a log message
	testObject0.SetLogLevel(LogLevelDebug)
	testObject0.ImportantValue1.Set(10)     // will produce a log message
	testObject0.LessImportantValue1.Set(20) // will produce a log message
	testObject0.SetLogLevel(LogLevelInfo)
	testObject0.LessImportantValue1.Set(40) // will not produce a log message
	testObject0.ImportantValue1.Set(100)    // will produce a log message
	testObject0.SetLogLevel(LogLevelWarning)
	testObject0.LessImportantValue1.Set(40) // will not produce a log message
	testObject0.ImportantValue1.Set(100)    // will not produce a log message

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

	t.Logger = NewEmbeddedLogger(logger, "TestObject", t.IsEvicted, func(embeddedLogger *Logger) {
		t.ImportantValue1.LogUpdates(embeddedLogger, LogLevelInfo, "ImportantValue1")
		t.ImportantValue2.LogUpdates(embeddedLogger, LogLevelInfo, "ImportantValue2")
		t.LessImportantValue1.LogUpdates(embeddedLogger, LogLevelDebug, "LessImportantValue1")
	})

	return t
}
