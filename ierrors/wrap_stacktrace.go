//go:build stacktrace
// +build stacktrace

package ierrors

import (
	"fmt"
	"runtime"
)

func stacktrace() string {
	var result string

	var programCounter [32]uintptr
	entries := runtime.Callers(3, programCounter[:])
	frames := runtime.CallersFrames(programCounter[:entries])
	for {
		frame, more := frames.Next()
		if (frame == runtime.Frame{}) {
			break
		}

		result = fmt.Sprintf("%s%s\n\t%s:%d\n", result, frame.Function, frame.File, frame.Line)
		if !more {
			// remove last newline
			result = result[:len(result)-1]
			break
		}
	}

	return result
}

// Wrap annotates an error with a message and a stacktrace.
func Wrap(err error, message string) error {
	return fmt.Errorf("%w: %s\n%s", err, message, stacktrace())
}

// Wrapf annotates an error with a message format specifier, arguments and a stacktrace.
func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s\n%s", err, fmt.Sprintf(format, args...), stacktrace())
}
