package logger

import (
	"io/ioutil"
	"os"
	"testing"

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

	// override the default config to write to temp file
	defaultCfg.OutputPaths = append(defaultCfg.OutputPaths, temp.Name())

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

func getLogs(t require.TestingT, file *os.File) string {
	byteContents, err := ioutil.ReadAll(file)
	require.NoError(t, err, "Couldn't read log contents from file.")
	return string(byteContents)
}
