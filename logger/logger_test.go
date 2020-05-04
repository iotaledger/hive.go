package logger

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Level:    "info",
				Encoding: "console",
			},
			expectRx: `INFO\tlogger/logger_test.go:\d+\tinfo\n` +
				`WARN\tlogger/logger_test.go:\d+\twarn\n`,
		},
		{
			name: "json",
			cfg: Config{
				Level:    "info",
				Encoding: "json",
			},
			expectRx: `{"level":"INFO","caller":"logger/logger_test.go:\d+","msg":"info"}\n` +
				`{"level":"WARN","caller":"logger/logger_test.go:\d+","msg":"warn"}`,
		},
		{
			name: "debug",
			cfg: Config{
				Level: "debug",
			},
			expectRx: `DEBUG\tlogger/logger_test.go:\d+\tdebug\n` +
				`INFO\tlogger/logger_test.go:\d+\tinfo\n` +
				`WARN\tlogger/logger_test.go:\d+\twarn\n`,
		},
		{
			name: "noCaller",
			cfg: Config{
				DisableCaller: true,
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

	v := viper.New()
	v.Set(ViperKeyLevel, cfg.Level)
	v.Set(ViperKeyDisableCaller, cfg.DisableCaller)
	v.Set(ViperKeyDisableStacktrace, cfg.DisableStacktrace)
	v.Set(ViperKeyEncoding, cfg.Encoding)
	v.Set(ViperKeyOutputPaths, cfg.OutputPaths)
	v.Set(ViperKeyDisableEvents, cfg.DisableEvents)
	require.Error(t, InitGlobalLogger(v))

	initGlobal(t, defaultCfg)()
}

func TestInitGlobalTwice(t *testing.T) {
	v := viper.New()
	v.Set(ViperKeyLevel, defaultCfg.Level)
	v.Set(ViperKeyDisableCaller, defaultCfg.DisableCaller)
	v.Set(ViperKeyDisableStacktrace, defaultCfg.DisableStacktrace)
	v.Set(ViperKeyEncoding, defaultCfg.Encoding)
	v.Set(ViperKeyOutputPaths, defaultCfg.OutputPaths)
	v.Set(ViperKeyDisableEvents, defaultCfg.DisableEvents)

	require.NoError(t, InitGlobalLogger(v))
	assert.Errorf(t, InitGlobalLogger(v), ErrGlobalLoggerAlreadyInitialized.Error())
}

func initGlobal(t require.TestingT, cfg Config) func() {
	// load into viper
	v := viper.New()
	v.Set(ViperKeyLevel, cfg.Level)
	v.Set(ViperKeyDisableCaller, cfg.DisableCaller)
	v.Set(ViperKeyDisableStacktrace, cfg.DisableStacktrace)
	v.Set(ViperKeyEncoding, cfg.Encoding)
	v.Set(ViperKeyOutputPaths, cfg.OutputPaths)
	v.Set(ViperKeyDisableEvents, cfg.DisableEvents)

	err := InitGlobalLogger(v)
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
