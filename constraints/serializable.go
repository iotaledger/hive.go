package constraints

// Serializable is a type constraint that ensures that the type can be serialized to bytes.
type Serializable interface {
	Bytes() ([]byte, error)
}

// Deserializable is a type constraint that ensures that the type can be deserialized from bytes.
type Deserializable interface {
	FromBytes([]byte) (int, error)
}

type MarshalablePtr[V any] interface {
	*V
	Serializable
	Deserializable
}
