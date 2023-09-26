package serializer

// Byter is a type constraint that ensures that the type can be serialized to bytes.
type Byter interface {
	Bytes() ([]byte, error)
}

// FromByter is a type constraint that ensures that the type can be deserialized from bytes.
type FromByter interface {
	FromBytes([]byte) (int, error)
}

type MarshalablePtr[V any] interface {
	*V
	Byter
	FromByter
}
