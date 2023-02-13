package variadic

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/cockroachdb/errors"
	"github.com/iotaledger/hive.go/core/generics/lo"

	"github.com/iotaledger/hive.go/codegen"
)

// Variadic is a template that translates code into a set of variadic implementations with generic type parameters.
//
// It supports the following pipelines:
//
// +---------------+----------+------------------------------------------------------------+---------------------------+
// | Pipeline      | Type     | Description                                                | Example Output            |
// +---------------+----------+------------------------------------------------------------+---------------------------+
// | hasParams     | bool     | true if the currently generated variadic has parameters.   | true   				   |
// | paramCount    | int      | the current number of variadic parameters.                 | 2                         |
// | params        | string   | the current list of variadic parameters (without types).   | i.e. "arg1, arg2"         |
// | typedParams   | string   | the current list of variadic parameters (including types). | i.e. "arg1 T1, arg2 T2"   |
// | types         | string   | the current list of variadic types.                        | i.e. "T1, T2"             |
// +---------------+----------+------------------------------------------------------------+---------------------------+
type Variadic struct {
	// currentParamCount is the number of parameters of the currently generated variadic instance.
	currentParamCount int

	// maxParamCount is the maximum number of parameters of the generated variadic instances.
	maxParamCount int

	// Template embeds the default logic of the template framework.
	*codegen.Template
}

// New creates a new Variadic template.
func New() *Variadic {
	v := new(Variadic)
	v.Template = codegen.NewTemplate(template.FuncMap{
		"hasParams":   v.hasParams,
		"paramCount":  v.paramCount,
		"params":      v.params,
		"typedParams": v.typedParams,
		"types":       v.types,
	})

	return v
}

// Generate generates a file containing the desired number of variadic instances (it can receive an optional generator
// function that overrides the way the Content is generated).
func (v *Variadic) Generate(fileName string, minParamCount, maxParamCount int, optGenerator ...func() (string, error)) error {
	v.currentParamCount = minParamCount
	v.maxParamCount = maxParamCount

	return v.Template.Generate(fileName, lo.First(optGenerator, v.GenerateContent))
}

// GenerateContent generates multiple variadic instances of the template according to the current configuration.
func (v *Variadic) GenerateContent() (string, error) {
	var variadicInstances []string
	for ; v.currentParamCount <= v.maxParamCount; v.currentParamCount++ {
		generatedContent, err := v.Template.GenerateContent()
		if err != nil {
			return "", errors.Errorf("failed to generate variadic %d: %w", v.currentParamCount, err)
		}

		variadicInstances = append(variadicInstances, generatedContent)
	}

	return strings.Join(variadicInstances, "\n\n"), nil
}

// hasParams is a pipeline for the template that returns true if the currently generated variadic has parameters.
func (v *Variadic) hasParams() bool {
	return v.currentParamCount != 0
}

// paramCount is a pipeline for the template that returns the current number of variadic parameters.
func (v *Variadic) paramCount() int {
	return v.currentParamCount
}

// params is a pipeline for the template that returns the current list of variadic parameters (without types).
func (v *Variadic) params() string {
	return variadicString("arg%d", v.currentParamCount)
}

// typedParams is a pipeline for the template that returns the current list of variadic parameters (including types).
func (v *Variadic) typedParams() string {
	return variadicString("arg%d T%d", v.currentParamCount)
}

// types is a pipeline for the template that returns the current list of variadic types.
func (v *Variadic) types() string {
	return variadicString("T%d", v.currentParamCount)
}

// variadicString is a utility function that builds a variadic comma separated string from a template.
//
// Example: variadicString("func(%d)", 3) returns  "func(1), func(2), func(3)"
func variadicString(template string, count int) string {
	var results []string
	for i := 1; i <= count; i++ {
		results = append(results, strings.ReplaceAll(template, "%d", strconv.Itoa(i)))
	}

	return strings.Join(results, ", ")
}
