package config

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
	flag "github.com/spf13/pflag"

	"github.com/izuc/zipp.foundation/app/configuration"
	"github.com/izuc/zipp.foundation/apputils/parameter"
)

type parameterMapJSON struct {
	*orderedmap.OrderedMap
}

func newParameterMapJSON() *parameterMapJSON {
	return &parameterMapJSON{
		OrderedMap: orderedmap.New(),
	}
}

func (p *parameterMapJSON) AddEntry(entry interface{}) {
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
