package codegen

// DynamicTemplate is a Template that allows to dynamically adjust the token mappings at runtime.
type DynamicTemplate[T any] struct {
	dynamicTokenMappings map[string]func(T) string

	*Template
}

// NewDynamicTemplate creates a new DynamicTemplate with the given token mappings.
func NewDynamicTemplate[T any](dynamicTokenMappings map[string]func(T) string) *DynamicTemplate[T] {
	return &DynamicTemplate[T]{
		Template:             NewTemplate(make(map[string]string)),
		dynamicTokenMappings: dynamicTokenMappings,
	}
}

// TranscribeContent first derives the tokens that related the Content of the Template by replacing the tokens with the given argument.
func (d *DynamicTemplate[T]) TranscribeContent(arg T) string {
	for token, mapping := range d.dynamicTokenMappings {
		d.Template.TokenMappings[token] = mapping(arg)
	}

	return d.Template.TranscribeContent()
}
