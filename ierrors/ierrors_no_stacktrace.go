//go:build !stacktrace

package ierrors

import (
	"fmt"
)

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

// Wrap annotates an error with a message.
// Wrap adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrap(err error, message string) error {
	return fmt.Errorf("%w: %s", err, message)
}

// Wrapf annotates an error with a message format specifier and arguments.
// Wrapf adds a stacktrace to the error if there was no stacktrace
// in the error tree yet and if the build flag "stacktrace" is set.
func Wrapf(err error, format string, args ...interface{}) error {
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
