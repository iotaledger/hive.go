package event

import (
	"go.uber.org/atomic"
)

var (
	idCounter = atomic.NewUint64(0)
)

type Closure[T any] struct {
	ID       uint64
	Function func(event T)
}

func NewClosure[T any](function func(event T)) *Closure[T] {
	closure := &Closure[T]{
		ID:       idCounter.Inc(),
		Function: function,
	}

	return closure
}
