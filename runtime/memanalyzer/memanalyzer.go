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
	visitedPtrs := make(map[uintptr]bool)
	fmt.Fprint(stringBuilder, strings.Repeat("-", 80)+"\n")
	memoryReport(reflect.ValueOf(ptr).Elem(), 0, stringBuilder, visitedPtrs)
	fmt.Fprint(stringBuilder, strings.Repeat("-", 80))

	return stringBuilder.String()
}

func MemSize(ptr interface{}) uintptr {
	return memsize.Scan(ptr).Total
}

func memoryReport(v reflect.Value, indent int, stringBuilder *strings.Builder, visitedPtrs map[uintptr]bool) {
	if indent/2 > maxDepth || visitedPtrs[getPointer(v)] {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if r == "bad indir" {
				fmt.Println("Recovered from 'bad indir' panic. Ignoring it.")
			} else {
				// If it's a different kind of panic, you might want to re-panic.
				panic(r)
			}
		}
	}()

	t := v.Type()

	if t.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
		t = v.Type()
	}

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
		for numField := 0; numField < t.NumField(); numField++ {
			fT := t.Field(numField)

			var fV reflect.Value
			// if the field is a struct or unexported, we have to unsafely obtain a pointer to it
			if fT.Type.Kind() != reflect.Ptr || !fT.IsExported() {
				vF := v.Field(numField)
				if !vF.CanAddr() {
					fmt.Fprintf(stringBuilder, "%*s%s %s = NOT ADDRESSABLEs\n", indent, "", fT.Name, fT.Type)
					continue
				}
				fV = reflect.NewAt(fT.Type, unsafe.Pointer(vF.UnsafeAddr()))
			} else {
				fV = v.Field(numField)
			}

			if fV.IsNil() {
				continue
			}

			if ptr := getPointer(fV); ptr != 0 {
				visitedPtrs[ptr] = true
			}

			fmt.Fprintf(stringBuilder, "%*s%s %s = %s\n", indent, "", fT.Name, fT.Type, memsize.HumanSize(memsize.Scan(fV.Interface()).Total))

			memoryReport(fV.Elem(), indent+2, stringBuilder, visitedPtrs)
		}
	}
}

func getPointer(value reflect.Value) uintptr {
	switch value.Kind() {
	case reflect.Ptr, reflect.UnsafePointer:
		return value.Pointer()

	case reflect.Struct:
		if value.CanAddr() {
			return value.Addr().Pointer()
		} else {
			return 0
		}

	case reflect.Interface:
		if value.IsNil() {
			return 0
		}
		return getPointer(value.Elem())

	default:
		// Unsupported types, you may want to extend this.
		return 0
	}
}
