//go:build stacktrace
// +build stacktrace

package ierrors

import (
	"errors"
	"fmt"
	"runtime"
)

func stacktrace() string {
	var stack string

	var programCounter [32]uintptr
	entries := runtime.Callers(4, programCounter[:])
	frames := runtime.CallersFrames(programCounter[:entries])
	for {
		frame, more := frames.Next()
		if (frame == runtime.Frame{}) {
			break
		}

		stack = fmt.Sprintf("%s%s\n\t%s:%d\n", stack, frame.Function, frame.File, frame.Line)
		if !more {
			// remove last newline
			stack = stack[:len(stack)-1]
			break
		}
	}

	return stack
}

// errorWithStacktrace is an implementation of an error with a stacktrace.
type errorWithStacktrace struct {
	err        error
	stacktrace string
}

func (e *errorWithStacktrace) Error() string {
	return fmt.Sprintf("%s\n%s", e.err.Error(), e.stacktrace)
}

func (e *errorWithStacktrace) Unwrap() error {
	return e.err
}

func newErrorWithStacktrace(err error, stacktrace string) *errorWithStacktrace {
	return &errorWithStacktrace{
		err:        err,
		stacktrace: stacktrace,
	}
}

// ensureStacktraceUniqueness checks if the given error
// already contains a stacktrace in it's error tree.
// if yes, it simply returns the err.
// if not, it returns a new error with a stacktrace appended.
func ensureStacktraceUniqueness(err error) error {
	var errWithStacktrace *errorWithStacktrace
	if errors.As(err, &errWithStacktrace) {
		return err
	}

	return newErrorWithStacktrace(err, stacktrace())
}

// Errorf formats according to a format specifier and returns the string as a
// value that satisfies error.
//
// If the format specifier includes a %w verb with an error operand,
// the returned error will implement an Unwrap method returning the operand.
// If there is more than one %w verb, the returned error will implement an
// Unwrap method returning a []error containing all the %w operands in the
// order they appear in the arguments.
// It is invalid to supply the %w verb with an operand that does not implement
// the error interface. The %w verb is otherwise a synonym for %v.
// Errorf adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Errorf(format string, args ...any) error {
	// check if the error tree already contains an error with a stacktrace
	return ensureStacktraceUniqueness(fmt.Errorf(format, args...))
}

// Wrap annotates an error with a message.
// Wrap adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrap(err error, message string) error {
	return ensureStacktraceUniqueness(fmt.Errorf("%s: %w", message, err))
}

// Wrapf annotates an error with a message format specifier and arguments.
// Wrapf adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrapf(err error, format string, args ...interface{}) error {
	// check if the passed args also contain an error
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			// wrap the other errors as well
			return ensureStacktraceUniqueness(fmt.Errorf("%w: %w", fmt.Errorf(format, args...), err))
		}
	}

	return ensureStacktraceUniqueness(fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err))
}

// WithStack adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithStack(err error) error {
	return ensureStacktraceUniqueness(err)
}
