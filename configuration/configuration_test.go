package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func tempFile(t *testing.T, pattern string) (string, *os.File) {
	tmpfile, err := ioutil.TempFile("", pattern)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Remove(tmpfile.Name())
		require.NoError(t, err)
	})

	return tmpfile.Name(), tmpfile
}

func TestFetchGlobalFlags(t *testing.T) {
	flag.String("A", "321", "test")
	flag.Set("A", "321")

	config := New()

	err := config.LoadFlagSet(flag.CommandLine)
	require.NoError(t, err)

	val := config.String("A")
	require.EqualValues(t, "321", val)
}

func TestFetchFlagset(t *testing.T) {
	testFlagSet := flag.NewFlagSet("", flag.ContinueOnError)
	testFlagSet.String("A", "321", "test")
	testFlagSet.Set("A", "321")

	flag.Parse()
	config := New()

	err := config.LoadFlagSet(testFlagSet)
	require.NoError(t, err)

	val := config.String("A")
	require.EqualValues(t, "321", val)
}

func TestFetchEnvVars(t *testing.T) {
	testFlagSet := flag.NewFlagSet("", flag.ContinueOnError)
	testFlagSet.String("B", "322", "test")
	testFlagSet.Set("B", "322")

	os.Setenv("TEST_B", "321")

	os.Setenv("TEST_C", "321")

	config := New()

	err := config.LoadFlagSet(testFlagSet)
	require.NoError(t, err)

	err = config.LoadEnvironmentVars("TEST")
	require.NoError(t, err)

	val := config.String("B")
	require.EqualValues(t, "321", val)

	_, exists := config.All()["c"]
	require.False(t, exists, "expected read config value to not exist")
}

func TestFetchJSONFile(t *testing.T) {
	conf := make(map[string]int)
	conf["C"] = 321

	jsonConfFileName, jsonConfFile := tempFile(t, "config*.json")

	content, err := json.MarshalIndent(conf, "", "    ")
	require.NoError(t, err)

	_, err = jsonConfFile.Write(content)
	require.NoError(t, err)

	err = jsonConfFile.Close()
	require.NoError(t, err)

	config := New()

	err = config.LoadFile(jsonConfFileName)
	require.NoError(t, err)

	val := config.Int("C")
	require.EqualValues(t, 321, val)
}

func TestFetchYAMLFile(t *testing.T) {
	conf := make(map[string]int)
	conf["D"] = 321

	yamlConfFileName, yamlConfFile := tempFile(t, "config*.yaml")

	content, err := yaml.Marshal(conf)
	require.NoError(t, err)

	_, err = yamlConfFile.Write(content)
	require.NoError(t, err)

	err = yamlConfFile.Close()
	require.NoError(t, err)

	config := New()

	err = config.LoadFile(yamlConfFileName)
	require.NoError(t, err)

	val := config.Int("D")
	require.EqualValues(t, 321, val)
}

func TestMergeParameters(t *testing.T) {
	conf := make(map[string]int)
	conf["E"] = 321

	testFlagSet := flag.NewFlagSet("", flag.ContinueOnError)
	testFlagSet.Int("F", 321, "test")

	os.Setenv("TEST_F", "322")

	jsonConfFileName, jsonConfFile := tempFile(t, "config*.json")

	content, err := json.MarshalIndent(conf, "", "    ")
	require.NoError(t, err)

	_, err = jsonConfFile.Write(content)
	require.NoError(t, err)

	err = jsonConfFile.Close()
	require.NoError(t, err)

	config := New()

	err = config.LoadFile(jsonConfFileName)
	require.NoError(t, err)

	err = config.LoadFlagSet(testFlagSet)
	require.NoError(t, err)

	err = config.LoadEnvironmentVars("TEST")
	require.NoError(t, err)

	var exists bool

	_, exists = config.All()["e"]
	require.True(t, exists, "expected read config value to exist")

	// all keys should be lower cased
	_, exists = config.All()["E"]
	require.False(t, exists, "expected read config value to not exist")

	_, exists = config.All()["f"]
	require.True(t, exists, "expected read config value to exist")

	// all keys should be lower cased
	_, exists = config.All()["F"]
	require.False(t, exists, "expected read config value to not exist")

	_, exists = config.All()["g"]
	require.False(t, exists, "expected read config value to not exist")

	val := config.Int("E")
	require.EqualValues(t, 321, val)

	valStr := config.String("F")
	require.EqualValues(t, "322", valStr)

	val = config.Int("F")
	require.EqualValues(t, 322, val)
}

func TestSaveConfigFile(t *testing.T) {
	config1 := New()
	config1.Set("test.integer", 321)
	config1.Set("test.slice", []string{"string1", "string2", "string3"})
	config1.Set("test.bool.ignore", true)

	jsonConfFileName, _ := tempFile(t, "config*.json")
	err := config1.StoreFile(jsonConfFileName, []string{"test.bool.ignore"})
	require.NoError(t, err)

	config2 := New()

	err = config2.LoadFile(jsonConfFileName)
	require.NoError(t, err)

	valueInteger := config2.Int("test.integer")
	require.EqualValues(t, 321, valueInteger)

	valueSlice := config2.Strings("test.slice")
	require.EqualValues(t, []string{"string1", "string2", "string3"}, valueSlice)

	valueIgnoredBool := config2.Bool("test.bool.ignore")
	require.EqualValues(t, false, valueIgnoredBool)
}

func TestBindParameters(t *testing.T) {
	var parameters = struct {
		TestField  int64 `shorthand:"t" usage:"you can do stuff with this parameter"`
		TestField1 bool  `name:"bernd" default:"true" usage:"batman was here"`
		Nested     struct {
			Key       string `default:"elephant" usage:"nestedKey elephant"`
			SubNested struct {
				Key string `default:"duck" usage:"nestedKey duck"`
			}
		}
		Nested1 struct {
			Key string `name:"bird" shorthand:"b" default:"bird" usage:"nestedKey bird"`
		} `name:"renamedNested"`
		Batman []string      `default:"a,b" usage:"robin"`
		ItsAMe time.Duration `default:"60s" usage:"mario"`
	}{
		// assign default value outside of tag
		TestField: 13,
		// assign default value inside of tag (is overriden by default value of tag)
		Batman: []string{"a", "b", "c"},
	}

	config := New()
	config.BindParameters("configuration", &parameters)

	err := config.LoadFlagSet(flag.CommandLine)
	assert.NoError(t, err)

	// read in ENV variables
	// load the env vars after default values from flags were set (otherwise the env vars are not added because the keys don't exist)
	err = config.LoadEnvironmentVars("test")
	assert.NoError(t, err)

	// load the flags again to overwrite env vars that were also set via command line
	err = config.LoadFlagSet(flag.CommandLine)
	assert.NoError(t, err)

	config.UpdateBoundParameters()

	testFieldFlag := flag.Lookup("configuration.testField")
	assert.Equal(t, "you can do stuff with this parameter", testFieldFlag.Usage)
	assert.Equal(t, "13", testFieldFlag.DefValue)
	assert.Equal(t, "t", testFieldFlag.Shorthand)
	assert.Equal(t, "configuration.testField", testFieldFlag.Name)
	assert.Equal(t, "configuration.testField", config.GetParameterPath(&parameters.TestField))

	testField1Flag := flag.Lookup("configuration.bernd")
	assert.Equal(t, "batman was here", testField1Flag.Usage)
	assert.Equal(t, "true", testField1Flag.DefValue)
	assert.Equal(t, "", testField1Flag.Shorthand)
	assert.Equal(t, "configuration.bernd", testField1Flag.Name)
	assert.Equal(t, "configuration.bernd", config.GetParameterPath(&parameters.TestField1))

	elephantFlag := flag.Lookup("configuration.nested.key")
	assert.Equal(t, "nestedKey elephant", elephantFlag.Usage)
	assert.Equal(t, "elephant", elephantFlag.DefValue)
	assert.Equal(t, "", elephantFlag.Shorthand)
	assert.Equal(t, "configuration.nested.key", elephantFlag.Name)
	assert.Equal(t, "configuration.nested.key", config.GetParameterPath(&parameters.Nested.Key))

	duckFlag := flag.Lookup("configuration.nested.subNested.key")
	assert.Equal(t, "nestedKey duck", duckFlag.Usage)
	assert.Equal(t, "duck", duckFlag.DefValue)
	assert.Equal(t, "", duckFlag.Shorthand)
	assert.Equal(t, "configuration.nested.subNested.key", duckFlag.Name)
	assert.Equal(t, "configuration.nested.subNested.key", config.GetParameterPath(&parameters.Nested.SubNested.Key))

	birdFlag := flag.Lookup("configuration.renamedNested.bird")
	assert.Equal(t, "nestedKey bird", birdFlag.Usage)
	assert.Equal(t, "bird", birdFlag.DefValue)
	assert.Equal(t, "b", birdFlag.Shorthand)
	assert.Equal(t, "configuration.renamedNested.bird", birdFlag.Name)
	assert.Equal(t, "configuration.renamedNested.bird", config.GetParameterPath(&parameters.Nested1.Key))

	batmanFlag := flag.Lookup("configuration.batman")
	assert.Equal(t, "robin", batmanFlag.Usage)
	assert.Equal(t, []string{"a", "b"}, parameters.Batman)
	assert.Equal(t, "", batmanFlag.Shorthand)
	assert.Equal(t, "configuration.batman", batmanFlag.Name)
	assert.Equal(t, "configuration.batman", config.GetParameterPath(&parameters.Batman))

	ItsAMeFlag := flag.Lookup("configuration.itsAMe")
	assert.Equal(t, "mario", ItsAMeFlag.Usage)
	assert.Equal(t, 60*time.Second, parameters.ItsAMe)
	assert.Equal(t, "", ItsAMeFlag.Shorthand)
	assert.Equal(t, "configuration.itsAMe", ItsAMeFlag.Name)
	assert.Equal(t, "configuration.itsAMe", config.GetParameterPath(&parameters.ItsAMe))
}
