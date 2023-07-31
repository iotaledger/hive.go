package types

import "fmt"

type Tuple[A, B any] struct {
	A A
	B B
}

func NewTuple[A, B any](a A, b B) *Tuple[A, B] {
	return &Tuple[A, B]{
		A: a,
		B: b,
	}
}

func (t Tuple[A, B]) String() string {
	return fmt.Sprintf("(%v, %v)", t.A, t.B)
}
