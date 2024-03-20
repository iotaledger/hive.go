package memanalyzer

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/fjl/memsize"
)

const maxDepth = 10

// MemoryReport returns a human-readable report of the memory usage of the given struct pointer, useful to find leaks.
// Please note that this function "stops the world" when scanning referenced memory from a specific struct field,
// therefore it can produce significant hiccups when called on big, nested structures.
func MemoryReport(ptr interface{}) string {
	stringBuilder := &strings.Builder{}
	fmt.Fprint(stringBuilder, strings.Repeat("-", 80)+"\n")
	memoryReport(reflect.ValueOf(ptr).Elem(), 0, stringBuilder)
	fmt.Fprint(stringBuilder, strings.Repeat("-", 80))

	return stringBuilder.String()
}

func MemSize(ptr interface{}) uintptr {
	return memsize.Scan(ptr).Total
}

func memoryReport(v reflect.Value, indent int, stringBuilder *strings.Builder) {
	if indent/2 > maxDepth {
		return
	}

	t := v.Type()

	// dereference pointers to structs
	if t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct {
		if v.IsNil() {
			return
		}
		v = v.Elem()
		t = v.Type()
	}

	// walk down the fields
	if t.Kind() == reflect.Struct {
		for numField := range t.NumField() {
			fT := t.Field(numField)

			var fV reflect.Value
			// if the field is a struct or unexported, we have to unsafely obtain a pointer to it
			if fT.Type.Kind() != reflect.Ptr || !fT.IsExported() {
				fV = reflect.NewAt(fT.Type, unsafe.Pointer(v.Field(numField).UnsafeAddr()))
			} else {
				fV = v.Field(numField)
			}

			if fV.IsNil() {
				continue
			}

			fmt.Fprintf(stringBuilder, "%*s%s %s = %s\n", indent, "", fT.Name, fT.Type, memsize.HumanSize(memsize.Scan(fV.Interface()).Total))

			memoryReport(fV.Elem(), indent+2, stringBuilder)
		}
	}
}
