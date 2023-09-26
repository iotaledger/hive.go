package logger

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/app/configuration"
	"github.com/izuc/zipp.foundation/logger"
)

func TestInitGlobalAfterError(t *testing.T) {
	// create invalid config
	cfg := logger.DefaultCfg
	cfg.Level = "invalid"

	c := configuration.New()
	require.NoError(t, c.Set(logger.ConfigurationKeyLevel, cfg.Level))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableCaller, cfg.DisableCaller))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableStacktrace, cfg.DisableStacktrace))
	require.NoError(t, c.Set(logger.ConfigurationKeyEncoding, cfg.Encoding))
	require.NoError(t, c.Set(logger.ConfigurationKeyOutputPaths, cfg.OutputPaths))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableEvents, cfg.DisableEvents))
	require.Error(t, InitGlobalLogger(c))

	initGlobal(t, logger.DefaultCfg)
}

func initGlobal(t require.TestingT, cfg logger.Config) {
	c := configuration.New()
	require.NoError(t, c.Set(logger.ConfigurationKeyLevel, cfg.Level))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableCaller, cfg.DisableCaller))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableStacktrace, cfg.DisableStacktrace))
	require.NoError(t, c.Set(logger.ConfigurationKeyEncoding, cfg.Encoding))
	require.NoError(t, c.Set(logger.ConfigurationKeyOutputPaths, cfg.OutputPaths))
	require.NoError(t, c.Set(logger.ConfigurationKeyDisableEvents, cfg.DisableEvents))

	err := InitGlobalLogger(c)
	require.NoError(t, err, "Failed to init global logger.")
}
