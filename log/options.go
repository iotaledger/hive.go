package log

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/iotaledger/hive.go/runtime/options"
)

// Options contains the configuration options for a Logger.
type Options struct {
	// Name is the name of the logger.
	Name string

	// Level is the log level of the logger.
	Level Level

	// TimeFormat is the time format of the logger.
	TimeFormat string

	// Output is the output of the logger.
	Output io.Writer

	// Handler is the handler of the logger.
	Handler slog.Handler
}

// newOptions creates an Options object that is derived from the given options overriding the default values.
func newOptions(opts ...options.Option[Options]) *Options {
	o := options.Apply(&Options{
		Name:       "",
		Level:      LevelInfo,
		TimeFormat: time.DateTime,
		Output:     os.Stdout,
	}, opts)

	if o.Handler == nil {
		o.Handler = NewTextHandler(o)
	}

	return o
}

// WithName is an option to set the name of a Logger.
func WithName(name string) options.Option[Options] {
	return func(opts *Options) {
		opts.Name = name
	}
}

// WithLevel is an option to set the log level of a Logger.
func WithLevel(level Level) options.Option[Options] {
	return func(opts *Options) {
		opts.Level = level
	}
}

// WithTimeFormat is an option to set the time format of a Logger.
func WithTimeFormat(timeFormat string) options.Option[Options] {
	return func(opts *Options) {
		opts.TimeFormat = timeFormat
	}
}

// WithOutput is an option to set the output of a Logger.
func WithOutput(output io.Writer) options.Option[Options] {
	return func(opts *Options) {
		opts.Output = output
	}
}

// WithHandler is an option to set the handler of a Logger.
func WithHandler(handler slog.Handler) options.Option[Options] {
	return func(opts *Options) {
		opts.Handler = handler
	}
}
