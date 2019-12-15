package parameter

import (
	"encoding/json"
	"strings"

	"github.com/iotaledger/hive.go/logger"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// flags
	configName    = flag.StringP("config", "c", "config", "Filename of the config file without the file extension")
	configDirPath = flag.StringP("config-dir", "d", ".", "Path to the directory containing the config file")
	// Viper
	NodeConfig = viper.New()
	log        = logger.NewLogger("NodeConfig")
)

// FetchConfig fetches config values from a dir defined via CLI flag --config-dir (or the current working dir if not set).
//
// It automatically reads in a single config file starting with "config" (can be changed via the --config CLI flag)
// and ending with: .json, .toml, .yaml or .yml (in this sequence).
func FetchConfig(printConfig bool, ignoreSettingsAtPrint ...[]string) error {
	err := NodeConfig.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Error(err)
	}

	flag.Parse()

	// adjust viper to wanted locations
	NodeConfig.SetConfigName(*configName)
	NodeConfig.AddConfigPath(*configDirPath)
	log.Infof("Loading parameters from config dir '%s' using '%s' as file prefix...\n", *configDirPath, *configName)

	// read in the config file (whatever it finds)
	if err := NodeConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Errorf("No config file found via '%s/%s.[json,toml,yaml,yml]'. Loading default settings.", *configDirPath, *configName)
		} else {
			log.Panicf("Error while loading config from %s: %s", *configDirPath, err)
		}
	} else {
		log.Infof("read parameters from %s", NodeConfig.ConfigFileUsed())
	}

	// Print parameters if printConfig is true
	if printConfig {
		settings := NodeConfig.AllSettings()
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

		cfg, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			log.Errorf("Error: %s\n", err)
		}
		log.Infof("Parameters loaded: \n %+v\n", string(cfg))
	}

	return nil
}

var plugins = make(map[string]int)

func AddPlugin(name string, status int) {
	if _, exists := plugins[name]; exists {
		panic("duplicate plugin - \"" + name + "\" was defined already")
	}

	plugins[name] = status

	Events.AddPlugin.Trigger(name, status)
}

func GetPlugins() map[string]int {
	return plugins
}
