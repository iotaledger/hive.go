package stack

import (
	"sync"
)

// threadSafeStack implements a thread safe Stack.
type threadSafeStack[T any] struct {
	stack *simpleStack[T]
	mutex sync.RWMutex
}

// newThreadSafeStack returns a new thread safe Stack.
func newThreadSafeStack[T any]() *threadSafeStack[T] {
	return &threadSafeStack[T]{
		stack: newSimpleStack[T](),
	}
}

// Push pushes an element onto the top of this Stack.
func (s *threadSafeStack[T]) Push(element T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stack.Push(element)
}

// Pop removes and returns the top element of this Stack.
func (s *threadSafeStack[T]) Pop() (value T, exists bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stack.Pop()
}

// Peek returns the top element of this Stack without removing it.
func (s *threadSafeStack[T]) Peek() (value T, exists bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.stack.Peek()
}

// Clear removes all elements from this Stack.
func (s *threadSafeStack[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stack.Clear()
}

// Size returns the amount of elements in this Stack.
func (s *threadSafeStack[T]) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.stack.Size()
}

// IsEmpty checks if this Stack is empty.
func (s *threadSafeStack[T]) IsEmpty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.stack.IsEmpty()
}

// code contract - make sure the type implements the interface.
var _ Stack[int] = &threadSafeStack[int]{}
