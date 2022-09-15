package constraints

// Serializable is a type constraint that ensures that the type can be serialized to and from bytes.
type Serializable[V any] interface {
	*V

	FromBytes([]byte) (int, error)
	Bytes() ([]byte, error)
}
