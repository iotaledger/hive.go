package stringify

import (
	"strings"

	"github.com/kr/text"
)

func Struct(name string, fields ...*structField) string {
	return structBuilder{
		name:   name,
		fields: fields,
	}.String()
}

type structBuilder struct {
	name   string
	fields []*structField
}

func (stringifyStruct structBuilder) String() (result string) {
	result = stringifyStruct.name + " {\n"

	for _, field := range stringifyStruct.fields {
		result += text.Indent(field.String()+"\n", strings.Repeat(" ", INDENTATION_SIZE))
	}

	result += "}"

	return result
}
