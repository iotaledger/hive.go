package stack

import (
	"sync"
)

// threadSafeStack implements a thread safe Stack.
type threadSafeStack struct {
	sync.RWMutex
	elements []interface{}
}

// newThreadSafeStack returns a new thread safe Stack.
func newThreadSafeStack() *threadSafeStack {
	return &threadSafeStack{}
}

// Push pushes an element onto the top of this Stack.
func (s *threadSafeStack) Push(element interface{}) {
	s.Lock()
	defer s.Unlock()

	s.elements = append(s.elements, element)
}

// Pop removes and returns the top element of this Stack.
func (s *threadSafeStack) Pop() interface{} {
	s.Lock()
	defer s.Unlock()

	if len(s.elements) == 0 {
		return nil
	}

	index := len(s.elements) - 1
	element := s.elements[index]
	// erase element to avoid memory leaks for long lasting stacks
	s.elements[index] = nil
	s.elements = s.elements[:index]

	return element
}

// Peek returns the top element of this Stack without removing it.
func (s *threadSafeStack) Peek() interface{} {
	s.RLock()
	defer s.RUnlock()

	if len(s.elements) == 0 {
		return nil
	}

	return s.elements[len(s.elements)-1]
}

// Clear removes all elements from this Stack.
func (s *threadSafeStack) Clear() {
	s.Lock()
	defer s.Unlock()

	// erase elements to avoid memory leaks for long lasting stacks
	for index := range s.elements {
		s.elements[index] = nil
	}

	s.elements = s.elements[:0]
}

// Size returns the amount of elements in this Stack.
func (s *threadSafeStack) Size() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.elements)
}

// IsEmpty checks if this Stack is empty.
func (s *threadSafeStack) IsEmpty() bool {
	s.RLock()
	defer s.RUnlock()

	return len(s.elements) == 0
}

// code contract - make sure the type implements the interface
var _ Stack = &threadSafeStack{}
