package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// The Logger uses the sugared logger.
type Logger = zap.SugaredLogger

// A Level is a logging priority. Higher levels are more important.
type Level = zapcore.Level

const (
	// LevelDebug logs are typically voluminous, and are usually disabled in production.
	LevelDebug = zapcore.DebugLevel
	// LevelInfo is the default logging priority.
	LevelInfo = zapcore.InfoLevel
	// LevelWarn logs are more important than Info, but don't need individual human review.
	LevelWarn = zapcore.WarnLevel
	// LevelError logs are high-priority.
	// If an application is running as expected, there shouldn't be any error-level logs.
	LevelError = zapcore.ErrorLevel
	// LevelPanic logs a message, then panics.
	LevelPanic = zapcore.PanicLevel
	// LevelFatal logs a message, then calls os.Exit(1).
	LevelFatal = zapcore.FatalLevel
)

// ErrGlobalLoggerAlreadyInitialized is returned when InitGlobalLogger is called more than once.
var ErrGlobalLoggerAlreadyInitialized = errors.New("global logger already initialized")

var (
	level = zap.NewAtomicLevel()

	globalLogger            *Logger
	globalLoggerLock        sync.Mutex  // prevents multiple initializations at the same time
	globalLoggerInitialized atomic.Bool // true, if the global logger was successfully initialized
)

// SetGlobalLogger sets the provided logger as the global logger.
func SetGlobalLogger(root *Logger) error {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()

	if globalLoggerInitialized.Load() {
		return ErrGlobalLoggerAlreadyInitialized
	}
	globalLogger = root
	globalLoggerInitialized.Store(true)

	return nil
}

func getEncoderConfig(cfg Config) (zapcore.EncoderConfig, error) {
	// create a deep copy of all basic types (the func pointers are also fine)
	encoderConfig := defaultEncoderConfig

	if cfg.EncodingConfig.EncodeTime != "" {
		var timeEncoder zapcore.TimeEncoder
		if err := timeEncoder.UnmarshalText([]byte(cfg.EncodingConfig.EncodeTime)); err != nil {
			return zapcore.EncoderConfig{}, err
		}
		encoderConfig.EncodeTime = timeEncoder
	}

	return encoderConfig, nil
}

// NewRootLogger creates a new root logger from the provided configuration.
func NewRootLogger(cfg Config) (*Logger, error) {
	var (
		cores []zapcore.Core
		opts  []zap.Option
	)

	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	encoderConfig, err := getEncoderConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load encoder config: %w", err)
	}

	enc, err := newEncoder(cfg.Encoding, encoderConfig)
	if err != nil {
		return nil, err
	}

	var enabler zapcore.LevelEnabler = level

	// write errors generated by the logger to stderr
	opts = append(opts, zap.ErrorOutput(zapcore.Lock(os.Stderr)))

	// create the logger only if there is at least one output path
	if len(cfg.OutputPaths) > 0 {
		ws, _, err := zap.Open(cfg.OutputPaths...)
		if err != nil {
			return nil, err
		}

		core := zapcore.NewCore(enc, ws, enabler)
		cores = append(cores, core)

		// add required options
		opts = append(opts, buildOptions(cfg)...)
	}

	// add the event logging
	if !cfg.DisableEvents {
		cores = append(cores, NewEventCore(enabler))
	}

	// create the logger
	logger := zap.New(zapcore.NewTee(cores...), opts...)

	return logger.Sugar(), nil
}

// SetLevel alters the logging level of the global logger.
func SetLevel(l Level) {
	level.SetLevel(l)
}

// NewLogger returns a new named child of the global root logger.
func NewLogger(name string) *Logger {
	if !globalLoggerInitialized.Load() {
		panic("global logger not initialized")
	}

	return globalLogger.Named(name)
}

// NewExampleLogger builds a Logger that's designed to be only used in tests or examples.
// It writes debug and above logs to standard out as JSON, but omits the timestamp and calling function to keep
// example output short and deterministic.
func NewExampleLogger(name string) *Logger {
	root := zap.NewExample()

	return root.Named(name).Sugar()
}

// NewNopLogger returns a no-op Logger.
// It never writes out logs or internal errors.
func NewNopLogger() *Logger {
	return zap.NewNop().Sugar()
}

func newEncoder(name string, cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	switch strings.ToLower(name) {
	case "console", "":
		return zapcore.NewConsoleEncoder(cfg), nil
	case "json":
		return zapcore.NewJSONEncoder(cfg), nil
	}

	return nil, fmt.Errorf("no encoder registered for name %q", name)
}

func buildOptions(cfg Config) []zap.Option {
	var opts []zap.Option

	if !cfg.DisableCaller {
		// add caller to the log
		opts = append(opts, zap.AddCaller())
	}
	if !cfg.DisableStacktrace {

		var stacktraceLevel Level
		if err := stacktraceLevel.UnmarshalText([]byte(cfg.StacktraceLevel)); err != nil {
			stacktraceLevel = LevelPanic
		}

		opts = append(opts, zap.AddStacktrace(stacktraceLevel))
	}

	return opts
}
