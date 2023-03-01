package logger

import "go.uber.org/zap/zapcore"

const (
	ConfigurationKeyLevel             = "logger.level"
	ConfigurationKeyDisableCaller     = "logger.disableCaller"
	ConfigurationKeyDisableStacktrace = "logger.disableStacktrace"
	ConfigurationKeyStacktraceLevel   = "logger.stacktraceLevel"
	ConfigurationKeyEncoding          = "logger.encoding"
	ConfigurationKeyEncodingConfig    = "logger.encodingConfig"
	ConfigurationKeyOutputPaths       = "logger.outputPaths"
	ConfigurationKeyDisableEvents     = "logger.disableEvents"
)

type EncodingConfig struct {
	// EncodeTime sets the logger's timestamp encoding. Valid values are "nanos", "millis", "iso8601", "rfc3339" and "rfc3339nano".
	// The default is "rfc3339".
	EncodeTime string `name:"timeEncoder" default:"rfc3339" usage:"sets the logger's timestamp encoding. (options: \"nanos\", \"millis\", \"iso8601\", \"rfc3339\" and \"rfc3339nano\")" json:"timeEncoder"`
}

// Config holds the settings to configure a root logger instance.
type Config struct {
	// Level is the minimum enabled logging level.
	// The default is "info".
	Level string `default:"info" usage:"the minimum enabled logging level" json:"level"`
	// DisableCaller stops annotating logs with the calling function's file name and line number.
	// By default, logs are not annotated.
	DisableCaller bool `default:"true" usage:"stops annotating logs with the calling function's file name and line number" json:"disableCaller"`
	// DisableStacktrace disables automatic stacktrace capturing.
	DisableStacktrace bool `default:"false" usage:"disables automatic stacktrace capturing" json:"disableStacktrace"`
	// StacktraceLevel is the level stacktraces are captured and above.
	// The default is "panic".
	StacktraceLevel string `default:"panic" usage:"the level stacktraces are captured and above" json:"stacktraceLevel"`
	// Encoding sets the logger's encoding. Valid values are "json" and "console".
	// The default is "console".
	Encoding string `default:"console" usage:"the logger's encoding (options: \"json\", \"console\")" json:"encoding"`
	// EncodingConfig sets the logger's encoding config.
	EncodingConfig EncodingConfig `default:"console" usage:"the logger's encoding config" json:"encodingConfig"`
	// OutputPaths is a list of URLs, file paths or stdout/stderr to write logging output to.
	// The default is ["stdout"].
	OutputPaths []string `default:"stdout" usage:"a list of URLs, file paths or stdout/stderr to write logging output to" json:"outputPaths"`
	// DisableEvents prevents log messages from being raced as events.
	// By default, the corresponding events are not triggered.
	DisableEvents bool `default:"true" usage:"prevents log messages from being raced as events" json:"disableEvents"`
}

var DefaultCfg = Config{
	Level:             "info",
	DisableCaller:     true,
	DisableStacktrace: false,
	StacktraceLevel:   "panic",
	Encoding:          "console",
	EncodingConfig: EncodingConfig{
		EncodeTime: "rfc3339",
	},
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
