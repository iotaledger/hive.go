package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	reflectutils "github.com/iotaledger/hive.go/runtime/reflect"
)

var (
	// ErrConfigDoesNotExist is returned if the config file is unknown.
	ErrConfigDoesNotExist = ierrors.New("config does not exist")
	// ErrUnknownConfigFormat is returned if the format of the config file is unknown.
	ErrUnknownConfigFormat = ierrors.New("unknown config file format")
)

// Configuration holds config parameters from several sources (file, env vars, flags).
type Configuration struct {
	config *koanf.Koanf
	// boundParameters keeps track of all parameters that were bound using the BindParameters function.
	boundParameters map[string]*BoundParameter
	// boundParametersMapping keeps track of all parameter names that were bound using the BindParameters function.
	boundParametersMapping map[reflect.Value]string
}

// New returns a new configuration.
func New() *Configuration {
	return &Configuration{
		config:                 koanf.New("."),
		boundParameters:        make(map[string]*BoundParameter),
		boundParametersMapping: make(map[reflect.Value]string),
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
				//nolint:forcetypeassert // false positive, nested map[string]interface{} is expected
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
	exists, isDir, err := ioutils.PathExists(filePath)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrNotExist
	}
	if isDir {
		return ierrors.Errorf("given path is a directory instead of a file %s", filePath)
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
func (c *Configuration) StoreFile(filePath string, perm os.FileMode, ignoreSettingsAtStore ...[]string) error {
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

				//nolint:forcetypeassert // false positive, nested map[string]interface{} is expected
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
		return ierrors.Wrap(err, "unable to marshal config file")
	}

	if err := os.WriteFile(filePath, data, perm); err != nil {
		return ierrors.Wrap(err, "unable to save config file")
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
		mapKey := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".")
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
	Name         string
	ShortHand    string
	Usage        string
	DefaultVal   any
	BoundPointer any
	BoundType    reflect.Type
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
func (c *Configuration) BindParameters(flagset *flag.FlagSet, namespace string, pointerToStruct interface{}) {
	val := reflect.ValueOf(pointerToStruct).Elem()
	for i := range val.NumField() {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		name := namespace + "."
		if tagName, exists := typeField.Tag.Lookup("name"); exists {
			name += tagName
		} else {
			name += LowerCamelCase(typeField.Name)
		}

		shortHand, _ := typeField.Tag.Lookup("shorthand")
		usage, _ := typeField.Tag.Lookup("usage")

		addParameter := func(name string, shortHand string, usage string, defaultVal any, valueField reflect.Value) {
			c.boundParameters[strings.ToLower(name)] = &BoundParameter{
				Name:         name,
				ShortHand:    shortHand,
				Usage:        usage,
				DefaultVal:   defaultVal,
				BoundPointer: valueField.Addr().Interface(),
				BoundType:    valueField.Type(),
			}
			c.boundParametersMapping[valueField.Addr()] = name
		}

		if tagNoFlag, exists := typeField.Tag.Lookup("noflag"); exists && tagNoFlag == "true" {
			if err := c.SetDefault(name, valueField.Interface()); err != nil {
				panic(fmt.Sprintf("could not set default value of %s, error: %s", name, err))
			}
			addParameter(name, shortHand, usage, valueField.Interface(), valueField)

			continue
		}

		// only use the default value from the tags if the value is the zero value
		isZeroValue := valueField.IsZero()

		//nolint:forcetypeassert // false positive
		switch defaultValue := valueField.Interface().(type) {
		case bool:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseBool(tagDefaultValue)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = value
			}
			flagset.BoolVarP(valueField.Addr().Interface().(*bool), name, shortHand, defaultValue, usage)

		case *bool:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseBool(tagDefaultValue)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = &value
			}
			flagset.BoolVarP(valueField.Interface().(*bool), name, shortHand, *defaultValue, usage)

		case time.Duration:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				parsedDuration, err := time.ParseDuration(tagDefaultValue)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = parsedDuration
			}
			flagset.DurationVarP(valueField.Addr().Interface().(*time.Duration), name, shortHand, defaultValue, usage)

		case float32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseFloat(tagDefaultValue, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = float32(value)
			}
			flagset.Float32VarP(valueField.Addr().Interface().(*float32), name, shortHand, defaultValue, usage)

		case *float32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseFloat(tagDefaultValue, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueFloat32 := float32(value)
				defaultValue = &valueFloat32
			}
			flagset.Float32VarP(valueField.Interface().(*float32), name, shortHand, *defaultValue, usage)

		case float64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseFloat(tagDefaultValue, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = value
			}
			flagset.Float64VarP(valueField.Addr().Interface().(*float64), name, shortHand, defaultValue, usage)

		case *float64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseFloat(tagDefaultValue, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = &value
			}
			flagset.Float64VarP(valueField.Interface().(*float64), name, shortHand, *defaultValue, usage)

		case int:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = int(value)
			}
			flagset.IntVarP(valueField.Addr().Interface().(*int), name, shortHand, defaultValue, usage)

		case *int:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueInt := int(value)
				defaultValue = &valueInt
			}
			flagset.IntVarP(valueField.Interface().(*int), name, shortHand, *defaultValue, usage)

		case int8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 8)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = int8(value)
			}
			flagset.Int8VarP(valueField.Addr().Interface().(*int8), name, shortHand, defaultValue, usage)

		case *int8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 8)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueInt8 := int8(value)
				defaultValue = &valueInt8
			}
			flagset.Int8VarP(valueField.Interface().(*int8), name, shortHand, *defaultValue, usage)

		case int16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 16)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = int16(value)
			}
			flagset.Int16VarP(valueField.Addr().Interface().(*int16), name, shortHand, defaultValue, usage)

		case *int16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 16)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueInt16 := int16(value)
				defaultValue = &valueInt16
			}
			flagset.Int16VarP(valueField.Interface().(*int16), name, shortHand, *defaultValue, usage)

		case int32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = int32(value)
			}
			flagset.Int32VarP(valueField.Addr().Interface().(*int32), name, shortHand, defaultValue, usage)

		case *int32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueInt32 := int32(value)
				defaultValue = &valueInt32
			}
			flagset.Int32VarP(valueField.Interface().(*int32), name, shortHand, *defaultValue, usage)

		case int64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = value
			}
			flagset.Int64VarP(valueField.Addr().Interface().(*int64), name, shortHand, defaultValue, usage)

		case *int64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = &value
			}
			flagset.Int64VarP(valueField.Interface().(*int64), name, shortHand, *defaultValue, usage)

		case string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				defaultValue = tagDefaultValue
			}
			flagset.StringVarP(valueField.Addr().Interface().(*string), name, shortHand, defaultValue, usage)

		case *string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				defaultValue = &tagDefaultValue
			}
			flagset.StringVarP(valueField.Interface().(*string), name, shortHand, *defaultValue, usage)

		case uint:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = uint(value)
			}
			flagset.UintVarP(valueField.Addr().Interface().(*uint), name, shortHand, defaultValue, usage)

		case *uint:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueUint := uint(value)
				defaultValue = &valueUint
			}
			flagset.UintVarP(valueField.Interface().(*uint), name, shortHand, *defaultValue, usage)

		case uint8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 8)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = uint8(value)
			}
			flagset.Uint8VarP(valueField.Addr().Interface().(*uint8), name, shortHand, defaultValue, usage)

		case *uint8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 8)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueUint8 := uint8(value)
				defaultValue = &valueUint8
			}
			flagset.Uint8VarP(valueField.Interface().(*uint8), name, shortHand, *defaultValue, usage)

		case uint16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 16)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = uint16(value)
			}
			flagset.Uint16VarP(valueField.Addr().Interface().(*uint16), name, shortHand, defaultValue, usage)

		case *uint16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 16)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueUint16 := uint16(value)
				defaultValue = &valueUint16
			}
			flagset.Uint16VarP(valueField.Interface().(*uint16), name, shortHand, *defaultValue, usage)

		case uint32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = uint32(value)
			}
			flagset.Uint32VarP(valueField.Addr().Interface().(*uint32), name, shortHand, defaultValue, usage)

		case *uint32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 32)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				valueUint32 := uint32(value)
				defaultValue = &valueUint32
			}
			flagset.Uint32VarP(valueField.Interface().(*uint32), name, shortHand, *defaultValue, usage)

		case uint64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = value
			}
			flagset.Uint64VarP(valueField.Addr().Interface().(*uint64), name, shortHand, defaultValue, usage)

		case *uint64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
				}

				defaultValue = &value
			}
			flagset.Uint64VarP(valueField.Interface().(*uint64), name, shortHand, *defaultValue, usage)

		case []string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				if len(tagDefaultValue) == 0 {
					defaultValue = []string{}
				} else {
					defaultValue = strings.Split(tagDefaultValue, ",")
				}
			}
			flagset.StringSliceVarP(valueField.Addr().Interface().(*[]string), name, shortHand, defaultValue, usage)

		case map[string]string:
			if _, exists := typeField.Tag.Lookup("default"); exists && isZeroValue {
				panic(fmt.Sprintf("passing default value of '%s' via tag not allowed", name))
			}
			flagset.StringToStringVarP(valueField.Addr().Interface().(*map[string]string), name, shortHand, defaultValue, usage)

		default:
			if valueField.Kind() == reflect.Slice {
				panic(fmt.Sprintf("could not bind '%s' because it is a slice value. did you forget the 'noflag:\"true\"' tag?", name))
			}

			// recursively walk the value, but do no add it as a parameter
			c.BindParameters(flagset, name, valueField.Addr().Interface())

			continue
		}
		addParameter(name, shortHand, usage, valueField.Interface(), valueField)
	}
}

// UpdateBoundParameters updates parameters that were bound using the BindParameters method with the current values in
// the configuration.
func (c *Configuration) UpdateBoundParameters() {
	for _, boundParameter := range c.boundParameters {
		parameterName := boundParameter.Name

		//nolint:forcetypeassert // type switch with reflect.Type
		switch boundParameter.BoundType {
		case reflectutils.BoolType:
			*(boundParameter.BoundPointer.(*bool)) = c.Bool(parameterName)
		case reflectutils.TimeDurationType:
			*(boundParameter.BoundPointer.(*time.Duration)) = c.Duration(parameterName)
		case reflectutils.Float32Type:
			*(boundParameter.BoundPointer.(*float32)) = float32(c.Float64(parameterName))
		case reflectutils.Float64Type:
			*(boundParameter.BoundPointer.(*float64)) = c.Float64(parameterName)
		case reflectutils.IntType:
			*(boundParameter.BoundPointer.(*int)) = c.Int(parameterName)
		case reflectutils.Int8Type:
			*(boundParameter.BoundPointer.(*int8)) = int8(c.Int(parameterName))
		case reflectutils.Int16Type:
			*(boundParameter.BoundPointer.(*int16)) = int16(c.Int(parameterName))
		case reflectutils.Int32Type:
			*(boundParameter.BoundPointer.(*int32)) = int32(c.Int(parameterName))
		case reflectutils.Int64Type:
			*(boundParameter.BoundPointer.(*int64)) = c.Int64(parameterName)
		case reflectutils.StringType:
			*(boundParameter.BoundPointer.(*string)) = c.String(parameterName)
		case reflectutils.UintType:
			*(boundParameter.BoundPointer.(*uint)) = uint(c.Int(parameterName))
		case reflectutils.Uint8Type:
			*(boundParameter.BoundPointer.(*uint8)) = uint8(c.Int(parameterName))
		case reflectutils.Uint16Type:
			*(boundParameter.BoundPointer.(*uint16)) = uint16(c.Int(parameterName))
		case reflectutils.Uint32Type:
			*(boundParameter.BoundPointer.(*uint32)) = uint32(c.Int(parameterName))
		case reflectutils.Uint64Type:
			*(boundParameter.BoundPointer.(*uint64)) = uint64(c.Int64(parameterName))
		case reflectutils.StringSliceType:
			*(boundParameter.BoundPointer.(*[]string)) = c.Strings(parameterName)
		case reflectutils.StringMapType:
			*(boundParameter.BoundPointer.(*map[string]string)) = c.StringMap(parameterName)

		default:
			// we need to create a new empty object of the same type,
			// otherwise we may only overwrite first values of a slice of bigger length
			newBoundParameterPointer := reflect.New(boundParameter.BoundType)
			newBoundParameter := newBoundParameterPointer.Interface()

			// if we don't know the type, we try to unmarshal it
			if err := c.Unmarshal(parameterName, newBoundParameter); err != nil {
				panic(fmt.Sprintf("could not unmarshal value of '%s', error: %s", parameterName, err))
			}

			// Overwrite the original value with the new value
			reflect.ValueOf(boundParameter.BoundPointer).Elem().Set(newBoundParameterPointer.Elem())
		}
	}
}

// GetParameterPath returns the path to the parameter with the given name.
// It needs to be called with a pointer to the struct field that was bound to retrieve the path.
func (c *Configuration) GetParameterPath(parameter any) string {
	return c.boundParametersMapping[reflect.ValueOf(parameter)]
}
