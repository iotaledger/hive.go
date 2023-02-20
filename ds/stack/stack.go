package stack

// Stack is a stack of elements.
type Stack[T any] interface {
	// Push pushes an element onto the top of this Stack.
	Push(element T)

	// Pop removes and returns the top element of this Stack and whether the element exists.
	Pop() (T, bool)

	// Peek returns the top element of this Stack without removing it.
	Peek() (T, bool)

	// Clear removes all elements from this Stack.
	Clear()

	// Size returns the amount of elements in this Stack.
	Size() int

	// IsEmpty checks if this Stack is empty.
	IsEmpty() bool
}

// New returns a new Stack that is thread safe if the optional threadSafe parameter is set to true.
func New[T any](threadSafe ...bool) Stack[T] {
	if len(threadSafe) >= 1 && threadSafe[0] {
		return newThreadSafeStack[T]()
	}

	return newSimpleStack[T]()
}
