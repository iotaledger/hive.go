package logger

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/configuration"
)

func init() {
	defaultEncoderConfig.TimeKey = "" // no timestamps in tests
}

func TestNewRootLogger(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expectRx string
	}{
		{
			name: "console",
			cfg: Config{
				Level:           "info",
				Encoding:        "console",
				StacktraceLevel: "error",
			},
			expectRx: `INFO\tlogger/logger_test.go:\d+\tinfo\n` +
				`WARN\tlogger/logger_test.go:\d+\twarn\n`,
		},
		{
			name: "json",
			cfg: Config{
				Level:           "info",
				Encoding:        "json",
				StacktraceLevel: "error",
			},
			expectRx: `{"level":"INFO","caller":"logger/logger_test.go:\d+","msg":"info"}\n` +
				`{"level":"WARN","caller":"logger/logger_test.go:\d+","msg":"warn"}`,
		},
		{
			name: "debug",
			cfg: Config{
				Level:           "debug",
				StacktraceLevel: "error",
			},
			expectRx: `DEBUG\tlogger/logger_test.go:\d+\tdebug\n` +
				`INFO\tlogger/logger_test.go:\d+\tinfo\n` +
				`WARN\tlogger/logger_test.go:\d+\twarn\n`,
		},
		{
			name: "noCaller",
			cfg: Config{
				DisableCaller:   true,
				StacktraceLevel: "error",
			},
			expectRx: "INFO\tinfo\n" +
				"WARN\twarn\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp, err := ioutil.TempFile("", "hive-logger-test")
			require.NoError(t, err, "Failed to create temp file.")
			defer os.Remove(temp.Name())

			tt.cfg.OutputPaths = []string{temp.Name()}

			logger, err := NewRootLogger(tt.cfg)
			require.NoError(t, err, "Unexpected error constructing logger.")

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")

			assert.Regexp(t, tt.expectRx, getLogs(t, temp), "Unexpected log output.")
		})
	}
}

func TestNewLogger(t *testing.T) {
	temp, err := ioutil.TempFile("", "hive-logger-test")
	require.NoError(t, err, "Failed to create temp file.")
	defer os.Remove(temp.Name())

	// override the default config to also write to temp file
	cfg := defaultCfg
	cfg.OutputPaths = append(cfg.OutputPaths, temp.Name())

	// init the global logger for that temp file and de-init afterwards
	defer initGlobal(t, cfg)()

	t.Run("info", func(t *testing.T) {
		logger := NewLogger("test")
		logger.Info("info")

		logs := getLogs(t, temp)
		assert.Regexp(t, `info\n`, logs, "Unexpected log output.")
	})

	t.Run("setLevel", func(t *testing.T) {
		logger := NewLogger("test")
		SetLevel(LevelDebug)
		logger.Debug("debug1")
		SetLevel(LevelInfo)
		logger.Debug("debug2")

		logs := getLogs(t, temp)
		assert.Regexp(t, `debug1\n`, logs, "Unexpected log output.")
		assert.NotRegexp(t, `debug2\n`, logs, "Unexpected log output.")
	})
}

func TestNewLoggerWithoutInit(t *testing.T) {
	assert.Panics(t, func() { NewLogger("test") })
}

func TestInitGlobalAfterError(t *testing.T) {
	// create invalid config
	cfg := defaultCfg
	cfg.Level = "invalid"

	c := configuration.New()
	require.NoError(t, c.Set(ConfigurationKeyLevel, cfg.Level))
	require.NoError(t, c.Set(ConfigurationKeyDisableCaller, cfg.DisableCaller))
	require.NoError(t, c.Set(ConfigurationKeyDisableStacktrace, cfg.DisableStacktrace))
	require.NoError(t, c.Set(ConfigurationKeyEncoding, cfg.Encoding))
	require.NoError(t, c.Set(ConfigurationKeyOutputPaths, cfg.OutputPaths))
	require.NoError(t, c.Set(ConfigurationKeyDisableEvents, cfg.DisableEvents))
	require.Error(t, InitGlobalLogger(c))

	initGlobal(t, defaultCfg)()
}

func TestInitGlobalTwice(t *testing.T) {
	c := configuration.New()
	require.NoError(t, c.Set(ConfigurationKeyLevel, defaultCfg.Level))
	require.NoError(t, c.Set(ConfigurationKeyDisableCaller, defaultCfg.DisableCaller))
	require.NoError(t, c.Set(ConfigurationKeyDisableStacktrace, defaultCfg.DisableStacktrace))
	require.NoError(t, c.Set(ConfigurationKeyEncoding, defaultCfg.Encoding))
	require.NoError(t, c.Set(ConfigurationKeyOutputPaths, defaultCfg.OutputPaths))
	require.NoError(t, c.Set(ConfigurationKeyDisableEvents, defaultCfg.DisableEvents))

	require.NoError(t, InitGlobalLogger(c))
	assert.Errorf(t, InitGlobalLogger(c), ErrGlobalLoggerAlreadyInitialized.Error())
}

func initGlobal(t require.TestingT, cfg Config) func() {
	c := configuration.New()
	require.NoError(t, c.Set(ConfigurationKeyLevel, cfg.Level))
	require.NoError(t, c.Set(ConfigurationKeyDisableCaller, cfg.DisableCaller))
	require.NoError(t, c.Set(ConfigurationKeyDisableStacktrace, cfg.DisableStacktrace))
	require.NoError(t, c.Set(ConfigurationKeyEncoding, cfg.Encoding))
	require.NoError(t, c.Set(ConfigurationKeyOutputPaths, cfg.OutputPaths))
	require.NoError(t, c.Set(ConfigurationKeyDisableEvents, cfg.DisableEvents))

	err := InitGlobalLogger(c)
	require.NoError(t, err, "Failed to init global logger.")

	// de-initialize the global logger
	return func() {
		logger = nil
		initialized.UnSet()
		mu = sync.Mutex{}
	}
}

func getLogs(t require.TestingT, file *os.File) string {
	byteContents, err := ioutil.ReadAll(file)
	require.NoError(t, err, "Couldn't read log contents from file.")

	return string(byteContents)
}
