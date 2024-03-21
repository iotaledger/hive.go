package app

import (
	"fmt"
	"io"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/log"
)

type ConfigurationSets []*ConfigurationSet

func (c ConfigurationSets) FlagSetsMap() map[string]*flag.FlagSet {
	flagsets := make(map[string]*flag.FlagSet)
	for _, config := range c {
		flagsets[config.flagSetName] = config.flagSet
	}

	return flagsets
}

func (c ConfigurationSets) FlagSets() []*flag.FlagSet {
	flagsets := make([]*flag.FlagSet, len(c))
	for i, config := range c {
		flagsets[i] = config.flagSet
	}

	return flagsets
}

func (c ConfigurationSets) ConfigsMap() map[string]*configuration.Configuration {
	configs := make(map[string]*configuration.Configuration)
	for _, config := range c {
		configs[config.flagSetName] = config.config
	}

	return configs
}

type ConfigurationSet struct {
	flagSet                 *flag.FlagSet
	config                  *configuration.Configuration
	configName              string
	flagSetName             string
	filePathFlagName        string
	filePathFlagProvideName string
	loadOnlyIfFlagDefined   bool
	loadFlagSet             bool
	loadEnvVars             bool
	defaultConfigPath       string
	shortHand               string
}

func NewConfigurationSet(
	configName string,
	filePathFlagName string,
	filePathFlagProvideName string,
	flagSetName string,
	loadOnlyIfFlagDefined bool,
	loadFlagSet bool,
	loadEnvVars bool,
	defaultConfigPath string,
	shortHand string) *ConfigurationSet {
	return &ConfigurationSet{
		flagSet:                 configuration.NewUnsortedFlagSet(flagSetName, flag.ContinueOnError),
		config:                  configuration.New(),
		configName:              configName,
		flagSetName:             flagSetName,
		filePathFlagName:        filePathFlagName,
		filePathFlagProvideName: filePathFlagProvideName,
		loadOnlyIfFlagDefined:   loadOnlyIfFlagDefined,
		loadFlagSet:             loadFlagSet,
		loadEnvVars:             loadEnvVars,
		defaultConfigPath:       defaultConfigPath,
		shortHand:               shortHand,
	}
}

// loadConfigurations loads the configuration files, the flag sets and the environment variables.
func loadConfigurations(configFilesFlagSet *flag.FlagSet, configurationSets []*ConfigurationSet) error {
	for _, configSet := range configurationSets {
		configPathFlag := configFilesFlagSet.Lookup(configSet.filePathFlagName)
		if configPathFlag == nil {
			return ierrors.Errorf("loading %s config file failed: config path flag not found", configSet.configName)
		}

		if configSet.loadOnlyIfFlagDefined {
			if configuration.HasFlag(flag.CommandLine, configSet.filePathFlagName) {
				// config file is only loaded if the flag was specified
				if err := configSet.config.LoadFile(configPathFlag.Value.String()); err != nil {
					return ierrors.Wrapf(err, "loading %s config file failed", configSet.configName)
				}
			}
		} else {
			if err := configSet.config.LoadFile(configPathFlag.Value.String()); err != nil {
				if configuration.HasFlag(flag.CommandLine, configSet.filePathFlagName) || !os.IsNotExist(err) {
					// if a file was explicitly specified or the default file exists but couldn't be parsed, raise the error
					return ierrors.Wrapf(err, "loading %s config file failed", configSet.configName)
				}
				fmt.Printf("No %s config file found via '%s'. Loading default settings.\n", configSet.configName, configPathFlag.Value.String())
			}
		}
	}

	for _, config := range configurationSets {
		if config.loadFlagSet {
			// load the flags to set the default values
			if err := config.config.LoadFlagSet(config.flagSet); err != nil {
				return err
			}
		}
	}

	for _, config := range configurationSets {
		if config.loadEnvVars {
			// load the env vars after default values from flags were set (otherwise the env vars are not added because the keys don't exist)
			if err := config.config.LoadEnvironmentVars(""); err != nil {
				return err
			}
		}
	}

	for _, config := range configurationSets {
		if config.loadFlagSet {
			// load the flags again to overwrite env vars that were also set via command line
			if err := config.config.LoadFlagSet(config.flagSet); err != nil {
				return err
			}
		}
	}

	for _, config := range configurationSets {
		// propagate values in the config back to bound parameters
		config.config.UpdateBoundParameters()
	}

	return nil
}

// LoggerConfig holds the settings to configure a logger instance.
type LoggerConfig struct {
	// Name is the optional name of the logger instance. All log messages are prefixed with that name.
	Name string `default:"" json:"name" usage:"the optional name of the logger instance. All log messages are prefixed with that name."`
	// Level is the minimum enabled logging level.
	// The default is "info".
	Level string `default:"info" json:"level" usage:"the minimum enabled logging level"`
	// TimeFormat sets the logger's timestamp format. Valid values are "layout", "ansic", "unixdate", "rubydate",
	// "rfc822", "rfc822z", "rfc850", "rfc1123", "rfc1123z", "rfc3339", "rfc3339nano", "kitchen", "stamp", "stampmilli",
	// "stampmicro", "stampnano", "datetime", "dateonly", "timeonly" and "iso8601".
	// The default is "rfc3339".
	TimeFormat string `name:"timeFormat" json:"timeFormat" default:"rfc3339" usage:"sets the logger's timestamp format. (options: \"rfc3339\", \"rfc3339nano\", \"datetime\", \"timeonly\", and \"iso8601\")"`
	// OutputPaths is a list of URLs, file paths or stdout/stderr to write logging output to.
	// The default is ["stdout"].
	OutputPaths []string `default:"stdout" json:"outputPaths" usage:"a list of file paths or stdout/stderr to write logging output to"`
}

func NewLoggerFromConfig(cfg *LoggerConfig) (log.Logger, error) {
	level, err := log.LevelFromString(cfg.Level)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to load log level")
	}

	timeFormat, err := getTimeFormat(cfg.TimeFormat)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to load time format")
	}

	outputs := make([]io.Writer, len(cfg.OutputPaths))
	for i, outputPath := range cfg.OutputPaths {
		switch outputPath {
		case "stdout":
			outputs[i] = os.Stdout
		case "stderr":
			outputs[i] = os.Stderr
		default:
			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, ierrors.Wrapf(err, "failed to open log file \"%s\"", outputPath)
			}

			outputs[i] = file
		}
	}

	return log.NewLogger(
		log.WithName(cfg.Name),
		log.WithLevel(level),
		log.WithTimeFormat(timeFormat),
		log.WithOutput(io.MultiWriter(outputs...)),
	), nil
}

func getTimeFormat(format string) (string, error) {
	switch format {
	case "layout":
		return time.Layout, nil
	case "ansic":
		return time.ANSIC, nil
	case "unixdate":
		return time.UnixDate, nil
	case "rubydate":
		return time.RubyDate, nil
	case "rfc822":
		return time.RFC822, nil
	case "rfc822z":
		return time.RFC822Z, nil
	case "rfc850":
		return time.RFC850, nil
	case "rfc1123":
		return time.RFC1123, nil
	case "rfc1123z":
		return time.RFC1123Z, nil
	case "rfc3339":
		return time.RFC3339, nil
	case "rfc3339nano":
		return time.RFC3339Nano, nil
	case "kitchen":
		return time.Kitchen, nil
	case "stamp":
		return time.Stamp, nil
	case "stampmilli":
		return time.StampMilli, nil
	case "stampmicro":
		return time.StampMicro, nil
	case "stampnano":
		return time.StampNano, nil
	case "datetime":
		return time.DateTime, nil
	case "dateonly":
		return time.DateOnly, nil
	case "timeonly":
		return time.TimeOnly, nil
	case "iso8601":
		return "2006-01-02T15:04:05.000Z0700", nil
	case "":
		return time.RFC3339, nil
	default:
		return "", ierrors.Errorf("unknown time format \"%s\"", format)
	}
}
