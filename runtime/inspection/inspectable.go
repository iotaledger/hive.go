package inspection

// Inspectable is an interface that is used to represent an object that can be automatically inspected.
type Inspectable interface {
	// Inspect returns the inspected version of the object.
	Inspect(session ...Session) InspectedObject
}
