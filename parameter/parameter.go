package parameter

import (
	"os"

	"github.com/iotaledger/hive.go/logger"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// Config file type defaults
const (
	FetchDefaultYAML = "yaml"
	FetchDefaultJSON = "json"
	FetchDefaultTOML = "toml"
)

var (
	NodeConfig = viper.New()
	log        = logger.NewLogger("NodeConfig")
)

// FetchConfig fetches parameters from a config file or command line arguments.
// printConfig: Print parameters as YAML
// defautFetch: Try to fetch config from config.(json | yaml | toml) if no config flag has been set, default: config.yaml
func FetchConfig(printConfig bool, defaultFetch string) error {

	var fetchConfig string

	err := NodeConfig.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Error(err)
	}

	switch defaultFetch {
	case FetchDefaultYAML:
		fetchConfig = "config.yaml"
	case FetchDefaultJSON:
		fetchConfig = "config.json"
	case FetchDefaultTOML:
		fetchConfig = "config.toml"
	default:
		fetchConfig = "config.yaml"
	}

	var configPath *string = flag.StringP("config", "c", fetchConfig, "Path to the config file")
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
				log.Panicf("Error while loading config from: %s (%s)\n", *configPath, err)
			}
		}
	}

	// Print parameters if printConfig is true
	if printConfig {
		cfg, err := yaml.Marshal(NodeConfig.AllSettings())
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
