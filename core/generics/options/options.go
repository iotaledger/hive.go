package options

// Option is a generic type for an optional parameter according to the functional paradigm.
type Option[T any] func(*T)

// Apply applies the given options o the container object.
func Apply[T any](obj *T, options []Option[T], optInitFunc ...func(instance *T)) (applied *T) {
	for _, option := range options {
		option(obj)
	}

	for _, initFunc := range optInitFunc {
		initFunc(obj)
	}

	return obj
}
