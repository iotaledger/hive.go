package dataflow

import (
	"github.com/iotaledger/hive.go/generics/stack"
)

type ChainedCommand[T any] func(param T, next func(param T) error) error

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

type callstack[T any] struct {
	chainedCommands stack.Stack[ChainedCommand[T]]
}

func newCallstack[T any](steps ...ChainedCommand[T]) *callstack[T] {
	chainedCommands := stack.New[ChainedCommand[T]](false)

	stepCount := len(steps)
	for i := 0; i < stepCount; i++ {
		chainedCommands.Push(steps[stepCount-i-1])
	}

	return &callstack[T]{
		chainedCommands: chainedCommands,
	}
}

func (n *callstack[T]) call(param T) (err error) {
	if n.chainedCommands.IsEmpty() {
		return nil
	}

	step, exists := n.chainedCommands.Pop()
	if !exists {
		return nil
	}

	return step(param, n.call)
}
