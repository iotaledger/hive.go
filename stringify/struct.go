package stringify

import (
	"strings"

	"github.com/kr/text"
)

// Struct creates a string representation of the given struct details.
func Struct(name string, fields ...*structField) string {
	return (&structBuilder{
		name:   name,
		fields: fields,
	}).String()
}

// StructBuilder returns a builder for the struct that can dynamically be modified.
func StructBuilder(name string, fields ...*structField) *structBuilder {
	return &structBuilder{
		name:   name,
		fields: fields,
	}
}

type structBuilder struct {
	name   string
	fields []*structField
}

// AddField dynamically adds a new field to the struct.
func (stringifyStruct *structBuilder) AddField(field *structField) {
	stringifyStruct.fields = append(stringifyStruct.fields, field)
}

func (stringifyStruct *structBuilder) String() (result string) {
	result = stringifyStruct.name + " {\n"

	for _, field := range stringifyStruct.fields {
		result += text.Indent(field.String()+"\n", strings.Repeat(" ", INDENTATION_SIZE))
	}

	result += "}"

	return result
}
