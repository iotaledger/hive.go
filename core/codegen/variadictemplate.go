package codegen

import (
	"strconv"
	"strings"
)

// VariadicTemplate is a template that can be used to generate variadic implementations with generic type parameters.
//
// It takes a file with the following tokens as an input and creates multiple variadic instances of the code:
//
//   - paramCount: the number of parameters of the generic variadic instance - i.e. 3
//   - typeParams: list of type parameters without their constraints - i.e. "[T1, T2, T3]"
//   - TypeParams: list of type parameters with their constraints - i.e. "[T1, T2, T3 any]"
//   - params: list of arguments without their type: "arg1, arg2, arg3"
//   - Params: list of arguments with their type: "arg1 T1, arg2 T2, arg3 T3"
//   - Types: list of types: "T1, T2, T3"
//
// Tokens have to be surrounded by /* and */ as comments in the source file. Neighboring whitespaces can be removed by
// adding a "-" in the beginning or the end of the token.
type VariadicTemplate struct {
	*DynamicTemplate[int]
}

// NewVariadicTemplate creates a new VariadicTemplate template from the given file.
func NewVariadicTemplate() *VariadicTemplate {
	// builds a variadic string - i.e. variadicString("func(%d)", 3) returns  "func(1), func(2), func(3)"
	var variadicString = func(template string, count int) string {
		var results []string
		for i := 1; i <= count; i++ {
			results = append(results, strings.ReplaceAll(template, "%d", strconv.Itoa(i)))
		}

		return strings.Join(results, ", ")
	}

	return &VariadicTemplate{
		DynamicTemplate: NewDynamicTemplate(map[string]func(int) string{
			"paramCount": func(i int) string { return strconv.Itoa(i) },
			"typeParams": func(i int) string { return "[" + variadicString("T%d", i) + "]" },
			"TypeParams": func(i int) string { return "[" + variadicString("T%d", i) + " any]" },
			"params":     func(i int) string { return variadicString("arg%d", i) },
			"Params":     func(i int) string { return variadicString("arg%d T%d", i) },
			"Types":      func(i int) string { return variadicString("T%d", i) },
		}),
	}
}

// Generate generates the file with the given name and the given number of variadic type parameters.
func (v *VariadicTemplate) Generate(fileName string, typeParameterCount int) error {
	return v.DynamicTemplate.Transcribe(fileName, func() string {
		var variadics []string
		for i := 1; i <= typeParameterCount; i++ {
			variadics = append(variadics, v.DynamicTemplate.TranscribeContent(i))
		}

		return strings.Join(variadics, "\n\n")
	})
}
