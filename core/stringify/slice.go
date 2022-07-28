package stringify

import (
	"reflect"
	"strings"

	"github.com/kr/text"
)

func Slice(value []interface{}) string {
	return sliceReflect(reflect.ValueOf(value))
}

func sliceReflect(value reflect.Value) (result string) {
	result += "["

	newLineVersion := false
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)

		valueString := Interface(item)
		if strings.Contains(valueString, "\n") {
			if !newLineVersion {
				result += "\n"

				newLineVersion = true
			}
			result += text.Indent(Interface(item)+",\n", strings.Repeat(" ", INDENTATION_SIZE))
		} else {
			result += Interface(item) + ", "
		}
	}

	if !newLineVersion && len(result) >= 2 {
		result = result[:len(result)-2]
	}

	result += "]"

	return
}
