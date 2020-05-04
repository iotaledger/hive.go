package logger

import "go.uber.org/zap/zapcore"

const (
	ViperKeyLevel             = "logger.level"
	ViperKeyDisableCaller     = "logger.disableCaller"
	ViperKeyDisableStacktrace = "logger.disableStacktrace"
	ViperKeyEncoding          = "logger.encoding"
	ViperKeyOutputPaths       = "logger.outputPaths"
	ViperKeyDisableEvents     = "logger.disableEvents"
)

// Config holds the settings to configure a root logger instance.
type Config struct {
	// Level is the minimum enabled logging level.
	// The default is "info".
	Level string `mapstructure:"level"`
	// DisableCaller stops annotating logs with the calling function's file name and line number.
	// By default, all logs are annotated.
	DisableCaller bool `mapstructure:"disableCaller"`
	// DisableStacktrace disables automatic stacktrace capturing.
	// By default, stacktraces are captured for LevelError and above in production.
	DisableStacktrace bool `mapstructure:"disableStacktrace"`
	// Encoding sets the logger's encoding. Valid values are "json" and "console".
	// The default is "console".
	Encoding string `mapstructure:"encoding"`
	// OutputPaths is a list of URLs, file paths or stdout/stderr to write logging output to.
	// The default is ["stdout"].
	OutputPaths []string `mapstructure:"outputPaths"`
	// DisableEvents prevents log messages from being raced as events.
	// By default, the corresponding events are triggered.
	DisableEvents bool `mapstructure:"disableEvents"`
}

var defaultCfg = Config{
	Level:         "info",
	Encoding:      "console",
	OutputPaths:   []string{"stdout"},
	DisableEvents: true,
}

var defaultEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	EncodeLevel:    zapcore.CapitalLevelEncoder,    // level in upper case
	EncodeTime:     zapcore.RFC3339TimeEncoder,     // timestamp according to RFC3339
	EncodeDuration: zapcore.SecondsDurationEncoder, // duration in seconds
	EncodeCaller:   zapcore.ShortCallerEncoder,     // caller according to package/file:line
}
