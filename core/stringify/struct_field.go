package stringify

type structField struct {
	name  string
	value interface{}
}

func StructField(name string, value interface{}) *structField {
	return &structField{
		name:  name,
		value: value,
	}
}

func (structField *structField) String() (result string) {
	return structField.name + ": " + Interface(structField.value)
}
