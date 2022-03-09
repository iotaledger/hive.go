package events

import (
	"go.uber.org/atomic"
)

var (
	idCounter = atomic.NewUint64(0)
)

type Closure struct {
	ID  uint64
	Fnc interface{}
}

func NewClosure(f interface{}) *Closure {
	closure := &Closure{
		ID:  idCounter.Inc(),
		Fnc: f,
	}

	return closure
}
