package reactive

func NewEntityLogger(parentLogger *Logger, entityType string, shutdownEvent Event) *Logger {
	embeddedLogger, shutdown := parentLogger.NestedLogger(entityType)
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
