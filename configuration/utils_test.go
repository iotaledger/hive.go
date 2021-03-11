package configuration

import (
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

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
