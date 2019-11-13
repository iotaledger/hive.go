package parameter

import (
	"encoding/json"
	"github.com/iotaledger/hive.go/logger"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	NodeConfig = viper.New()
	log        = logger.NewLogger("NodeConfig")
)

// FetchConfig fetches parameters from a config file or command line arguments
func FetchConfig() error {

	err := NodeConfig.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Error(err)
	}

	var configPath *string = flag.StringP("config", "c", "config.json", "Path to the config file")
	flag.Parse()

	if configPath != nil {
		// Check if config file exists
		_, err := os.Stat(*configPath)
		if os.IsNotExist(err) && !flag.CommandLine.Changed("config") {
			log.Error("No config file found. Loading default settings.")
		} else {
			log.Infof("Loading parameters from %s...\n", *configPath)
			NodeConfig.SetConfigFile(*configPath)
			err := NodeConfig.ReadInConfig()
			if err != nil {
				log.Errorf("Error while loading config from: %s (%s)\n", *configPath, err)
			}
		}
	}

	// Print parameters
	cfg, err := json.MarshalIndent(NodeConfig.AllSettings(), "", "  ")
	if err != nil {
		log.Errorf("Error: %s\n", err)
	}
	log.Infof("Parameters loaded: \n %+v\n", string(cfg))

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
