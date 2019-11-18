package parameter_test

import (
	"fmt"
	"github.com/iotaledger/hive.go/parameter"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"testing"
)

// we use a Windows path, just to please viper, as it otherwise
// decides to append Windows drive letters to unix paths, when running
// this test under Windows.
const confDir = "C:/config"

var memFS = afero.NewMemMapFs()

func TestMain(m *testing.M) {
	if err := memFS.MkdirAll(confDir, 0755); err != nil {
		panic(err)
	}
	// swap out used file system
	viper.SetFs(memFS)
	parameter.NodeConfig.SetFs(memFS)
	os.Exit(m.Run())
}

func TestFetchJSONConfig(t *testing.T) {
	if err := flag.Set("config-dir", confDir); err != nil {
		t.Fatal(err)
	}

	filename := fmt.Sprintf("%s/config.json", confDir)
	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err := jsonConfFile.WriteString(`{"a": 321}`); err != nil {
		t.Fatal(err)
	}

	if err := jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := parameter.FetchConfig(false); err != nil {
		t.Fatal(err)
	}

	val := parameter.NodeConfig.GetInt("a")
	if val != 321 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}

func TestFetchJSONConfigFlagConfigName(t *testing.T) {
	if err := flag.Set("config-dir", confDir); err != nil {
		t.Fatal(err)
	}

	filename := fmt.Sprintf("%s/conf.json", confDir)

	if err := flag.Set("config", "conf"); err != nil {
		t.Fatal(err)
	}

	jsonConfFile, err := memFS.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer memFS.Remove(filename)

	if _, err := jsonConfFile.WriteString(`{"b": 321}`); err != nil {
		t.Fatal(err)
	}

	if err := jsonConfFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := parameter.FetchConfig(false); err != nil {
		t.Fatal(err)
	}

	val := parameter.NodeConfig.GetInt("b")
	if val != 321 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}

func TestFetchYAMLConfig(t *testing.T) {
	if err := flag.Set("config", "config"); err != nil {
		t.Fatal(err)
	}
	if err := flag.Set("config-dir", confDir); err != nil {
		t.Fatal(err)
	}

	filename := fmt.Sprintf("%s/config.yml", confDir)
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

	if err := parameter.FetchConfig(false); err != nil {
		t.Fatal(err)
	}

	val := parameter.NodeConfig.GetInt("c")
	if val != 333 {
		t.Fatalf("expected read config value to be %d, but was %d", 321, val)
	}
}
