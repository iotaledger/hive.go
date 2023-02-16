package stack

// simpleStack implements a non-thread safe Stack.
type simpleStack[T any] []T

// newSimpleStack returns a new non-thread safe Stack.
func newSimpleStack[T any]() *simpleStack[T] {
	return new(simpleStack[T])
}

// Push pushes an element onto the top of this Stack.
func (s *simpleStack[T]) Push(element T) {
	*s = append(*s, element)
}

// Pop removes and returns the top element of this Stack.
func (s *simpleStack[T]) Pop() (value T, exists bool) {
	if s.IsEmpty() {
		return value, false
	}

	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]

	return element, true
}

// Peek returns the top element of this Stack without removing it.
func (s *simpleStack[T]) Peek() (value T, exists bool) {
	if (*s).IsEmpty() {
		return value, false
	}

	return (*s)[len(*s)-1], true
}

// Clear removes all elements from this Stack.
func (s *simpleStack[T]) Clear() {
	*s = (*s)[:0]
}

// Size returns the amount of elements in this Stack.
func (s *simpleStack[T]) Size() int {
	return len(*s)
}

// IsEmpty checks if this Stack is empty.
func (s *simpleStack[T]) IsEmpty() bool {
	return len(*s) == 0
}

// code contract - make sure the type implements the interface.
var _ Stack[int] = &simpleStack[int]{}
