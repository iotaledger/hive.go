package dataflow

import (
	"github.com/iotaledger/hive.go/generics/stack"
)

// region DataFlow /////////////////////////////////////////////////////////////////////////////////////////////////////

type Dataflow[T any] struct {
	steps []ChainedCommand[T]
}

func New[T any](steps ...ChainedCommand[T]) *Dataflow[T] {
	return &Dataflow[T]{
		steps: steps,
	}
}

func (d *Dataflow[T]) Run(param T) error {
	return newCallstack[T](d.steps...).call(param)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ChainedCommand ///////////////////////////////////////////////////////////////////////////////////////////////

type ChainedCommand[T any] func(param T, next func(param T) error) error

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region callStack ////////////////////////////////////////////////////////////////////////////////////////////////////

type callstack[T any] struct {
	stack.Stack[ChainedCommand[T]]
}

func newCallstack[T any](steps ...ChainedCommand[T]) (newCallStack *callstack[T]) {
	return (&callstack[T]{stack.New[ChainedCommand[T]](false)}).addSteps(steps...)
}

func (n *callstack[T]) addSteps(steps ...ChainedCommand[T]) (self *callstack[T]) {
	stepCount := len(steps)
	for i := 0; i < stepCount; i++ {
		n.Push(steps[stepCount-i-1])
	}

	return n
}

func (n *callstack[T]) call(param T) (err error) {
	if n.IsEmpty() {
		return nil
	}

	step, exists := n.Pop()
	if !exists {
		return nil
	}

	return step(param, n.call)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
