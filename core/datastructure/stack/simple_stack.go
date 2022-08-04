package stack

// simpleStack implements a non-thread safe Stack.
type simpleStack []interface{}

// newSimpleStack returns a new non-thread safe Stack.
func newSimpleStack() *simpleStack {
	return &simpleStack{}
}

// Push pushes an element onto the top of this Stack.
func (s *simpleStack) Push(element interface{}) {
	*s = append(*s, element)
}

// Pop removes and returns the top element of this Stack.
func (s *simpleStack) Pop() interface{} {
	if s.IsEmpty() {
		return nil
	}

	index := len(*s) - 1
	element := (*s)[index]
	// erase element to avoid memory leaks for long lasting stacks
	(*s)[index] = nil
	*s = (*s)[:index]

	return element
}

// Peek returns the top element of this Stack without removing it.
func (s *simpleStack) Peek() interface{} {
	if (*s).IsEmpty() {
		return nil
	}

	return (*s)[len(*s)-1]
}

// Clear removes all elements from this Stack.
func (s *simpleStack) Clear() {
	// erase elements to avoid memory leaks for long lasting stacks
	for index := range *s {
		(*s)[index] = nil
	}

	*s = (*s)[:0]
}

// Size returns the amount of elements in this Stack.
func (s *simpleStack) Size() int {
	return len(*s)
}

// IsEmpty checks if this Stack is empty.
func (s *simpleStack) IsEmpty() bool {
	return len(*s) == 0
}

// code contract - make sure the type implements the interface
var _ Stack = &simpleStack{}
