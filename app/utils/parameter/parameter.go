package parameter

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/configuration"
)

type Parameter struct {
	Name        string
	Description string
	Type        string
	Default     any
	DefaultStr  string
}

func createParameter(defaultVal any, usage string, name string) *Parameter {

	typeName := ""
	defaultValue := defaultVal
	defaultValueStr := ""
	switch v := defaultVal.(type) {
	case int, int8, int16, int32, int64:
		defaultValueStr = fmt.Sprintf("%d", v)
		typeName = "int"
	case uint, uint8, uint16, uint32, uint64:
		defaultValueStr = fmt.Sprintf("%d", v)
		typeName = "uint"
	case float32, float64:
		defaultValueStr = fmt.Sprintf("%0.1f", v)
		typeName = "float"
	case bool:
		defaultValueStr = fmt.Sprintf("%t", v)
		typeName = "boolean"
	case string:
		defaultValueStr = fmt.Sprintf("\"%s\"", v)
		typeName = "string"
	case time.Duration:
		defaultValue = fmt.Sprintf("%s", durationShortened(v))
		defaultValueStr = fmt.Sprintf("\"%s\"", durationShortened(v))
		typeName = "string"
	case []string:
		defaultValueStr = ""
		if len(v) > 0 {
			defaultValueStr = strings.Join(v, "<br/>")
		}
		typeName = "array"
	case map[string]string:
		defaultValueStr = "["
		if len(v) > 0 {
			i := 0
			for key, value := range v {
				if i < len(v)-1 {
					defaultValueStr += fmt.Sprintf("%s=%s,", key, value)
				} else {
					defaultValueStr += fmt.Sprintf("%s=%s", key, value)
				}
				i++
			}
		}
		defaultValueStr += "]"
		typeName = "object"
	default:
		panic(fmt.Sprintf("unknown type (%s), name: %s", reflect.TypeOf(defaultVal).Name(), name))
	}

	var description string
	if len(usage) > 1 {
		description = strings.ToUpper(usage[:1]) + usage[1:]
	}

	return &Parameter{
		Name:        name,
		Description: description,
		Type:        typeName,
		Default:     defaultValue,
		DefaultStr:  defaultValueStr,
	}
}

func addParameter(groupsMap map[string]*ParameterGroup, defaultVal any, usage string, baseName string, name string) {

	group, exists := groupsMap[baseName]
	if !exists {
		panic(fmt.Sprintf("group not found: %s", baseName))
	}

	newParameter := createParameter(defaultVal, usage, name)
	group.Parameters = append(group.Parameters, newParameter)
	group.Entries = append(group.Entries, newParameter)
}

func getParameterValues(namespace string, valueField reflect.Value, typeField reflect.StructField) (string, string, any) {

	var name string
	if tagName, exists := typeField.Tag.Lookup("name"); exists {
		name += tagName
	} else {
		name += configuration.LowerCamelCase(typeField.Name)
	}
	usage, _ := typeField.Tag.Lookup("usage")

	if tagNoFlag, exists := typeField.Tag.Lookup("noflag"); exists && tagNoFlag == "true" {
		return name, usage, valueField.Interface()
	}

	defaultValue := valueField.Interface()
	switch valueField.Interface().(type) {
	case bool:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseBool(tagDefaultValue); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = value
			}
		}

	case time.Duration:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if parsedDuration, err := time.ParseDuration(tagDefaultValue); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = parsedDuration
			}
		}

	case float32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseFloat(tagDefaultValue, 32); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = float32(value)
			}
		}

	case float64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseFloat(tagDefaultValue, 64); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = value
			}
		}

	case int:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseInt(tagDefaultValue, 10, 64); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = int(value)
			}
		}

	case int8:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseInt(tagDefaultValue, 10, 8); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = int8(value)
			}
		}

	case int16:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseInt(tagDefaultValue, 10, 16); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = int16(value)
			}
		}

	case int32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseInt(tagDefaultValue, 10, 32); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = int32(value)
			}
		}

	case int64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseInt(tagDefaultValue, 10, 64); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = value
			}
		}

	case string:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			defaultValue = tagDefaultValue
		}

	case uint:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseUint(tagDefaultValue, 10, 64); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = uint(value)
			}
		}

	case uint8:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseUint(tagDefaultValue, 10, 8); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = uint8(value)
			}
		}

	case uint16:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseUint(tagDefaultValue, 10, 16); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = uint16(value)
			}
		}

	case uint32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseUint(tagDefaultValue, 10, 32); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = uint32(value)
			}
		}

	case uint64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if value, err := strconv.ParseUint(tagDefaultValue, 10, 64); err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			} else {
				defaultValue = value
			}
		}

	case []string:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			if tagDefaultValue == "" {
				defaultValue = []string{}
			} else {
				defaultValue = strings.Split(tagDefaultValue, ",")
			}
		}

	case map[string]string:
		if _, exists := typeField.Tag.Lookup("default"); exists {
			panic(fmt.Sprintf("passing default value of '%s' via tag not allowed", name))
		}

	default:
		if valueField.Kind() == reflect.Slice {
			panic(fmt.Sprintf("could not bind '%s' because it is a slice value. did you forget the 'noflag:\"true\"' tag?", name))
		}
	}

	return name, usage, defaultValue
}
