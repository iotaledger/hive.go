package log

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds/reactive"
)

// NewEntityLogger creates a logger for an entity of a specific type that can be embedded into the entity's struct.
// The logger's name is a combination of the name of the type and an ever-increasing instance counter. The logger is
// automatically closed when the shutdown event is triggered.
func NewEntityLogger(logger Logger, entityName string, shutdownEvent reactive.Event, initLogging func(embeddedLogger Logger)) Logger {
	if logger == nil {
		return nil
	}

	embeddedLogger, shutdown := logger.NewChildLogger(uniqueEntityName(entityName))
	shutdownEvent.OnTrigger(shutdown)

	initLogging(embeddedLogger)

	return embeddedLogger
}

// uniqueEntityName returns the name of an embedded instance of the given type.
func uniqueEntityName(name string) (uniqueName string) {
	entityNameCounter := func() int64 {
		instanceCounter, loaded := entityNameCounters.Load(name)
		if loaded {
			return instanceCounter.(*atomic.Int64).Add(1) - 1
		}

		instanceCounter, _ = entityNameCounters.LoadOrStore(name, &atomic.Int64{})

		return instanceCounter.(*atomic.Int64).Add(1) - 1
	}

	var nameBuilder strings.Builder

	nameBuilder.WriteString(name)
	nameBuilder.WriteString(strconv.FormatInt(entityNameCounter(), 10))

	return nameBuilder.String()
}

// entityNameCounters holds the instance counters for embedded loggers.
var entityNameCounters = sync.Map{}
