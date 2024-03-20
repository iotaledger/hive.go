//go:build !stacktrace

//nolint:goerr113
package ierrors

import (
	"errors"
	"fmt"
)

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
// Join adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Join(errs ...error) error {
	return errors.Join(errs...)
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

	return result
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
	return fmt.Errorf(format, args...)
}

// Wrap prepends an error with a message and wraps it into a new error.
// Wrap adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf prepends an error with a message format specifier and arguments
// and wraps it into a new error.
// Wrapf adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrapf(err error, format string, args ...interface{}) error {
	// check if the passed args also contain an error
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			return fmt.Errorf("%w: %w", fmt.Errorf(format, args...), err)
		}
	}

	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// WithMessage appends a message to the error and wraps it into a new error.
// WithMessage adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithMessage(err error, message string) error {
	return fmt.Errorf("%w: %s", err, message)
}

// WithMessagef appends a message format specifier and arguments to the error
// and wraps it into a new error.
// WithMessagef adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithMessagef(err error, format string, args ...interface{}) error {
	// check if the passed args also contain an error
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			return fmt.Errorf("%w: %w", err, fmt.Errorf(format, args...))
		}
	}

	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

// WithStack adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func WithStack(err error) error {
	return err
}
