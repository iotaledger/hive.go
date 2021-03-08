package configuration

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/pflag"
)

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
		default:
			BindParameters(valueField.Addr().Interface(), name)
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
