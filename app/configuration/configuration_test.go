package configuration

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/knadh/koanf/providers/rawbytes"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	reflectutils "github.com/iotaledger/hive.go/runtime/reflect"
)

func tempFile(t *testing.T, pattern string) (string, *os.File) {
	tmpfile, err := os.CreateTemp("", pattern)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Remove(tmpfile.Name())
		require.NoError(t, err)
	})

	return tmpfile.Name(), tmpfile
}

func TestFetchGlobalFlags(t *testing.T) {
	flag.String("A", "321", "test")
	require.NoError(t, flag.Set("A", "321"))

	config := New()

	err := config.LoadFlagSet(flag.CommandLine)
	require.NoError(t, err)

	val := config.String("A")
	require.EqualValues(t, "321", val)
}

func TestFetchFlagset(t *testing.T) {
	testFlagSet := flag.NewFlagSet("", flag.ContinueOnError)
	testFlagSet.String("A", "321", "test")
	require.NoError(t, testFlagSet.Set("A", "321"))

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
	require.NoError(t, testFlagSet.Set("B", "322"))

	t.Setenv("TEST_B", "321")

	t.Setenv("TEST_C", "321")

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

	t.Setenv("TEST_F", "322")

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
	require.NoError(t, config1.Set("test.integer", 321))
	require.NoError(t, config1.Set("test.slice", []string{"string1", "string2", "string3"}))
	require.NoError(t, config1.Set("test.bool.ignore", true))

	jsonConfFileName, _ := tempFile(t, "config*.json")
	err := config1.StoreFile(jsonConfFileName, 0o600, []string{"test.bool.ignore"})
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

type Otto struct {
	Name string
}

type Parameters struct {
	TestField    int64             `shorthand:"t" usage:"you can do stuff with this parameter"`
	TestField1   bool              `name:"bernd" default:"true" usage:"batman was here"`
	TestFieldMap map[string]string `usage:"the list of allowed ottos with their favorite drink"`
	Nested       struct {
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
	Ottos  []Otto        `noflag:"true"`
}

func TestBindAndUpdateParameters(t *testing.T) {
	parameters := Parameters{
		// assign default value outside of tag
		TestField: 13,

		TestFieldMap: map[string]string{
			"Hans":  "Rum",
			"Jonas": "Water",
			"Max":   "Beer",
		},

		// assign default value inside of tag (is overridden by default value of tag)
		Batman: []string{"a", "b", "c"},

		Ottos: []Otto{
			{Name: "Bruce"}, // Batman
			{Name: "Clark"}, // Superman
			{Name: "Barry"}, // The Flash
		},
	}

	config := New()
	flagset := NewUnsortedFlagSet("", flag.ContinueOnError)
	config.BindParameters(flagset, "configuration", &parameters)

	err := config.LoadFlagSet(flagset)
	assert.NoError(t, err)

	// read in ENV variables
	// load the env vars after default values from flags were set (otherwise the env vars are not added because the keys don't exist)
	err = config.LoadEnvironmentVars("test")
	assert.NoError(t, err)

	// load the flags again to overwrite env vars that were also set via command line
	err = config.LoadFlagSet(flagset)
	assert.NoError(t, err)

	config.UpdateBoundParameters()

	assertFlag(t, flagset, config, &parameters.TestField,
		"configuration.testField",
		"you can do stuff with this parameter",
		"13",
		"t",
		int64(13),
	)

	assertFlag(t, flagset, config, &parameters.TestField1,
		"configuration.bernd",
		"batman was here",
		"true",
		"",
		true,
	)

	assertFlag(t, flagset, config, &parameters.TestFieldMap,
		"configuration.testFieldMap",
		"the list of allowed ottos with their favorite drink",
		"",
		"",
		map[string]string{
			"Hans":  "Rum",
			"Jonas": "Water",
			"Max":   "Beer",
		},
	)

	assertFlag(t, flagset, config, &parameters.Nested.Key,
		"configuration.nested.key",
		"nestedKey elephant",
		"elephant",
		"",
		"elephant",
	)

	assertFlag(t, flagset, config, &parameters.Nested.SubNested.Key,
		"configuration.nested.subNested.key",
		"nestedKey duck",
		"duck",
		"",
		"duck",
	)

	assertFlag(t, flagset, config, &parameters.Nested1.Key,
		"configuration.renamedNested.bird",
		"nestedKey bird",
		"bird",
		"b",
		"bird",
	)

	assertFlag(t, flagset, config, &parameters.Batman,
		"configuration.batman",
		"robin",
		"[a,b,c]",
		"",
		[]string{"a", "b", "c"},
	)

	dur, err := time.ParseDuration("60s")
	assert.NoError(t, err)
	assertFlag(t, flagset, config, &parameters.ItsAMe,
		"configuration.itsAMe",
		"mario",
		dur.String(),
		"",
		60*time.Second,
	)

	ottosFlag := flagset.Lookup("configuration.ottos")
	assert.Nil(t, ottosFlag)
	expectedOttos := []Otto{
		{Name: "Bruce"}, // Batman
		{Name: "Clark"}, // Superman
		{Name: "Barry"}, // The Flash
	}
	assert.Equal(t, "configuration.ottos", config.GetParameterPath(&parameters.Ottos))
	require.EqualValues(t, expectedOttos, config.Get("configuration.ottos"))
	assert.EqualValues(t, expectedOttos, parameters.Ottos)

	// check loading of config file and update of bound parameters
	exampleJSON := `{
		"configuration": {
			"testField": 14,
			"batman": ["d", "e", "f"],
			"ottos": [
				{
					"name": "Hans"
				},
				{
					"name": "Jonas"
				}
			]
		}
	}`

	err = config.Load(rawbytes.Provider([]byte(exampleJSON)), &JSONLowerParser{})
	assert.NoError(t, err)

	config.UpdateBoundParameters()

	assertConfigValue(t, config, &parameters.TestField,
		"configuration.testField",
		int64(14),
	)

	assertConfigValue(t, config, &parameters.Batman,
		"configuration.batman",
		[]string{"d", "e", "f"},
	)

	ottosFlag = flagset.Lookup("configuration.ottos")
	assert.Nil(t, ottosFlag)

	expectedOttos = []Otto{
		{Name: "Hans"},  // three lines of code
		{Name: "Jonas"}, // ripped hans whisperer
	}

	assert.Equal(t, "configuration.ottos", config.GetParameterPath(&parameters.Ottos))

	newVal := reflect.New(reflect.TypeOf(&parameters.Ottos)).Elem().Interface()
	if err := config.Unmarshal("configuration.ottos", &newVal); err != nil {
		require.NoError(t, err)
	}

	assert.EqualValues(t, &expectedOttos, newVal)
	assert.EqualValues(t, expectedOttos, parameters.Ottos)
}

func assertFlag(t *testing.T, flagset *flag.FlagSet, config *Configuration, parametersField any, name, usage, defValue, shorthand string, expectedValue any) {
	f := flagset.Lookup(name)
	assert.Equal(t, usage, f.Usage)
	if reflect.TypeOf(parametersField).Elem() != reflectutils.StringMapType {
		assert.Equal(t, defValue, f.DefValue)
	}
	assert.Equal(t, shorthand, f.Shorthand)
	assert.Equal(t, name, f.Name)
	assert.Equal(t, name, config.GetParameterPath(parametersField))
	if reflect.TypeOf(parametersField).Elem() == reflectutils.TimeDurationType {
		assert.Equal(t, expectedValue, config.Duration(name))
	} else {
		assert.Equal(t, expectedValue, config.Get(name))
	}
	assert.Equal(t, expectedValue, reflect.ValueOf(parametersField).Elem().Interface())
}

func assertConfigValue(t *testing.T, config *Configuration, parametersField any, name string, expectedValue any) {
	assert.Equal(t, name, config.GetParameterPath(parametersField))

	switch reflect.TypeOf(parametersField).Elem() {
	case reflectutils.BoolType:
		assert.Equal(t, expectedValue, config.Bool(name))
	case reflectutils.TimeDurationType:
		assert.Equal(t, expectedValue, config.Duration(name))
	case reflectutils.Float32Type:
		assert.Equal(t, expectedValue, float32(config.Float64(name)))
	case reflectutils.Float64Type:
		assert.Equal(t, expectedValue, config.Float64(name))
	case reflectutils.IntType:
		assert.Equal(t, expectedValue, config.Int(name))
	case reflectutils.Int8Type:
		assert.Equal(t, expectedValue, int8(config.Int(name)))
	case reflectutils.Int16Type:
		assert.Equal(t, expectedValue, int16(config.Int(name)))
	case reflectutils.Int32Type:
		assert.Equal(t, expectedValue, int32(config.Int(name)))
	case reflectutils.Int64Type:
		assert.Equal(t, expectedValue, config.Int64(name))
	case reflectutils.StringType:
		assert.Equal(t, expectedValue, config.String(name))
	case reflectutils.UintType:
		assert.Equal(t, expectedValue, uint(config.Int(name)))
	case reflectutils.Uint8Type:
		assert.Equal(t, expectedValue, uint8(config.Int(name)))
	case reflectutils.Uint16Type:
		assert.Equal(t, expectedValue, uint16(config.Int(name)))
	case reflectutils.Uint32Type:
		assert.Equal(t, expectedValue, uint32(config.Int(name)))
	case reflectutils.Uint64Type:
		assert.Equal(t, expectedValue, uint64(config.Int64(name)))
	case reflectutils.StringSliceType:
		assert.Equal(t, expectedValue, config.Strings(name))
	case reflectutils.StringMapType:
		assert.Equal(t, expectedValue, config.StringMap(name))
	default:
		// if we don't know the type, we try to unmarshal it
		newVal := reflect.New(reflect.TypeOf(parametersField)).Elem().Interface()
		if err := config.Unmarshal(name, &newVal); err != nil {
			require.NoError(t, err)
		}

		assert.EqualValues(t, expectedValue, newVal)
	}

	assert.Equal(t, expectedValue, reflect.Indirect(reflect.ValueOf(parametersField).Elem()).Interface())
}
