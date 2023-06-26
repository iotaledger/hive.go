//go:build !stacktrace

package ierrors

import (
	"fmt"
)

// Wrapf annotates an error with a message.
func Wrap(err error, message string) error {
	return fmt.Errorf("%w: %s", err, message)
}

// Wrapf annotates an error with a message format specifier and arguments.
func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}
