package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

var (
	// ErrUnknownConfigFormat is returned if the format of the config file is unknown.
	ErrUnknownConfigFormat = errors.New("unknown config file format")
)

// Configuration holds config parameters from several sources (file, env vars, flags).
type Configuration struct {
	config *koanf.Koanf
}

// New returns a new configuration.
func New() *Configuration {
	return &Configuration{config: koanf.New(".")}
}

// Print prints the actual configuration, ignoreSettingsAtPrint are not shown.
func (c *Configuration) Print(ignoreSettingsAtPrint ...[]string) {
	settings := c.config.Raw()
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

// LoadFile loads paramerers from a JSON or YAML file and merges them into the loaded config.
// Existing keys will be overwritten.
func (c *Configuration) LoadFile(filePath string) error {

	var parser koanf.Parser

	switch filepath.Ext(filePath) {
	case ".json":
		parser = &JSONLowerParser{}
	case ".yaml", ".yml":
		parser = &YAMLLowerParser{}
	default:
		return ErrUnknownConfigFormat
	}

	if err := c.config.Load(file.Provider(filePath), parser); err != nil {
		return err
	}

	return nil
}

// LoadFlagSet loads parameters from a FlagSet (spf13/pflag lib) including
// default values and merges them into the loaded config.
// Existing keys will only be overwritten, if they were set via command line.
// If not given via command line, default values will only be used if they did not exist beforehand.
func (c *Configuration) LoadFlagSet(flagSet *flag.FlagSet) error {

	return c.config.Load(lowerPosflagProvider(flagSet, ".", c.config), nil)
}

// LoadEnvironmentVars loads parameters from env vars and merges them into the loaded config.
// The prefix is used to filter the env vars.
// Only existing keys will be overwritten, all other keys are ignored.
func (c *Configuration) LoadEnvironmentVars(prefix string) error {
	if prefix != "" {
		prefix += "_"
	}

	return c.config.Load(env.Provider(prefix, ".", func(s string) string {
		mapKey := strings.Replace(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".", -1)
		if !c.config.Exists(mapKey) {
			// only accept values from env vars that already exist in the config
			return ""
		}
		return mapKey
	}), nil)
}
