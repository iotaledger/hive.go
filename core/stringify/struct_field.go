package stringify

type StructField struct {
	name  string
	value interface{}
}

func NewStructField(name string, value interface{}) *StructField {
	return &StructField{
		name:  name,
		value: value,
	}
}

func (structField *StructField) String() (result string) {
	return structField.name + ": " + Interface(structField.value)
}
