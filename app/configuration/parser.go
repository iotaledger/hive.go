package configuration

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
)

func mapToLowerKeys(m map[string]interface{}) {
	for key, val := range m {
		switch v := val.(type) {
		case map[string]interface{}:
			// nested map: call recursively
			mapToLowerKeys(v)
		case map[interface{}]interface{}:
			// nested map: cast and call recursively
			stringMap := cast.ToStringMap(val)
			mapToLowerKeys(stringMap)
		}

		lower := strings.ToLower(key)
		if key != lower {
			// remove old key (not lower-cased)
			delete(m, key)
		}

		// update map
		m[lower] = val
	}
}

// JSONLowerParser implements a JSON parser.
// all config keys are lower cased.
type JSONLowerParser struct {
	prefix string
	indent string
}

// Unmarshal parses the given JSON bytes.
func (p *JSONLowerParser) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	mapToLowerKeys(out)

	return out, nil
}

// Marshal marshals the given config map to JSON bytes.
func (p *JSONLowerParser) Marshal(o map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(o, p.prefix, p.indent)
}

// YAMLLowerParser implements a YAML parser.
// all config keys are lower cased.
type YAMLLowerParser struct{}

// Unmarshal parses the given YAML bytes.
func (p *YAMLLowerParser) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	mapToLowerKeys(out)

	return out, nil
}

// Marshal marshals the given config map to YAML bytes.
func (p *YAMLLowerParser) Marshal(o map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(o)
}
