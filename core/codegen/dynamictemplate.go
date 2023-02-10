package codegen

type DynamicTemplate[T any] struct {
	tokenMappings map[string]func(T) string

	*Template
}

func NewDynamicTemplate[T any](tokenMappings map[string]func(T) string) *DynamicTemplate[T] {
	return &DynamicTemplate[T]{
		Template:      NewTemplate(make(map[string]string)),
		tokenMappings: tokenMappings,
	}
}

// TranscribedContent first derives the tokens that related the Content of the Template by replacing the tokens with the given argument.
func (d *DynamicTemplate[T]) TranscribedContent(arg T) (string, error) {
	for token, mapping := range d.tokenMappings {
		d.Template.TokenMappings[token] = mapping(arg)
	}

	return d.Template.TranscribedContent()
}
