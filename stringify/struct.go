package stringify

import (
	"strings"

	"github.com/kr/text"
)

// Struct creates a string representation of the given struct details.
func Struct(name string, fields ...*StructField) string {
	return (&StructBuilder{
		name:   name,
		fields: fields,
	}).String()
}

// NewStructBuilder returns a builder for the struct that can dynamically be modified.
func NewStructBuilder(name string, fields ...*StructField) *StructBuilder {
	return &StructBuilder{
		name:   name,
		fields: fields,
	}
}

type StructBuilder struct {
	name   string
	fields []*StructField
}

// AddField dynamically adds a new field to the struct.
func (stringifyStruct *StructBuilder) AddField(field *StructField) {
	stringifyStruct.fields = append(stringifyStruct.fields, field)
}

func (stringifyStruct *StructBuilder) String() (result string) {
	result = stringifyStruct.name + " {\n"

	for _, field := range stringifyStruct.fields {
		result += text.Indent(field.String()+"\n", strings.Repeat(" ", IndentationSize))
	}

	result += "}"

	return result
}
