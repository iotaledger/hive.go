package parameter

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	defaultConfig     *viper.Viper
	defaultConfigInit sync.Once
)

// PrintConfig prints the actual configuration, ignoreSettingsAtPrint are not shown
func PrintConfig(config *viper.Viper, ignoreSettingsAtPrint ...[]string) {
	settings := config.AllSettings()
	if len(ignoreSettingsAtPrint) > 0 {
		for _, ignoredSetting := range ignoreSettingsAtPrint[0] {
			parameter := settings
			ignoredSettingSplitted := strings.Split(ignoredSetting, ".")
			for lvl, parameterName := range ignoredSettingSplitted {
				if lvl == len(ignoredSettingSplitted)-1 {
					delete(parameter, parameterName)
					continue
				}
				parameter = parameter[parameterName].(map[string]interface{})
			}
		}
	}

	if cfg, err := json.MarshalIndent(settings, "", "  "); err == nil {
		fmt.Printf("Parameters loaded: \n %+v\n", string(cfg))
	}
}

func DefaultConfig() *viper.Viper {
	defaultConfigInit.Do(func() {
		configName := *flag.StringP("config", "c", "config", "Filename of the config file without the file extension")
		configDirPath := *flag.StringP("config-dir", "d", ".", "Path to the directory containing the config file")

		defaultConfig = viper.New()
		err := LoadConfigFile(defaultConfig, configDirPath, configName, true, true)
		if err != nil {
			log.Panicf("Error loading config: %s", err)
		}
	})

	return defaultConfig
}

// LoadConfigFile fetches config values from a dir defined in "configDir" (or the current working dir if not set)
// into a given viper instance.
//
// It automatically reads in a single config with name defined in "configName"
// and ending with: .json, .toml, .yaml or .yml (in this sequence).
func LoadConfigFile(config *viper.Viper, configDir string, configName string, bindFlags bool, loadDefault bool) error {
	flag.Parse()
	if bindFlags {
		err := config.BindPFlags(flag.CommandLine)
		if err != nil {
			return err
		}
	}

	// adjust viper to wanted locations
	config.SetConfigName(configName)
	config.AddConfigPath(configDir)

	// read in the config file (whatever it finds)
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && loadDefault {
			log.Printf("No config file found via '%s/%s.[json,toml,yaml,yml]'. Loading default settings.", configDir, configName)
		} else {
			return err
		}
	}

	return nil
}
