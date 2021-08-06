package configuration

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/pflag"
)

// boundParameters keeps track of all parameters that were bound using the BindParameters function.
var boundParameters = make(map[string]*BoundParameter)

// BoundParameter stores the pointer and the type of values that were bound using the BindParameters function.
type BoundParameter struct {
	boundPointer interface{}
	boundType    string
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
func BindParameters(pointerToStruct interface{}, optionalNamespace ...string) {
	var prefix string
	if len(optionalNamespace) == 0 {
		prefix = lowerCamelCase(callerShortPackageName())
	} else {
		prefix = optionalNamespace[0]
	}

	val := reflect.ValueOf(pointerToStruct).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		name := prefix + "."
		if tagName, exists := typeField.Tag.Lookup("name"); exists {
			name += tagName
		} else {
			name += lowerCamelCase(typeField.Name)
		}

		switch defaultValue := valueField.Interface().(type) {
		case bool:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.BoolVarP(valueField.Addr().Interface().(*bool), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.BoolVar(valueField.Addr().Interface().(*bool), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "bool",
			}
		case time.Duration:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if parsedDuration, err := time.ParseDuration(tagDefaultValue); err != nil {
					panic(err)
				} else {
					defaultValue = parsedDuration
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.DurationVarP(valueField.Addr().Interface().(*time.Duration), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.DurationVar(valueField.Addr().Interface().(*time.Duration), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "time.Duration",
			}
		case float32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Float32VarP(valueField.Addr().Interface().(*float32), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Float32Var(valueField.Addr().Interface().(*float32), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "float32",
			}
		case float64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Float64VarP(valueField.Addr().Interface().(*float64), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Float64Var(valueField.Addr().Interface().(*float64), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "float64",
			}
		case int:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.IntVarP(valueField.Addr().Interface().(*int), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.IntVar(valueField.Addr().Interface().(*int), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "int",
			}
		case int8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Int8VarP(valueField.Addr().Interface().(*int8), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Int8Var(valueField.Addr().Interface().(*int8), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "int8",
			}
		case int16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Int16VarP(valueField.Addr().Interface().(*int16), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Int16Var(valueField.Addr().Interface().(*int16), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "int16",
			}
		case int32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Int32VarP(valueField.Addr().Interface().(*int32), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Int32Var(valueField.Addr().Interface().(*int32), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "int32",
			}
		case int64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Int64VarP(valueField.Addr().Interface().(*int64), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Int64Var(valueField.Addr().Interface().(*int64), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "int64",
			}
		case string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.StringVarP(valueField.Addr().Interface().(*string), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.StringVar(valueField.Addr().Interface().(*string), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "string",
			}
		case uint:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.UintVarP(valueField.Addr().Interface().(*uint), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.UintVar(valueField.Addr().Interface().(*uint), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "uint",
			}
		case uint8:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Uint8VarP(valueField.Addr().Interface().(*uint8), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Uint8Var(valueField.Addr().Interface().(*uint8), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "uint8",
			}
		case uint16:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Uint16VarP(valueField.Addr().Interface().(*uint16), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Uint16Var(valueField.Addr().Interface().(*uint16), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "uint16",
			}
		case uint32:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Uint32VarP(valueField.Addr().Interface().(*uint32), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Uint32Var(valueField.Addr().Interface().(*uint32), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "uint32",
			}
		case uint64:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				if _, err := fmt.Sscan(tagDefaultValue, &defaultValue); err != nil {
					panic(err)
				}
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.Uint64VarP(valueField.Addr().Interface().(*uint64), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.Uint64Var(valueField.Addr().Interface().(*uint64), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "uint64",
			}
		case []string:
			if tagDefaultValue, exists := typeField.Tag.Lookup("default"); exists {
				parsedValue, err := csv.NewReader(strings.NewReader(tagDefaultValue)).Read()
				if err != nil {
					panic(err)
				}
				defaultValue = parsedValue
			}

			if shortHand, exists := typeField.Tag.Lookup("shorthand"); exists {
				pflag.StringSliceVarP(valueField.Addr().Interface().(*[]string), name, shortHand, defaultValue, typeField.Tag.Get("usage"))
			} else {
				pflag.StringSliceVar(valueField.Addr().Interface().(*[]string), name, defaultValue, typeField.Tag.Get("usage"))
			}

			boundParameters[name] = &BoundParameter{
				boundPointer: valueField.Addr().Interface(),
				boundType:    "[]string",
			}
		default:
			BindParameters(valueField.Addr().Interface(), name)
		}
	}
}

// UpdateBoundParameters updates parameters that were bound using the BindParameters method with the current values in
// the configuration.
func UpdateBoundParameters(configuration *Configuration) {
	for parameterName, boundParameter := range boundParameters {
		switch boundParameter.boundType {
		case "bool":
			*(boundParameter.boundPointer.(*bool)) = configuration.Bool(parameterName)
		case "time.Duration":
			*(boundParameter.boundPointer.(*time.Duration)) = configuration.Duration(parameterName)
		case "float32":
			*(boundParameter.boundPointer.(*float32)) = float32(configuration.Float64(parameterName))
		case "float64":
			*(boundParameter.boundPointer.(*float64)) = configuration.Float64(parameterName)
		case "int":
			*(boundParameter.boundPointer.(*int)) = configuration.Int(parameterName)
		case "int8":
			*(boundParameter.boundPointer.(*int8)) = int8(configuration.Int(parameterName))
		case "int16":
			*(boundParameter.boundPointer.(*int16)) = int16(configuration.Int(parameterName))
		case "int32":
			*(boundParameter.boundPointer.(*int32)) = int32(configuration.Int(parameterName))
		case "int64":
			*(boundParameter.boundPointer.(*int64)) = configuration.Int64(parameterName)
		case "string":
			*(boundParameter.boundPointer.(*string)) = configuration.String(parameterName)
		case "uint":
			*(boundParameter.boundPointer.(*uint)) = uint(configuration.Int(parameterName))
		case "uint8":
			*(boundParameter.boundPointer.(*uint8)) = uint8(configuration.Int(parameterName))
		case "uint16":
			*(boundParameter.boundPointer.(*uint16)) = uint16(configuration.Int(parameterName))
		case "uint32":
			*(boundParameter.boundPointer.(*uint32)) = uint32(configuration.Int(parameterName))
		case "uint64":
			*(boundParameter.boundPointer.(*uint64)) = uint64(configuration.Int64(parameterName))
		case "[]string":
			*(boundParameter.boundPointer.(*[]string)) = configuration.Strings(parameterName)
		}
	}
}

func lowerCamelCase(str string) string {
	runes := []rune(str)
	runeCount := len(runes)

	if runeCount == 0 || unicode.IsLower(runes[0]) {
		return str
	}

	runes[0] = unicode.ToLower(runes[0])
	if runeCount == 1 || unicode.IsLower(runes[1]) {
		return string(runes)
	}

	for i := 1; i < runeCount; i++ {
		if i+1 < runeCount && unicode.IsLower(runes[i+1]) {
			break
		}

		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

func callerShortPackageName() string {
	pc, _, _, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.IndexByte(funcName[lastSlash:], '.') + lastSlash

	return funcName[lastSlash+1 : firstDot]
}
