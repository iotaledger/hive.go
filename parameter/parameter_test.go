package parameter_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/parameter"
)

const (
	configName = "config"

	// we use a Windows path, just to please viper, as it otherwise
	// decides to append Windows drive letters to unix paths, when running
	// this test under Windows.
	//confDir = "C:/configDir"
	confDir = "/configDir"
)

var (
	memFS = afero.NewMemMapFs()
)

func TestMain(m *testing.M) {

	if err := memFS.MkdirAll(confDir, 0755); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestFetchJSONConfig(t *testing.T) {
	flag.String("a", "321", "test")
	flag.Set("a", "321")

	config := viper.New()
	config.SetFs(memFS)

	err := parameter.LoadConfigFile(config, confDir, configName, flag.CommandLine, true)
	if err != nil {
		t.Fatal(err)
	}

	val := config.GetString("a")
	if val != "321" {
		t.Fatalf("expected read config value to be %s, but was %s", "321", val)
	}
}

func TestFetchJSONConfigFlagConfigName(t *testing.T) {
	filename := fmt.Sprintf("%s/%s.json", confDir, configName)

	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err = jsonConfFile.WriteString(`{"b": 321}`); err != nil {
		t.Fatal(err)
	}

	if err = jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	config := viper.New()
	config.SetFs(memFS)

	err = parameter.LoadConfigFile(config, confDir, configName, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	val := config.GetInt("b")
	if val != 321 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}

func TestFetchYAMLConfig(t *testing.T) {
	filename := fmt.Sprintf("%s/%s.yml", confDir, configName)

	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err := jsonConfFile.WriteString(`c: 333`); err != nil {
		t.Fatal(err)
	}

	if err := jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	config := viper.New()
	config.SetFs(memFS)

	err = parameter.LoadConfigFile(config, confDir, configName, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	val := config.GetInt("c")
	if val != 333 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}

func TestFetchJSONConfigWithFileExtension(t *testing.T) {
	filename := fmt.Sprintf("%s/%s.json", confDir, configName)

	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err = jsonConfFile.WriteString(`{"b": 321}`); err != nil {
		t.Fatal(err)
	}

	if err = jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	config := viper.New()
	config.SetFs(memFS)

	err = parameter.LoadConfigFile(config, confDir, configName+".json", nil, false)
	if err != nil {
		t.Fatal(err)
	}

	val := config.GetInt("b")
	if val != 321 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}

func TestBindFlagSet(t *testing.T) {
	filename := fmt.Sprintf("%s/%s.json", confDir, configName)

	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err = jsonConfFile.WriteString(`{"b": 321}`); err != nil {
		t.Fatal(err)
	}

	if err = jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	config := viper.New()
	config.SetFs(memFS)

	testFlagSet := flag.NewFlagSet("", flag.ContinueOnError)
	testFlagSet.Bool("test_a", true, "this is a test")

	err = parameter.LoadConfigFile(config, confDir, configName, testFlagSet, false)
	if err != nil {
		t.Fatal(err)
	}

	val := config.GetInt("b")
	assert.Equalf(t, 321, val, "expected read config value to be %d, but was %d", 321, val)

	var exists bool
	_, exists = config.AllSettings()["test_a"]
	assert.True(t, exists, "expected read config value to exist")

	_, exists = config.AllSettings()["test_b"]
	assert.False(t, exists, "expected read config value to not exist")
}
