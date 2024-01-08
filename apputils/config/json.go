package config

import (
	"encoding/json"

	// we need to use this orderedmap implementation for serialization instead of our own,
	// because the generic orderedmap in hive.go doesn't support marshaling to json.
	// this orderedmap implementation uses map[string]any as underlying datastructure,
	// which is a must instead of map[K]V, otherwise we can't correctly sort nested maps during unmarshaling.
	"github.com/iancoleman/orderedmap"
	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/apputils/parameter"
)

type parameterMapJSON struct {
	*orderedmap.OrderedMap
}

func newParameterMapJSON() *parameterMapJSON {
	return &parameterMapJSON{
		OrderedMap: orderedmap.New(),
	}
}

func (p *parameterMapJSON) AddEntry(entry any) {
	switch v := entry.(type) {
	case *parameter.ParameterGroup:
		newParamMapJSONGroup := newParameterMapJSON()

		p.Set(v.Name, newParamMapJSONGroup)

		if v.Default == nil {
			for _, entry := range v.Entries {
				newParamMapJSONGroup.AddEntry(entry)
			}
		} else {
			p.Set(v.Name, v.Default)
		}

	case *parameter.Parameter:
		p.Set(v.Name, v.Default)

	default:
		panic(parameter.ErrUnknownEntryType)
	}
}

func (p *parameterMapJSON) PrettyPrint(prefix string, ident string) string {
	data, err := json.MarshalIndent(p, prefix, ident)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func prettyPrintParameterGroup(g *parameter.ParameterGroup, prefix string, indent string) string {
	paramMapJSON := newParameterMapJSON()
	paramMapJSON.AddEntry(g)

	return paramMapJSON.PrettyPrint(prefix, indent)
}

func GetDefaultAppConfigJSON(config *configuration.Configuration, flagset *flag.FlagSet, ignoreFlags map[string]struct{}) string {
	paramMapJSON := newParameterMapJSON()

	for _, group := range parameter.ParseConfigParameterGroups(config, flagset, ignoreFlags) {
		paramMapJSON.AddEntry(group)
	}

	return paramMapJSON.PrettyPrint("", "  ") + "\n"
}
