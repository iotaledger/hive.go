package constraints

// Ptr is a helper type to create a pointer type.
type Ptr[A any] interface {
	*A
}

func NewFromPtr[APtr Ptr[A], A any]() APtr {
	return new(A)
}