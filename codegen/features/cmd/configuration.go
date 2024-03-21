package main

import (
	"fmt"
	"reflect"
	"strings"
)

func newConfiguration(fileName, name, receiver, featuresStr, additionalFieldsStr string) any {
	// add all enabled features to a map
	features := make(map[string]bool)
	for _, feature := range strings.Split(featuresStr, ",") {
		features[feature] = true
	}

	additionalFields := make(map[string]string)
	for _, additionalField := range strings.Split(additionalFieldsStr, ",") {
		if len(additionalField) == 0 {
			continue
		}

		values := strings.Split(additionalField, "=")
		if len(values) != 2 {
			panic(fmt.Sprintf("failed to parse format of additional field, should be \"FieldName=FieldValue\", got \"%s\"", additionalField))
		}

		additionalFields[values[0]] = values[1]
	}

	// create a dynamic configuration struct
	structData := map[string]any{
		"FileName": fileName,
		"Name":     name,
		"Receiver": strings.ToLower(receiver),
		"Features": features,
	}

	// add the additional fields
	for key, value := range additionalFields {
		structData[key] = value
	}

	// gather all struct fields
	structFields := make([]reflect.StructField, 0, len(structData))
	for key, value := range structData {
		structFields = append(structFields, reflect.StructField{
			Name: key,
			Type: reflect.TypeOf(value),
		})
	}

	// create a struct type at runtime
	structType := reflect.StructOf(structFields)

	// create a new struct instance using reflection
	newStruct := reflect.New(structType).Elem()

	// set the values of the dynamic fields
	for i := range newStruct.NumField() {
		field := newStruct.Field(i)
		fieldName := structType.Field(i).Name

		if value, ok := structData[fieldName]; ok {
			field.Set(reflect.ValueOf(value))
		}
	}

	return newStruct.Interface()
}
