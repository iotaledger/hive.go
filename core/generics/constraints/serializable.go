package constraints

// Serializable is a type constraint that ensures that the type can be serialized to bytes.
type Serializable interface {
	Bytes() ([]byte, error)
}

// Deserializable is a type constraint that ensures that the type can be deserialized from bytes.
type Deserializable interface {
	FromBytes([]byte) (int, error)
}

// Marshalable is a type constraint that ensures that the type can be serialized and deserialized to and from bytes.
type Marshalable interface {
	Serializable
	Deserializable
}

// MarshalablePtr is a pointer type constraint to a Marshable.
type MarshalablePtr[V any] interface {
	*V
	Marshalable
}
