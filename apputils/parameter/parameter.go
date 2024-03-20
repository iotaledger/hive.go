package parameter

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/app/configuration"
)

const (
	TypeNameInt     = "int"
	TypeNameUint    = "uint"
	TypeNameBoolean = "boolean"
	TypeNameString  = "string"
	TypeNameFloat   = "float"
)

type Parameter struct {
	Name        string
	Description string
	Type        string
	Default     any
	DefaultStr  string
}

func createParameter(defaultVal any, usage string, name string) *Parameter {
	defaultValue := defaultVal

	var typeName string
	var defaultValueStr string
	switch v := defaultVal.(type) {
	case int, int8, int16, int32, int64:
		defaultValueStr = fmt.Sprintf("%d", v)
		typeName = TypeNameInt
	case *int:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameInt
	case *int8:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameInt
	case *int16:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameInt
	case *int32:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameInt
	case *int64:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameInt
	case uint, uint8, uint16, uint32, uint64:
		defaultValueStr = fmt.Sprintf("%d", v)
		typeName = TypeNameUint
	case *uint:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameUint
	case *uint8:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameUint
	case *uint16:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameUint
	case *uint32:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameUint
	case *uint64:
		defaultValueStr = fmt.Sprintf("%d", *v)
		typeName = TypeNameUint
	case float32, float64:
		defaultValueStr = fmt.Sprintf("%0.1f", v)
		typeName = TypeNameFloat
	case *float32:
		defaultValueStr = fmt.Sprintf("%0.1f", *v)
		typeName = TypeNameFloat
	case *float64:
		defaultValueStr = fmt.Sprintf("%0.1f", *v)
		typeName = TypeNameFloat
	case bool:
		defaultValueStr = fmt.Sprintf("%t", v)
		typeName = TypeNameBoolean
	case *bool:
		defaultValueStr = fmt.Sprintf("%t", *v)
		typeName = TypeNameBoolean
	case string:
		defaultValueStr = fmt.Sprintf("\"%s\"", v)
		typeName = TypeNameString
	case *string:
		defaultValueStr = fmt.Sprintf("\"%s\"", *v)
		typeName = TypeNameString
	case time.Duration:
		defaultValue = durationShortened(v)
		defaultValueStr = fmt.Sprintf("\"%s\"", durationShortened(v))
		typeName = TypeNameString
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
		switch reflect.ValueOf(v).Kind() {
		case reflect.Slice, reflect.Map:
			defaultValueStr = "see example below"
			typeName = "object"
		default:
			panic(fmt.Sprintf("unknown type (%s), name: %s", reflect.ValueOf(defaultVal).Type().String(), name))
		}
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

func getParameterValues(valueField reflect.Value, typeField reflect.StructField) (string, string, any) {
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
	case bool, *bool:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseBool(tagDefaultValue)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = value
		}

	case time.Duration:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			parsedDuration, err := time.ParseDuration(tagDefaultValue)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = parsedDuration
		}

	case float32, *float32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseFloat(tagDefaultValue, 32)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = float32(value)
		}

	case float64, *float64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseFloat(tagDefaultValue, 64)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = value
		}

	case int, *int:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = int(value)
		}

	case int8, *int8:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseInt(tagDefaultValue, 10, 8)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = int8(value)
		}

	case int16, *int16:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseInt(tagDefaultValue, 10, 16)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = int16(value)
		}

	case int32, *int32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseInt(tagDefaultValue, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = int32(value)
		}

	case int64, *int64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseInt(tagDefaultValue, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = value
		}

	case string, *string:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			defaultValue = tagDefaultValue
		}

	case uint, *uint:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = uint(value)
		}

	case uint8, *uint8:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseUint(tagDefaultValue, 10, 8)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = uint8(value)
		}

	case uint16, *uint16:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseUint(tagDefaultValue, 10, 16)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = uint16(value)
		}

	case uint32, *uint32:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseUint(tagDefaultValue, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = uint32(value)
		}

	case uint64, *uint64:
		if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
			value, err := strconv.ParseUint(tagDefaultValue, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("could not parse default value of '%s', error: %s", name, err))
			}

			defaultValue = value
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
