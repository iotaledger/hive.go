package logger

import (
	"github.com/izuc/zipp.foundation/app/configuration"
	"github.com/izuc/zipp.foundation/logger"
)

// InitGlobalLogger initializes the global logger from the provided config.
func InitGlobalLogger(config *configuration.Configuration) error {
	root, err := NewRootLoggerFromConfiguration(config)
	if err != nil {
		return err
	}

	return logger.SetGlobalLogger(root)
}

// NewRootLoggerFromConfiguration creates a new root logger from the provided configuration.
func NewRootLoggerFromConfiguration(config *configuration.Configuration) (*logger.Logger, error) {
	cfg := logger.DefaultCfg

	// get config values one by one
	// config.UnmarshalKey does not recognize a configuration group when defined with pflags
	if val := config.String(logger.ConfigurationKeyLevel); val != "" {
		cfg.Level = val
	}
	if val := config.Get(logger.ConfigurationKeyDisableCaller); val != nil {
		cfg.DisableCaller = val.(bool)
	}
	if val := config.Get(logger.ConfigurationKeyDisableStacktrace); val != nil {
		cfg.DisableStacktrace = val.(bool)
	}
	if val := config.String(logger.ConfigurationKeyStacktraceLevel); val != "" {
		cfg.StacktraceLevel = val
	}
	if val := config.String(logger.ConfigurationKeyEncoding); val != "" {
		cfg.Encoding = val
	}
	if val := config.Strings(logger.ConfigurationKeyOutputPaths); len(val) > 0 {
		cfg.OutputPaths = val
	}
	if val := config.Get(logger.ConfigurationKeyDisableEvents); val != nil {
		cfg.DisableEvents = val.(bool)
	}

	return logger.NewRootLogger(cfg)
}
