package parameter

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
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

// LoadConfigFile fetches config values from a dir defined in "configDir" (or the current working dir if not set)
// into a given viper instance.
//
// It automatically reads in a single config with name defined in "configName"
// and ending with: .json, .toml, .yaml or .yml (in this sequence).
func LoadConfigFile(config *viper.Viper, configDir string, configName string, bindFlags bool, loadDefault bool, dontParseFlags ...bool) error {
	if len(dontParseFlags) == 0 || !dontParseFlags[0] {
		flag.Parse()
	}
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
