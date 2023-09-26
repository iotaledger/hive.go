package app

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/izuc/zipp.foundation/app/configuration"
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
			return fmt.Errorf("loading %s config file failed: config path flag not found", configSet.configName)
		}

		if configSet.loadOnlyIfFlagDefined {
			if configuration.HasFlag(flag.CommandLine, configSet.filePathFlagName) {
				// config file is only loaded if the flag was specified
				if err := configSet.config.LoadFile(configPathFlag.Value.String()); err != nil {
					return fmt.Errorf("loading %s config file failed: %w", configSet.configName, err)
				}
			}
		} else {
			if err := configSet.config.LoadFile(configPathFlag.Value.String()); err != nil {
				if configuration.HasFlag(flag.CommandLine, configSet.filePathFlagName) || !os.IsNotExist(err) {
					// if a file was explicitly specified or the default file exists but couldn't be parsed, raise the error
					return fmt.Errorf("loading %s config file failed: %w", configSet.configName, err)
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
