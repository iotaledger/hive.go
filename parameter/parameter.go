package parameter

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
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
			ignoredSettingSplitted := strings.Split(strings.ToLower(ignoredSetting), ".")
			for lvl, parameterName := range ignoredSettingSplitted {
				if lvl == len(ignoredSettingSplitted)-1 {
					delete(parameter, parameterName)
					continue
				}

				par, exists := parameter[parameterName]
				if !exists {
					// parameter not found in settings
					break
				}

				parameter = par.(map[string]interface{})
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
// and ending with: .json, .toml, .yaml or .yml (in this sequence) or the file extension in configName if available.
func LoadConfigFile(config *viper.Viper, configDir string, configName string, bindFlagSet *flag.FlagSet, loadDefault bool, dontParseFlags ...bool) error {
	if len(dontParseFlags) == 0 || !dontParseFlags[0] {
		flag.Parse()
	}

	if bindFlagSet != nil {
		err := config.BindPFlags(bindFlagSet)
		if err != nil {
			return err
		}
	}

	// We want to check whether the extension is valid and supported.
	// When a file called config.default.json exists and we get config.default passed that is not a valid extension
	// and we don't want to try to read a file with that name.
	supportedExtension := isExtensionSupported(filepath.Ext(configName))

	// adjust viper to wanted locations
	if supportedExtension {
		// If the file has an valid extension use the file directly
		file := filepath.Join(configDir, configName)
		config.SetConfigFile(file)
	} else {
		// No valid extension available, set the base config name and viper will try .json, .yaml, etc.
		config.SetConfigName(configName)
		config.AddConfigPath(configDir)
	}

	// read in the config file (whatever it finds)
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && loadDefault {
			if supportedExtension {
				log.Printf("No config file found via '%s/%s'. Loading default settings.", configDir, configName)
			} else {
				log.Printf("No config file found via '%s/%s.[json,toml,yaml,yml]'. Loading default settings.", configDir, configName)
			}
		} else {
			return err
		}
	}

	return nil
}

// isExtensionSupported checks whether the passed extension is supported by viper.
// Extension passed must have a preceding dot, e.g. ".json"
func isExtensionSupported(extension string) bool {
	if len(extension) == 0 {
		// Empty string => no extension available
		return false
	}

	// The extension is passed with the dot in front which we don't want for checking
	extensionWithoutDot := extension[1:]

	for _, ext := range viper.SupportedExts {
		if ext == extensionWithoutDot {
			return true
		}
	}
	return false
}
