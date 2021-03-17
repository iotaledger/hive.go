package configuration

import (
	"os"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBoundParameters(t *testing.T) {
	var parameters = struct {
		TestField int64 `shorthand:"t" default:"13" usage:"you can do stuff with this parameter"`
	}{}

	BindParameters(&parameters, "test")

	os.Setenv("test_test.testField", "321")

	testConfig := New()
	if err := testConfig.LoadFlagSet(flag.CommandLine); err != nil {
		panic(err)
	}

	// read in ENV variables
	// load the env vars after default values from flags were set (otherwise the env vars are not added because the keys don't exist)
	if err := testConfig.LoadEnvironmentVars("test"); err != nil {
		panic(err)
	}

	// load the flags again to overwrite env vars that were also set via command line
	if err := testConfig.LoadFlagSet(flag.CommandLine); err != nil {
		panic(err)
	}

	assert.False(t, parameters.TestField == testConfig.Int64("test.testField"))

	UpdateBoundParameters(testConfig)

	assert.True(t, parameters.TestField == testConfig.Int64("test.testField"))
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
		Batman []string `default:"a,b" usage:"robin"`
	}{
		// assign default value outside of tag
		TestField: 13,
	}

	BindParameters(&parameters)

	testFieldFlag := flag.Lookup("configuration.testField")
	assert.Equal(t, "you can do stuff with this parameter", testFieldFlag.Usage)
	assert.Equal(t, "13", testFieldFlag.DefValue)
	assert.Equal(t, "t", testFieldFlag.Shorthand)

	testField1Flag := flag.Lookup("configuration.bernd")
	assert.Equal(t, "batman was here", testField1Flag.Usage)
	assert.Equal(t, "true", testField1Flag.DefValue)
	assert.Equal(t, "", testField1Flag.Shorthand)

	elephantFlag := flag.Lookup("configuration.nested.key")
	assert.Equal(t, "nestedKey elephant", elephantFlag.Usage)
	assert.Equal(t, "elephant", elephantFlag.DefValue)
	assert.Equal(t, "", elephantFlag.Shorthand)

	duckFlag := flag.Lookup("configuration.nested.subNested.key")
	assert.Equal(t, "nestedKey duck", duckFlag.Usage)
	assert.Equal(t, "duck", duckFlag.DefValue)
	assert.Equal(t, "", duckFlag.Shorthand)

	birdFlag := flag.Lookup("configuration.renamedNested.bird")
	assert.Equal(t, "nestedKey bird", birdFlag.Usage)
	assert.Equal(t, "bird", birdFlag.DefValue)
	assert.Equal(t, "b", birdFlag.Shorthand)

	batmanFlag := flag.Lookup("configuration.batman")
	assert.Equal(t, "robin", batmanFlag.Usage)
	assert.Equal(t, []string{"a", "b"}, parameters.Batman)
	assert.Equal(t, "", batmanFlag.Shorthand)
}
