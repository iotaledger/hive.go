package logger

import "github.com/iotaledger/hive.go/ds/reactive"

func NewEntityLogger(parentLogger *Logger, namespace string, shutdownEvent reactive.Event) *Logger {
	embeddedLogger, shutdown := parentLogger.NestedLogger(namespace)
	if embeddedLogger != nil {
		shutdownEvent.OnTrigger(shutdown)
	}

	return embeddedLogger
}

func LogReactiveVariableUpdate[Type any](logFunc func(msg string, args ...any), varName string) func(oldValue, newValue Type) {
	return func(oldValue, newValue Type) {
		logFunc(varName+" updated", "oldValue", oldValue, "newValue", newValue)
	}
}
