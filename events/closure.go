package events

import (
	"sync/atomic"
)

var idCounter uint64

type Closure struct {
	ID  uint64
	Fnc interface{}
}

func NewClosure(f interface{}) *Closure {
	closure := &Closure{
		ID:  atomic.AddUint64(&idCounter, 1),
		Fnc: f,
	}

	return closure
}
