package codegen

import (
	"github.com/iotaledger/hive.go/core/generics/lo"
)

// DynamicTemplate is a Template that allows to dynamically adjust the token mappings at runtime.
type DynamicTemplate[T any] struct {
	TokenMappings map[string]func(T) string

	*Template
}

// NewDynamicTemplate creates a new DynamicTemplate with the given token mappings.
func NewDynamicTemplate[T any](tokenMappings map[string]func(T) string) *DynamicTemplate[T] {
	return &DynamicTemplate[T]{
		TokenMappings: tokenMappings,
		Template:      NewTemplate(make(map[string]string)),
	}
}

// Generate generates the file with the given name and the given number of variadic type parameters.
func (d *DynamicTemplate[T]) Generate(fileName string, arg T, optGenerator ...func(T) string) error {
	return d.Template.Generate(fileName, func() string {
		return lo.First(optGenerator, d.GenerateContent)(arg)
	})
}

// GenerateContent translates the tokens in the content to tokens relating to the given argument.
func (d *DynamicTemplate[T]) GenerateContent(arg T) string {
	for token, mapping := range d.TokenMappings {
		d.Template.TokenMappings[token] = mapping(arg)
	}

	return d.Template.GenerateContent()
}
