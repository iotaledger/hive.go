package stringify

import (
	"reflect"
	"strings"

	"github.com/kr/text"
)

func Map(value interface{}) string {
	return mapReflect(reflect.ValueOf(value))
}

func mapReflect(value reflect.Value) (result string) {
	result = "map{"

	mapKeys := value.MapKeys()
	if len(mapKeys) >= 1 {
		result += "\n"
	}

	for _, mapKey := range mapKeys {
		item := value.MapIndex(mapKey)

		result += text.Indent("["+Interface(mapKey)+"]: "+Interface(item)+",\n", strings.Repeat(" ", IndentationSize))
	}

	result += "}"

	return
}
