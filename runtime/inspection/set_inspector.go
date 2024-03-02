package inspection

// SetInspector is a utility function that can be used to create a manual inspection function for a set.
func SetInspector[T comparable](setToInspect setInterface[T], inspect func(inspectedSet InspectedObject, element T)) func(inspectedSet InspectedObject) {
	return func(inspectedSet InspectedObject) {
		_ = setToInspect.ForEach(func(element T) error {
			inspect(inspectedSet, element)

			return nil
		})
	}
}

// setInterface is an interface that is used to iterate over a setInterface of elements.
type setInterface[T comparable] interface {
	ForEach(consumer func(element T) error) error
}
