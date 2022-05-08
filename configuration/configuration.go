package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/iotaledger/hive.go/ioutils"
	reflectutils "github.com/iotaledger/hive.go/reflect"
)

var (
	// ErrConfigDoesNotExist is returned if the config file is unknown.
	ErrConfigDoesNotExist = errors.New("config does not exist")
	// ErrUnknownConfigFormat is returned if the format of the config file is unknown.
	ErrUnknownConfigFormat = errors.New("unknown config file format")
)

// Configuration holds config parameters from several sources (file, env vars, flags).
type Configuration struct {
	config *koanf.Koanf
	// boundParameters keeps track of all parameters that were bound using the BindParameters function.
	boundParameters map[string]*BoundParameter
}

// New returns a new configuration.
func New() *Configuration {
	return &Configuration{
		config:          koanf.New("."),
		boundParameters: make(map[string]*BoundParameter),
	}
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

// LoadFile loads parameters from a JSON or YAML file and merges them into the loaded config.
// Existing keys will be overwritten.
func (c *Configuration) LoadFile(filePath string) error {

	exists, err := ioutils.PathExists(filePath)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrNotExist
	}

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

// StoreFile stores the current config to a JSON or YAML file.
// ignoreSettingsAtStore will not be stored to the file.
func (c *Configuration) StoreFile(filePath string, ignoreSettingsAtStore ...[]string) error {

	settings := c.config.Raw()
	if len(ignoreSettingsAtStore) > 0 {
		for _, ignoredSetting := range ignoreSettingsAtStore[0] {
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

	var parser koanf.Parser

	switch filepath.Ext(filePath) {
	case ".json":
		parser = &JSONLowerParser{
			prefix: "",
			indent: "  ",
		}
	case ".yaml", ".yml":
		parser = &YAMLLowerParser{}
	default:
		return ErrUnknownConfigFormat
	}

	data, err := parser.Marshal(settings)
	if err != nil {
		return fmt.Errorf("unable to marshal config file: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0666); err != nil {
		return fmt.Errorf("unable to save config file: %w", err)
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

// Koanf returns the underlying Koanf instance.
func (c *Configuration) Koanf() *koanf.Koanf {
	return c.config
}

// Load takes a Provider that either provides a parsed config map[string]interface{}
// in which case pa (Parser) can be nil, or raw bytes to be parsed, where a Parser
// can be provided to parse. Additionally, options can be passed which modify the
// load behavior, such as passing a custom merge function.
func (c *Configuration) Load(p koanf.Provider, pa koanf.Parser, opts ...koanf.Option) error {
	return c.config.Load(p, pa, opts...)
}

// BoundParameter stores the pointer and the type of values that were bound using the BindParameters function.
type BoundParameter struct {
	boundPointer interface{}
	boundType    reflect.Type
}

// BindParameters is a utility function that allows to define and bind a set of parameters in a single step by using a
// struct as the registry and definition for the created configuration parameters. It parses the relevant information
// from the struct using reflection and optionally provided information in the tags of its fields.
//
// The parameter names are determined by the names of the fields in the struct but they can be overridden by providing a
// name tag.
// The default value is determined by the value of the field in the struct but it can be overridden by
// providing a default tag.
// The usage information are determined by the usage tag of the field.
//
// The method supports nested structs which get translates to parameter names in the following way:
// --level1.level2.level3.parameterName
//
// The first level is determined by the package of struct but it can be overridden by providing an optional namespace
// parameter.
func (c *Configuration) BindParameters(namespace string, pointerToStruct interface{}) {
	val := reflect.ValueOf(pointerToStruct).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		name := namespace + "."
		if tagName, exists := typeField.Tag.Lookup("name"); exists {
			name += tagName
		} else {
			name += lowerCamelCase(typeField.Name)
		}

		shortHand, _ := typeField.Tag.Lookup("shorthand")
		usage, _ := typeField.Tag.Lookup("usage")

		switch defaultValue := valueField.Interface().(type) {
		case bool:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.BoolVarP(valueField.Addr().Interface().(*bool), name, shortHand, defaultValue, usage)
		case time.Duration:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if parsedDuration, err := time.ParseDuration(tagDefaultValue); err != nil {
					panic(err)
				} else {
					defaultValue = parsedDuration
				}
			}

			flag.DurationVarP(valueField.Addr().Interface().(*time.Duration), name, shortHand, defaultValue, usage)
		case float32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Float32VarP(valueField.Addr().Interface().(*float32), name, shortHand, defaultValue, usage)
		case float64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Float64VarP(valueField.Addr().Interface().(*float64), name, shortHand, defaultValue, usage)
		case int:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.IntVarP(valueField.Addr().Interface().(*int), name, shortHand, defaultValue, usage)
		case int8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Int8VarP(valueField.Addr().Interface().(*int8), name, shortHand, defaultValue, usage)
		case int16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Int16VarP(valueField.Addr().Interface().(*int16), name, shortHand, defaultValue, usage)
		case int32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Int32VarP(valueField.Addr().Interface().(*int32), name, shortHand, defaultValue, usage)
		case int64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Int64VarP(valueField.Addr().Interface().(*int64), name, shortHand, defaultValue, usage)
		case string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.StringVarP(valueField.Addr().Interface().(*string), name, shortHand, defaultValue, usage)
		case uint:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.UintVarP(valueField.Addr().Interface().(*uint), name, shortHand, defaultValue, usage)
		case uint8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Uint8VarP(valueField.Addr().Interface().(*uint8), name, shortHand, defaultValue, usage)
		case uint16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Uint16VarP(valueField.Addr().Interface().(*uint16), name, shortHand, defaultValue, usage)
		case uint32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Uint32VarP(valueField.Addr().Interface().(*uint32), name, shortHand, defaultValue, usage)
		case uint64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			flag.Uint64VarP(valueField.Addr().Interface().(*uint64), name, shortHand, defaultValue, usage)
		case []string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				defaultValue = strings.Split(tagDefaultValue, ",")
			}

			flag.StringSliceVarP(valueField.Addr().Interface().(*[]string), name, shortHand, defaultValue, usage)
		default:
			c.BindParameters(name, valueField.Addr().Interface())
		}

		c.boundParameters[name] = &BoundParameter{
			boundPointer: valueField.Addr().Interface(),
			boundType:    valueField.Type(),
		}
	}
}

// UpdateBoundParameters updates parameters that were bound using the BindParameters method with the current values in
// the configuration.
func (c *Configuration) UpdateBoundParameters() {
	for parameterName, boundParameter := range c.boundParameters {
		switch boundParameter.boundType {
		case reflectutils.BoolType:
			*(boundParameter.boundPointer.(*bool)) = c.Bool(parameterName)
		case reflectutils.TimeDurationType:
			*(boundParameter.boundPointer.(*time.Duration)) = c.Duration(parameterName)
		case reflectutils.Float32Type:
			*(boundParameter.boundPointer.(*float32)) = float32(c.Float64(parameterName))
		case reflectutils.Float64Type:
			*(boundParameter.boundPointer.(*float64)) = c.Float64(parameterName)
		case reflectutils.IntType:
			*(boundParameter.boundPointer.(*int)) = c.Int(parameterName)
		case reflectutils.Int8Type:
			*(boundParameter.boundPointer.(*int8)) = int8(c.Int(parameterName))
		case reflectutils.Int16Type:
			*(boundParameter.boundPointer.(*int16)) = int16(c.Int(parameterName))
		case reflectutils.Int32Type:
			*(boundParameter.boundPointer.(*int32)) = int32(c.Int(parameterName))
		case reflectutils.Int64Type:
			*(boundParameter.boundPointer.(*int64)) = c.Int64(parameterName)
		case reflectutils.StringType:
			*(boundParameter.boundPointer.(*string)) = c.String(parameterName)
		case reflectutils.UintType:
			*(boundParameter.boundPointer.(*uint)) = uint(c.Int(parameterName))
		case reflectutils.Uint8Type:
			*(boundParameter.boundPointer.(*uint8)) = uint8(c.Int(parameterName))
		case reflectutils.Uint16Type:
			*(boundParameter.boundPointer.(*uint16)) = uint16(c.Int(parameterName))
		case reflectutils.Uint32Type:
			*(boundParameter.boundPointer.(*uint32)) = uint32(c.Int(parameterName))
		case reflectutils.Uint64Type:
			*(boundParameter.boundPointer.(*uint64)) = uint64(c.Int64(parameterName))
		case reflectutils.StringSliceType:
			*(boundParameter.boundPointer.(*[]string)) = c.Strings(parameterName)
		}
	}
}
