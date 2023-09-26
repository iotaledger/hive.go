package configuration

import (
	"strings"
	"time"

	"github.com/knadh/koanf"
)

// All returns a map of all flattened key paths and their values.
// Note that it uses maps.Copy to create a copy that uses
// json.Marshal which changes the numeric types to float64.
func (c *Configuration) All() map[string]interface{} {
	return c.config.All()
}

// Get returns the raw, uncast interface{} value of a given key path
// in the config map. If the key path does not exist, nil is returned.
func (c *Configuration) Get(path string) interface{} {
	return c.config.Get(strings.ToLower(path))
}

// Exists returns true if the given key path exists in the conf map.
func (c *Configuration) Exists(path string) bool {
	return c.config.Exists(strings.ToLower(path))
}

// Int64 returns the int64 value of a given key path or 0 if the path
// does not exist or if the value is not a valid int64.
func (c *Configuration) Int64(path string) int64 {
	return c.config.Int64(strings.ToLower(path))
}

// Int64s returns the []int64 slice value of a given key path or an
// empty []int64 slice if the path does not exist or if the value
// is not a valid int slice.
func (c *Configuration) Int64s(path string) []int64 {
	return c.config.Int64s(strings.ToLower(path))
}

// Int64Map returns the map[string]int64 value of a given key path
// or an empty map[string]int64 if the path does not exist or if the
// value is not a valid int64 map.
func (c *Configuration) Int64Map(path string) map[string]int64 {
	return c.config.Int64Map(strings.ToLower(path))
}

// Int returns the int value of a given key path or 0 if the path
// does not exist or if the value is not a valid int.
func (c *Configuration) Int(path string) int {
	return c.config.Int(strings.ToLower(path))
}

// Ints returns the []int slice value of a given key path or an
// empty []int slice if the path does not exist or if the value
// is not a valid int slice.
func (c *Configuration) Ints(path string) []int {
	return c.config.Ints(strings.ToLower(path))
}

// IntMap returns the map[string]int value of a given key path
// or an empty map[string]int if the path does not exist or if the
// value is not a valid int map.
func (c *Configuration) IntMap(path string) map[string]int {
	return c.config.IntMap(strings.ToLower(path))
}

// Float64 returns the float64 value of a given key path or 0 if the path
// does not exist or if the value is not a valid float64.
func (c *Configuration) Float64(path string) float64 {
	return c.config.Float64(strings.ToLower(path))
}

// Float64s returns the []float64 slice value of a given key path or an
// empty []float64 slice if the path does not exist or if the value
// is not a valid float64 slice.
func (c *Configuration) Float64s(path string) []float64 {
	return c.config.Float64s(strings.ToLower(path))
}

// Float64Map returns the map[string]float64 value of a given key path
// or an empty map[string]float64 if the path does not exist or if the
// value is not a valid float64 map.
func (c *Configuration) Float64Map(path string) map[string]float64 {
	return c.config.Float64Map(strings.ToLower(path))
}

// Duration returns the time.Duration value of a given key path assuming
// that the key contains a valid numeric value.
func (c *Configuration) Duration(path string) time.Duration {
	return c.config.Duration(strings.ToLower(path))
}

// Time attempts to parse the value of a given key path and return time.Time
// representation. If the value is numeric, it is treated as a UNIX timestamp
// and if it's string, a parse is attempted with the given layout.
func (c *Configuration) Time(path, layout string) time.Time {
	return c.config.Time(strings.ToLower(path), layout)
}

// String returns the string value of a given key path or "" if the path
// does not exist or if the value is not a valid string.
func (c *Configuration) String(path string) string {
	return c.config.String(strings.ToLower(path))
}

// Strings returns the []string slice value of a given key path or an
// empty []string slice if the path does not exist or if the value
// is not a valid string slice.
func (c *Configuration) Strings(path string) []string {
	return c.config.Strings(strings.ToLower(path))
}

// StringMap returns the map[string]string value of a given key path
// or an empty map[string]string if the path does not exist or if the
// value is not a valid string map.
func (c *Configuration) StringMap(path string) map[string]string {
	return c.config.StringMap(strings.ToLower(path))
}

// StringsMap returns the map[string][]string value of a given key path
// or an empty map[string][]string if the path does not exist or if the
// value is not a valid strings map.
func (c *Configuration) StringsMap(path string) map[string][]string {
	return c.config.StringsMap(strings.ToLower(path))
}

// Bytes returns the []byte value of a given key path or an empty
// []byte slice if the path does not exist or if the value is not a valid string.
func (c *Configuration) Bytes(path string) []byte {
	return c.config.Bytes(strings.ToLower(path))
}

// Bool returns the bool value of a given key path or false if the path
// does not exist or if the value is not a valid bool representation.
// Accepted string representations of bool are the ones supported by strconv.ParseBool.
func (c *Configuration) Bool(path string) bool {
	return c.config.Bool(strings.ToLower(path))
}

// Bools returns the []bool slice value of a given key path or an
// empty []bool slice if the path does not exist or if the value
// is not a valid bool slice.
func (c *Configuration) Bools(path string) []bool {
	return c.config.Bools(strings.ToLower(path))
}

// BoolMap returns the map[string]bool value of a given key path
// or an empty map[string]bool if the path does not exist or if the
// value is not a valid bool map.
func (c *Configuration) BoolMap(path string) map[string]bool {
	return c.config.BoolMap(strings.ToLower(path))
}

// MapKeys returns a sorted string list of keys in a map addressed by the
// given path. If the path is not a map, an empty string slice is
// returned.
func (c *Configuration) MapKeys(path string) []string {
	return c.config.MapKeys(strings.ToLower(path))
}

// Unmarshal unmarshals a given key path into the given struct using
// the mapstructure lib. If no path is specified, the whole map is unmarshaled.
// `koanf` is the struct field tag used to match field names. To customize,
// use UnmarshalWithConf(). It uses the mitchellh/mapstructure package.
func (c *Configuration) Unmarshal(path string, o interface{}) error {
	return c.config.Unmarshal(strings.ToLower(path), o)
}

// UnmarshalWithConf is like Unmarshal but takes configuration params in UnmarshalConf.
// See mitchellh/mapstructure's DecoderConfig for advanced customization
// of the unmarshal behavior.
func (c *Configuration) UnmarshalWithConf(path string, o interface{}, uc koanf.UnmarshalConf) error {
	return c.config.UnmarshalWithConf(strings.ToLower(path), o, uc)
}

// BoundParameter returns the parameter that was bound to the configuration.
func (c *Configuration) BoundParameter(path string) *BoundParameter {
	return c.boundParameters[strings.ToLower(path)]
}
