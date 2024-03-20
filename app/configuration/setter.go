package configuration

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/structs"
)

// SetDefault sets the default value for the key (case-insensitive).
// Default is only applied if no value is provided via flag, file or env vars.
func (c *Configuration) SetDefault(path string, value interface{}, parser ...koanf.Parser) error {
	if c.config.Exists(strings.ToLower(path)) {
		// do not override values that already exist in the config
		return nil
	}

	return c.Set(path, value, parser...)
}

// SetDefaultStruct sets the default value for the key (case-insensitive).
// Default is only applied if no value is provided via flag, file or env vars.
func (c *Configuration) SetDefaultStruct(path string, value interface{}, fieldTag string, parser ...koanf.Parser) error {
	if c.config.Exists(strings.ToLower(path)) {
		// do not override values that already exist in the config
		return nil
	}

	return c.SetStruct(path, value, fieldTag, parser...)
}

// Set sets the value for the key (case-insensitive).
func (c *Configuration) Set(path string, value interface{}, parser ...koanf.Parser) error {
	var p koanf.Parser

	if len(parser) > 0 {
		// optional parser
		p = parser[0]
	}

	// koanf does not provide any special functions to set default values but uses the Provider interface to enable it.
	// Load default values using the confmap provider.
	// We provide a flat map with the "." delimiter.
	// A nested map can be loaded by setting the delimiter to an empty string "".
	return c.config.Load(confmap.Provider(map[string]interface{}{
		strings.ToLower(path): value,
	}, "."), p)
}

// SetStruct sets the value of the struct for the key (case-insensitive).
//
//nolint:revive
func (c *Configuration) SetStruct(path string, value interface{}, fieldTag string, parser ...koanf.Parser) error {
	var p koanf.Parser

	if len(parser) > 0 {
		// optional parser
		p = parser[0]
	}

	// Load default values using the structs provider.
	// We provide a struct along with the struct tag `fieldTag` to the
	// provider.
	return c.config.Load(structs.Provider(value, fieldTag), p)
}
