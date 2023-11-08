package serializer

// Byter is a type constraint that ensures that the type can be serialized to bytes.
type Byter interface {
	Bytes() ([]byte, error)
}
