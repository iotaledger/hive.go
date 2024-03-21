package parameter

import (
	"fmt"
	"reflect"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	ErrUnknownEntryType = ierrors.New("unknown entry type")
)

//nolint:revive // better be explicit here
type ParameterGroup struct {
	Parameters []*Parameter
	SubGroups  []*ParameterGroup
	Entries    []any
	Name       string
	BaseName   string
	Level      int
	Default    any
}

func createGroup(name string, baseName string, level int, defaultVal any) *ParameterGroup {
	return &ParameterGroup{
		Parameters: make([]*Parameter, 0),
		SubGroups:  make([]*ParameterGroup, 0),
		Name:       name,
		BaseName:   baseName,
		Level:      level,
		Default:    defaultVal,
	}
}

func addGroup(groupsMap map[string]*ParameterGroup, groups []*ParameterGroup, baseName string, groupName string, level int, defaultVal any) ([]*ParameterGroup, string) {
	groupBaseName := groupName
	if baseName != "" {
		groupBaseName = fmt.Sprintf("%s.%s", baseName, groupName)
	}

	// check if the group already exists
	if _, exists := groupsMap[groupBaseName]; !exists {
		newGroup := createGroup(groupName, groupBaseName, level, defaultVal)
		groupsMap[groupBaseName] = newGroup
		groups = append(groups, newGroup)

		// check if this is a subgroup
		if baseName != "" {
			if parent, exists := groupsMap[baseName]; exists {
				parent.SubGroups = append(parent.SubGroups, newGroup)
				parent.Entries = append(parent.Entries, newGroup)
			}
		}
	}

	return groups, groupBaseName
}

func analyzeBoundParameter(groupsMap map[string]*ParameterGroup, groups []*ParameterGroup, boundParam *configuration.BoundParameter, baseName string, name string, level int) []*ParameterGroup {
	if strings.Contains(name, ".") {
		// name still contains a separator => create a group and walk deeper
		groupName, keyName, _ := strings.Cut(name, ".")
		var groupBaseName string
		groups, groupBaseName = addGroup(groupsMap, groups, baseName, groupName, level, nil)
		analyzeBoundParameter(groupsMap, groups, boundParam, groupBaseName, keyName, level+1)
	} else {
		// no separator found, this must be a parameter, or a slice of a struct
		if reflect.TypeOf(boundParam.DefaultVal).Kind() == reflect.Slice {
			// slice found, check if contains a ptr to a struct, or a struct
			typ := reflect.TypeOf(boundParam.DefaultVal).Elem()

			var elem reflect.Value
			if typ.Kind() == reflect.Ptr {
				elem = reflect.New(typ.Elem()).Elem()
			}
			if typ.Kind() == reflect.Struct {
				elem = reflect.New(typ).Elem()
			}

			if elem.IsValid() {
				// valid struct or pointer, add this as a group and analyze all fields to create parameters
				var groupBaseName string
				groups, groupBaseName = addGroup(groupsMap, groups, baseName, name, level, boundParam.DefaultVal)

				for i := range elem.NumField() {
					valueField := elem.Field(i)
					typeField := elem.Type().Field(i)

					name, usage, defaultVal := getParameterValues(valueField, typeField)
					addParameter(groupsMap, defaultVal, usage, groupBaseName, name)
				}

				return groups
			}
		}

		// parameter is either no slice, or a slice of basic types
		addParameter(groupsMap, boundParam.DefaultVal, boundParam.Usage, baseName, name)
	}

	return groups
}

func ParseConfigParameterGroups(config *configuration.Configuration, flagset *flag.FlagSet, ignoreFlags map[string]struct{}) []*ParameterGroup {
	configKeys := config.Koanf().All()

	flagset.SortFlags = false

	groupsMap := make(map[string]*ParameterGroup)
	groups := make([]*ParameterGroup, 0)

	// collect all keys
	flagset.VisitAll(func(f *flag.Flag) {
		delete(configKeys, strings.ToLower(f.Name))

		if _, ignore := ignoreFlags[strings.ToLower(f.Name)]; ignore {
			return
		}

		boundParam := config.BoundParameter(f.Name)
		if boundParam == nil {
			panic(f.Name)
		}

		groups = analyzeBoundParameter(groupsMap, groups, boundParam, "", f.Name, 0)
	})

	// analyze all missing parameters
	for key := range configKeys {
		if _, ignore := ignoreFlags[strings.ToLower(key)]; ignore {
			continue
		}

		boundParam := config.BoundParameter(key)
		if boundParam == nil {
			panic(key)
		}

		groups = analyzeBoundParameter(groupsMap, groups, boundParam, "", boundParam.Name, 0)
	}

	return groups
}
