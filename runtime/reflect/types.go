package reflect

import (
	"reflect"
	"time"
)

var (
	// BoolType is the reflect.Type of bool.
	BoolType = reflect.TypeOf(true)
	// TimeDurationType is the reflect.Type of time.Duration.
	TimeDurationType = reflect.TypeOf(time.Duration(0))
	// Float32Type is the reflect.Type of float32.
	Float32Type = reflect.TypeOf(float32(0))
	// Float64Type is the reflect.Type of float64.
	Float64Type = reflect.TypeOf(float64(0))
	// IntType is the reflect.Type of int.
	IntType = reflect.TypeOf(int(0))
	// Int8Type is the reflect.Type of int8.
	Int8Type = reflect.TypeOf(int8(0))
	// Int16Type is the reflect.Type of int16.
	Int16Type = reflect.TypeOf(int16(0))
	// Int32Type is the reflect.Type of int32.
	Int32Type = reflect.TypeOf(int32(0))
	// Int64Type is the reflect.Type of int64.
	Int64Type = reflect.TypeOf(int64(0))
	// StringType is the reflect.Type of string.
	StringType = reflect.TypeOf("")
	// UintType is the reflect.Type of uint.
	UintType = reflect.TypeOf(uint(0))
	// Uint8Type is the reflect.Type of uint8.
	Uint8Type = reflect.TypeOf(uint8(0))
	// Uint16Type is the reflect.Type of uint16.
	Uint16Type = reflect.TypeOf(uint16(0))
	// Uint32Type is the reflect.Type of uint32.
	Uint32Type = reflect.TypeOf(uint32(0))
	// Uint64Type is the reflect.Type of uint64.
	Uint64Type = reflect.TypeOf(uint64(0))
	// StringSliceType is the reflect.Type of []string.
	StringSliceType = reflect.TypeOf([]string{})
	// StringMapType is the reflect.Type of map[string]string.
	StringMapType = reflect.TypeOf(map[string]string{})
)
