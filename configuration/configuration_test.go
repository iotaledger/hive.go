package configuration_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/iotaledger/hive.go/configuration"
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

	config := configuration.New()

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
	config := configuration.New()

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

	config := configuration.New()

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

	config := configuration.New()

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

	config := configuration.New()

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

	config := configuration.New()

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
