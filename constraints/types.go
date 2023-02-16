package constraints

// Ptr is a helper type to create a pointer type.
type Ptr[A any] interface {
	*A
}
