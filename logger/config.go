package logger

import "go.uber.org/zap/zapcore"

const (
	ConfigurationKeyLevel             = "logger.level"
	ConfigurationKeyDisableCaller     = "logger.disableCaller"
	ConfigurationKeyDisableStacktrace = "logger.disableStacktrace"
	ConfigurationKeyStacktraceLevel   = "logger.stacktraceLevel"
	ConfigurationKeyEncoding          = "logger.encoding"
	ConfigurationKeyOutputPaths       = "logger.outputPaths"
	ConfigurationKeyDisableEvents     = "logger.disableEvents"
)

// Config holds the settings to configure a root logger instance.
type Config struct {
	// Level is the minimum enabled logging level.
	// The default is "info".
	Level string `json:"level"`
	// DisableCaller stops annotating logs with the calling function's file name and line number.
	// By default, logs are not annotated.
	DisableCaller bool `json:"disableCaller"`
	// DisableStacktrace disables automatic stacktrace capturing.
	DisableStacktrace bool `json:"disableStacktrace"`
	// StacktraceLevel is the level stacktraces are captured and above.
	// The default is "panic".
	StacktraceLevel string `json:"stacktraceLevel"`
	// Encoding sets the logger's encoding. Valid values are "json" and "console".
	// The default is "console".
	Encoding string `json:"encoding"`
	// OutputPaths is a list of URLs, file paths or stdout/stderr to write logging output to.
	// The default is ["stdout"].
	OutputPaths []string `json:"outputPaths"`
	// DisableEvents prevents log messages from being raced as events.
	// By default, the corresponding events are not triggered.
	DisableEvents bool `json:"disableEvents"`
}

var defaultCfg = Config{
	Level:             "info",
	DisableCaller:     true,
	DisableStacktrace: false,
	StacktraceLevel:   "panic",
	Encoding:          "console",
	OutputPaths:       []string{"stdout"},
	DisableEvents:     true,
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
