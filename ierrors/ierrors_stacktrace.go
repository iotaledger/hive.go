//go:build stacktrace
// +build stacktrace

//
//nolint:goerr113
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
	if err == nil {
		return nil
	}

	var errWithStacktrace *errorWithStacktrace
	if errors.As(err, &errWithStacktrace) {
		return err
	}

	return newErrorWithStacktrace(err, stacktrace())
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
// Join adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Join(errs ...error) error {
	return ensureStacktraceUniqueness(errors.Join(errs...))
}

// Chain chains multiple errors into a single error by wrapping them.
// Any nil error values are discarded.
// Chain returns nil if every value in errs is nil.
// Chain adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Chain(errs ...error) error {
	var result error
	for _, err := range errs {
		if err == nil {
			continue
		}

		if result == nil {
			result = err
			continue
		}

		result = fmt.Errorf("%w: %w", result, err)
	}

	return ensureStacktraceUniqueness(result)
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

// Wrap prepends an error with a message and wraps it into a new error.
// Wrap adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrap(err error, message string) error {
	return ensureStacktraceUniqueness(fmt.Errorf("%s: %w", message, err))
}

// Wrapf prepends an error with a message format specifier and arguments
// and wraps it into a new error.
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

// WithMessage appends a message to the error and wraps it into a new error.
// WithMessage adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithMessage(err error, message string) error {
	return ensureStacktraceUniqueness(fmt.Errorf("%w: %s", err, message))
}

// WithMessagef appends a message format specifier and arguments to the error
// and wraps it into a new error.
// WithMessagef adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithMessagef(err error, format string, args ...interface{}) error {
	// check if the passed args also contain an error
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			// wrap the other errors as well
			return ensureStacktraceUniqueness(fmt.Errorf("%w: %w", err, fmt.Errorf(format, args...)))
		}
	}

	return ensureStacktraceUniqueness(fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...)))
}

// WithStack adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithStack(err error) error {
	return ensureStacktraceUniqueness(err)
}
