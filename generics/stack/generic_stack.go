package stack

import "github.com/iotaledger/hive.go/datastructure/stack"

// genericStack implements generic wrapper for non-generic Stack.
type genericStack[T any] struct {
	stack.Stack
}

// newSimpleStack returns a new non-thread safe Stack.
func newGenericStack[T any](s stack.Stack) *genericStack[T] {
	return &genericStack[T]{
		Stack: s,
	}
}

// Push pushes an element onto the top of this Stack.
func (s *genericStack[T]) Push(element T) {
	s.Stack.Push(element)
}

// Pop removes and returns the top element of this Stack.
func (s *genericStack[T]) Pop() (value T, exists bool) {
	elem := s.Stack.Pop()
	if elem != nil {
		value = elem.(T)
		exists = true
	}
	return
}

// Peek returns the top element of this Stack without removing it.
func (s *genericStack[T]) Peek() (value T, exists bool) {
	elem := s.Stack.Peek()
	if elem != nil {
		value = elem.(T)
		exists = true
	}
	return
}

// code contract - make sure the type implements the interface
var _ Stack[int] = &genericStack[int]{}
