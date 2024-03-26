package log_test

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/ds/reactive"
	"github.com/iotaledger/hive.go/log"
)

// TestLogger tests the logger by using the traditional logging methods that align with the slog interface and the
// ability to create nested loggers with individual log levels.
func TestLogger(t *testing.T) {
	logger := log.NewLogger(log.WithName("node1"))
	logger.LogDebug("some log (invisible due to log level)")
	logger.SetLogLevel(log.LevelTrace)
	logger.LogTrace("created chain")

	networkLogger := logger.NewChildLogger("network")
	defer networkLogger.Shutdown()
	networkLogger.SetLogLevel(log.LevelInfo)
	networkLogger.LogInfo("instantiated chain (invisible)", "id", 1)

	chainLogger := logger.NewChildLogger("chain1")
	defer chainLogger.Shutdown()
	chainLogger.SetLogLevel(log.LevelDebug)
	chainLogger.LogDebug("attested weight updated (visible)", "oldWeight", 7, "newWeight", 10)
	logger.LogTrace("shutdown")

	time.Sleep(1 * time.Second) // wait for log message to be printed
}

// TestEntityBasedLogging tests the entity based logging.
func TestEntityBasedLogging(t *testing.T) {
	logger := log.NewLogger(log.WithName("node1"))

	testObject0 := NewTestObject(logger)
	testObject0.ImportantValue1.Set(1)      // will produce a log message
	testObject0.LessImportantValue1.Set(10) // will not produce a log message
	testObject0.SetLogLevel(log.LevelDebug)
	testObject0.ImportantValue1.Set(10)     // will produce a log message
	testObject0.LessImportantValue1.Set(20) // will produce a log message
	testObject0.SetLogLevel(log.LevelInfo)
	testObject0.LessImportantValue1.Set(40) // will not produce a log message
	testObject0.ImportantValue1.Set(100)    // will produce a log message
	testObject0.SetLogLevel(log.LevelWarning)
	testObject0.LessImportantValue1.Set(40) // will not produce a log message
	testObject0.ImportantValue1.Set(100)    // will not produce a log message

	testObject1 := NewTestObject(logger)
	testObject1.ImportantValue1.Set(1)      // will produce a log message
	testObject1.LessImportantValue1.Set(10) // will not produce a log message
	testObject1.SetLogLevel(log.LevelDebug)
	testObject1.ImportantValue1.Set(10)     // will produce a log message
	testObject1.LessImportantValue1.Set(20) // will produce a log message
	testObject1.SetLogLevel(log.LevelInfo)
	testObject1.LessImportantValue1.Set(40) // will not produce a log message
	testObject1.ImportantValue1.Set(100)    // will produce a log message
	testObject1.SetLogLevel(log.LevelWarning)
	testObject1.LessImportantValue1.Set(40) // will not produce a log message
	testObject1.ImportantValue1.Set(100)    // will not produce a log message

	time.Sleep(1 * time.Second) // wait for log message to be printed
}

type TestObject struct {
	ImportantValue1     reactive.Variable[uint64]
	ImportantValue2     reactive.Variable[uint64]
	LessImportantValue1 reactive.Variable[uint64]
	IsEvicted           reactive.Event

	log.Logger
}

func NewTestObject(logger log.Logger) *TestObject {
	t := &TestObject{
		ImportantValue1:     reactive.NewVariable[uint64](),
		ImportantValue2:     reactive.NewVariable[uint64](),
		LessImportantValue1: reactive.NewVariable[uint64](),
		IsEvicted:           reactive.NewEvent(),
	}

	t.Logger = logger.NewChildLogger("TestObject", true)

	t.ImportantValue1.LogUpdates(t.Logger, log.LevelInfo, "ImportantValue1")
	t.ImportantValue2.LogUpdates(t.Logger, log.LevelInfo, "ImportantValue2")
	t.LessImportantValue1.LogUpdates(t.Logger, log.LevelDebug, "LessImportantValue1")

	return t
}
